// package auth wraps the email code authentication logic.
package auth

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
	"github.com/megakuul/zen/pkg/api/v1/manager"
)

type Authenticator struct {
	emailCtrl *email.Controller

	sesClient *ses.Client
	sender    string
}

func New(emailCtrl *email.Controller, sesClient *ses.Client, sender string) *Authenticator {
	return &Authenticator{
		emailCtrl: emailCtrl,
		sesClient: sesClient,
		sender:    sender,
	}
}

// Authenticate performs the verification process, it takes a verifier and performs actions based on the verifier stage.
// Returns a bool that specifies whether the email is verified or not (earlier stages).
func (a *Authenticator) Authenticate(ctx context.Context, verifier *manager.Verifier) (bool, error) {
	switch verifier.Stage {
	case manager.VerifierStage_VERIFIER_STAGE_EMAIL:
		return false, a.processEmailStage(ctx, verifier.Email)
	case manager.VerifierStage_VERIFIER_STAGE_CODE:
		err := a.processCodeStage(ctx, verifier.Email, verifier.Code)
		if err != nil {
			return false, err
		}
		return true, nil
	default:
		return false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid stage; expected 'email' or 'code'"))
	}
}

// processEmailStage is executed if the user provides the email but not the verification code.
// -> generates a code, stores it in the database and sends it via email.
func (a *Authenticator) processEmailStage(ctx context.Context, emailAddr string) error {
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

	err := a.emailCtrl.PutCode(ctx, emailAddr, &email.Code{
		Code:      code.String(),
		ExpiresAt: time.Now().Unix(),
	})
	if err != nil {
		return connect.NewError(connect.CodeAlreadyExists, fmt.Errorf("email already sent"))
	}

	_, err = a.sesClient.SendEmail(ctx, &ses.SendEmailInput{
		Destination: &sestypes.Destination{
			ToAddresses: []string{emailAddr},
		},
		Source: aws.String(a.sender),
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
func (a *Authenticator) processCodeStage(ctx context.Context, emailAddr, submittedCode string) error {
	code, found, err := a.emailCtrl.GetCode(ctx, emailAddr)
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
