package management

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"

	"connectrpc.com/connect"
	"github.com/dchest/captcha"
	"github.com/google/uuid"
	"github.com/megakuul/zen/internal/auth"
	"github.com/megakuul/zen/internal/model/email"
	"github.com/megakuul/zen/internal/model/user"
	"github.com/megakuul/zen/internal/token"
	"github.com/megakuul/zen/pkg/api/v1/manager"
	"github.com/megakuul/zen/pkg/api/v1/manager/management"
)

type Management struct {
	logger        *slog.Logger
	verificator   *token.Verificator
	authenticator *auth.Authenticator
	userCtrl      *user.Controller
	emailCtrl     *email.Controller
}

func New(logger *slog.Logger, verify *token.Verificator, auth *auth.Authenticator, user *user.Controller, email *email.Controller) *Management {
	return &Management{
		logger:        logger,
		verificator:   verify,
		authenticator: auth,
		userCtrl:      user,
		emailCtrl:     email,
	}
}

func (m *Management) Register(ctx context.Context, r *connect.Request[management.RegisterRequest]) (*connect.Response[management.RegisterResponse], error) {
	if r.Msg.CaptchaId == "" {
		id := captcha.New()
		image := bytes.NewBuffer(nil)
		if err := captcha.WriteImage(image, id, 128, 64); err != nil {
			m.logger.Warn(fmt.Sprintf("captcha image failure: %v", err), "endpoint", "register")
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		return &connect.Response[management.RegisterResponse]{
			Msg: &management.RegisterResponse{CaptchaId: id, CaptchaBlob: image.Bytes()},
		}, nil
	}
	if !captcha.VerifyString(r.Msg.CaptchaId, r.Msg.CaptchaDigits) {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("incorrect captcha code"))
	}
	if r.Msg.User == nil || r.Msg.Verifier == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("no valid user configuration provided"))
	} else if r.Msg.User.Email == "" || r.Msg.Verifier.Email != r.Msg.User.Email {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("verified email does not match with the user email"))
	}

	// precheck registration, even though PutRegistration() is atomic, this is required
	// to prevent users from sending unnecessary mails and to decrease chance of a userunfriendly registration failure.
	_, found, err := m.emailCtrl.GetRegistration(ctx, r.Msg.User.Email)
	if err != nil {
		return nil, err
	} else if found {
		return nil, connect.NewError(connect.CodeAlreadyExists, fmt.Errorf("email already associated with an account"))
	}

	verified, err := m.authenticator.Authenticate(ctx, r.Msg.Verifier)
	if err != nil {
		return nil, err
	} else if !verified {
		return &connect.Response[management.RegisterResponse]{
			Msg: &management.RegisterResponse{},
		}, nil
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
		m.logger.Warn(fmt.Sprintf("profile registration failure: %v", err), "endpoint", "register")
		return nil, err
	}
	err = m.emailCtrl.PutRegistration(ctx, r.Msg.Verifier.Email, &email.Registration{
		User: userId,
	})
	if err != nil {
		m.logger.Error(fmt.Sprintf("email registration failure (orphaned profile left behind): %v", err), "endpoint", "register")
		return nil, err
	}
	return &connect.Response[management.RegisterResponse]{
		Msg: &management.RegisterResponse{},
	}, nil
}

func (m *Management) Get(ctx context.Context, r *connect.Request[management.GetRequest]) (*connect.Response[management.GetResponse], error) {
	claims, err := m.verificator.Verify(ctx, strings.TrimPrefix(r.Header().Get("Authorization"), "Bearer "))
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	profile, found, err := m.userCtrl.GetProfile(ctx, claims.Subject)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	} else if !found {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("user not found"))
	}

	return &connect.Response[management.GetResponse]{
		Msg: &management.GetResponse{
			User: &manager.User{
				Id:       claims.Subject,
				Email:    claims.Email,
				Username: profile.Username,
				Score:    profile.Score,
				Streak:   profile.Streak,
			},
		},
	}, nil
}

func (m *Management) Update(ctx context.Context, r *connect.Request[management.UpdateRequest]) (*connect.Response[management.UpdateResponse], error) {
	claims, err := m.verificator.Verify(ctx, strings.TrimPrefix(r.Header().Get("Authorization"), "Bearer "))
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	err = m.userCtrl.PutProfile(ctx, claims.Subject, &user.Profile{
		Username: r.Msg.User.Username,
		Description: r.Msg.User.Description,
	})
	if err!=nil {
		m.logger.Warn(fmt.Sprintf("profile update failure: %v", err), "endpoint", "update")
		return nil, err
	}
	return &connect.Response[management.UpdateResponse]{}, nil
}

func (m *Management) Delete(ctx context.Context, r *connect.Request[management.DeleteRequest]) (*connect.Response[management.DeleteResponse], error) {
	claims, err := m.verificator.Verify(ctx, strings.TrimPrefix(r.Header().Get("Authorization"), "Bearer "))
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	verified, err := m.authenticator.Authenticate(ctx, r.Msg.Verifier)
	if err!=nil {
		return nil, err
	} else if !verified {
		return &connect.Response[management.DeleteResponse]{}, nil
	}

	// TODO transaction would be the super duper clean way here
	err = m.userCtrl.DeleteProfile(ctx, claims.Subject)
	if err != nil {
		m.logger.Warn(fmt.Sprintf("profile deletion failure: %v", err), "endpoint", "delete")
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	err = m.emailCtrl.DeleteRegistration(ctx, claims.Email)
	if err != nil {
		m.logger.Error(fmt.Sprintf("email deletion failure (orphaned email left behind): %v", err), "endpoint", "delete")
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return nil, nil
}
