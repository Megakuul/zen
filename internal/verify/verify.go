// package verify wraps the email code verification logic used for authentication.
package verify

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	sestypes "github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/megakuul/zen/internal/model/email"
)

type Verificator struct {
	emailCtrl *email.Controller

	sesClient *ses.Client
	sender    string
}

func New(emailCtrl *email.Controller, sesClient *ses.Client, sender string) *Verificator {
	return &Verificator{
		emailCtrl: emailCtrl,
		sesClient: sesClient,
		sender:    sender,
	}
}

// Verify performs the verification process, it takes a verifier string that expects a certain format based on the verification stage.
// -> stage 1 = 'email:<email>' and stage 2 = 'code:<email>:<code>'
// Returns the extracted email and a bool that states whether this email is verified.
func (v *Verificator) Verify(ctx context.Context, verifier string) (string, bool, error) {
	blocks := strings.SplitN(verifier, ":", 2)
	if len(blocks) < 2 {
		return "", false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid verifier format; expected '<stage>:<value>'"))
	}
	stage, value := blocks[0], blocks[1]
	switch stage {
	case "email":
		return value, false, v.processEmailStage(ctx, value)
	case "code":
		blocks = strings.SplitN(value, ":", 2)
		if len(blocks) < 2 {
			return "", false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid code format; expected 'code:<email>:<code>'"))
		}
		err := v.processCodeStage(ctx, blocks[0], blocks[1])
		if err!=nil {
			return "", false, err
		}
		return blocks[0], true, nil
	default:
		return "", false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid stage; expected 'email' or 'code'"))
	}
}

// processEmailStage is executed if the user provides the email but not the verification code.
// -> generates a code, stores it in the database and sends it via email.
func (v *Verificator) processEmailStage(ctx context.Context, emailAddr string) error {
	codeChars := "23456789ABCDEFGHJKLMNOPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	code := strings.Builder{}
	for i := range 8 {
		if i == 3 {
			code.WriteByte('-')
		} else {
			codeCharsIdx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(codeChars)-1)))
			code.WriteByte(codeChars[codeCharsIdx.Int64()])
		}
	}

	err := v.emailCtrl.PutCode(ctx, emailAddr, &email.Code{
		Code:      code.String(),
		ExpiresAt: time.Now().Unix(),
	})
	if err != nil {
		return connect.NewError(connect.CodeAlreadyExists, fmt.Errorf("email already sent"))
	}

	_, err = v.sesClient.SendEmail(ctx, &ses.SendEmailInput{
		Destination: &sestypes.Destination{
			ToAddresses: []string{emailAddr},
		},
		Source: aws.String(v.sender),
		Message: &sestypes.Message{
			Subject: &sestypes.Content{
				Data: aws.String("Verification Code"), Charset: aws.String("UTF-8"),
			},
			Body: &sestypes.Body{Text: &sestypes.Content{
				Data: aws.String(fmt.Sprintf(
					"Your application verification code is '%s'", code.String(),
				)),
				Charset: aws.String("UTF-8"),
			}},
		},
	})
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	return nil
}

// processCodeStage is executed if the user provides the email and the verification.
// -> reads the code from database and compares it.
func (v *Verificator) processCodeStage(ctx context.Context, emailAddr, submittedCode string) error {
	code, found, err := v.emailCtrl.GetCode(ctx, emailAddr)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	} else if !found {
		return connect.NewError(connect.CodeNotFound, fmt.Errorf("code does not exist or has expired"))
	}
	if code.Code != "" && code.Code != submittedCode {
		return connect.NewError(connect.CodePermissionDenied, fmt.Errorf("incorrect code - permission denied"))
	}
	return nil
}
