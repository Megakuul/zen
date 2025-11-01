package management

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"
	"github.com/dchest/captcha"
	"github.com/google/uuid"
	"github.com/megakuul/zen/internal/model/email"
	"github.com/megakuul/zen/internal/model/user"
	"github.com/megakuul/zen/internal/verify"
	"github.com/megakuul/zen/pkg/api/v1/manager/management"
)

type Management struct {
	logger      *slog.Logger
	verificator *verify.Verificator
	userCtrl    *user.Controller
	emailCtrl   *email.Controller
}

func New(logger *slog.Logger, verificator *verify.Verificator, user *user.Controller, email *email.Controller) *Management {
	return &Management{
		logger:      logger,
		verificator: verificator,
		userCtrl:    user,
		emailCtrl:   email,
	}
}

func (m *Management) Register(ctx context.Context, r *connect.Request[management.RegisterRequest]) (*connect.Response[management.RegisterResponse], error) {
	if r.Msg.CaptchaId == "" {
		id := captcha.New()
		image := bytes.NewBuffer(nil)
		if err := captcha.WriteImage(image, id, 128, 64); err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		return &connect.Response[management.RegisterResponse]{
			Msg: &management.RegisterResponse{CaptchaId: id, CaptchaBlob: image.Bytes()},
		}, nil
	}
	if !captcha.VerifyString(r.Msg.CaptchaId, r.Msg.CaptchaDigits) {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("incorrect captcha code"))
	}
	if r.Msg.User == nil || r.Msg.User.Email == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("no valid user configuration provided"))
	}

	// precheck registration, even though PutRegistration() is atomic, this is required
	// to prevent users from sending unnecessary mails and to decrease chance of a userunfriendly registration failure.
	_, found, err := m.emailCtrl.GetRegistration(ctx, r.Msg.User.Email)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	} else if found {
		return nil, connect.NewError(connect.CodeAlreadyExists, fmt.Errorf("email already associated with an account"))
	}

	verifiedEmail, verified, err := m.verificator.Verify(ctx, r.Msg.Verifier)
	if err != nil {
		return nil, err
	} else if !verified {
		return &connect.Response[management.RegisterResponse]{
			Msg: &management.RegisterResponse{},
		}, nil
	} else if verifiedEmail != r.Msg.User.Email {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("verified email does not match with the user email"))
	}

	userId := uuid.New().String()

	// TODO: to make this super duper clean it should be a transaction, but this makes the code a bit messy.
	// -> so rn if email registration fails (extremly unlikely) I just leave an orphaned account (but who cares tbh)
	err = m.userCtrl.PutProfile(ctx, userId, &user.Profile{
		Username: r.Msg.User.Username,
		Streak:   0,
		Score:    0,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	err = m.emailCtrl.PutRegistration(ctx, r.Msg.User.Email, &email.Registration{
		User: userId,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return &connect.Response[management.RegisterResponse]{
		Msg: &management.RegisterResponse{},
	}, nil
}

func (s *Management) Get(ctx context.Context, r *connect.Request[management.GetRequest]) (*connect.Response[management.GetResponse], error) {
	return nil, nil
}

func (s *Management) Update(ctx context.Context, r *connect.Request[management.UpdateRequest]) (*connect.Response[management.UpdateResponse], error) {
	return nil, nil
}

func (s *Management) Delete(ctx context.Context, r *connect.Request[management.DeleteRequest]) (*connect.Response[management.DeleteResponse], error) {
	return nil, nil
}
