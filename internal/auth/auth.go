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

type Controller struct {
	emailCtrl *email.Model

	sesClient *ses.Client
	sender    string
}

func New(emailCtrl *email.Model, sesClient *ses.Client, sender string) *Controller {
	return &Controller{
		emailCtrl: emailCtrl,
		sesClient: sesClient,
		sender:    sender,
	}
}

// Authenticate performs the verification process, it takes a verifier and performs actions based on the verifier stage.
// Returns a bool that specifies whether the email is verified or not (earlier stages).
func (c *Controller) Authenticate(ctx context.Context, verifier *manager.Verifier) (bool, error) {
	switch verifier.Stage {
	case manager.VerifierStage_VERIFIER_STAGE_EMAIL:
		return false, c.processEmailStage(ctx, verifier.Email)
	case manager.VerifierStage_VERIFIER_STAGE_CODE:
		err := c.processCodeStage(ctx, verifier.Email, verifier.Code)
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
func (c *Controller) processEmailStage(ctx context.Context, emailAddr string) error {
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

	err := c.emailCtrl.PutCode(ctx, emailAddr, &email.Code{
		Code:      code.String(),
		ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
	})
	if err != nil {
		return err
	}

	_, err = c.sesClient.SendEmail(ctx, &ses.SendEmailInput{
		Destination: &sestypes.Destination{
			ToAddresses: []string{emailAddr},
		},
		Source: aws.String(c.sender),
		Message: &sestypes.Message{
			Subject: &sestypes.Content{
				Data: aws.String("Verification Code"), Charset: aws.String("UTF-8"),
			},
			Body: &sestypes.Body{Text: &sestypes.Content{
				Data:    aws.String(code.String()),
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
func (c *Controller) processCodeStage(ctx context.Context, emailAddr, submittedCode string) error {
	code, found, err := c.emailCtrl.GetCode(ctx, emailAddr)
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
