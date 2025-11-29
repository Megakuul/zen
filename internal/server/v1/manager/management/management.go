package management

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

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

type Service struct {
	logger     *slog.Logger
	tokenCtrl  *token.Controller
	authCtrl   *auth.Controller
	userModel  *user.Model
	emailModel *email.Model
}

func New(logger *slog.Logger, token *token.Controller, auth *auth.Controller, user *user.Model, email *email.Model) *Service {
	return &Service{
		logger:     logger,
		tokenCtrl:  token,
		authCtrl:   auth,
		userModel:  user,
		emailModel: email,
	}
}

func (s *Service) Register(ctx context.Context, r *connect.Request[management.RegisterRequest]) (*connect.Response[management.RegisterResponse], error) {
	if r.Msg.User == nil || r.Msg.Verifier == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("no valid user configuration provided"))
	} else if r.Msg.User.Email == "" || r.Msg.Verifier.Email != r.Msg.User.Email {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("verifier email does not match with the user email"))
	}

	if r.Msg.CaptchaId == "" {
		id := captcha.New()
		image := bytes.NewBuffer(nil)
		if err := captcha.WriteImage(image, id, 128, 64); err != nil {
			s.logger.Warn(fmt.Sprintf("captcha image failure: %v", err), "endpoint", "register")
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		return connect.NewResponse(&management.RegisterResponse{
			CaptchaId: id, CaptchaBlob: image.Bytes(),
		}), nil
	}
	// captcha check is only skipped when the user is in the code stage.
	if r.Msg.Verifier.Stage != manager.VerifierStage_VERIFIER_STAGE_CODE && !captcha.VerifyString(r.Msg.CaptchaId, r.Msg.CaptchaDigits) {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("incorrect captcha code"))
	}

	// precheck registration, even though PutRegistration() is atomic, this is required
	// to prevent users from sending unnecessary mails and to decrease chance of a userunfriendly registration failure.
	_, found, err := s.emailModel.GetRegistration(ctx, r.Msg.User.Email)
	if err != nil {
		return nil, err
	} else if found {
		return nil, connect.NewError(connect.CodeAlreadyExists, fmt.Errorf("email already associated with an account"))
	}

	verified, err := s.authCtrl.Authenticate(ctx, r.Msg.Verifier)
	if err != nil {
		return nil, err
	} else if !verified {
		return connect.NewResponse(&management.RegisterResponse{}), nil
	}
	userId := uuid.New().String()

	// TODO: to make this super duper clean it should be a transaction, but this makes the code a bit messy.
	// -> so rn if email registration fails (extremly unlikely) I just leave an orphaned account (but who cares tbh)
	err = s.userModel.PutProfile(ctx, userId, &user.Profile{
		Username:    r.Msg.User.Username,
		Description: r.Msg.User.Description,
		Leaderboard: r.Msg.User.Leaderboard,
		CreatedAt:   time.Now().Unix(),
		Streak:      0,
		Score:       0,
		MaxStreak:   0,
	})
	if err != nil {
		s.logger.Warn(fmt.Sprintf("profile registration failure: %v", err), "endpoint", "register")
		return nil, err
	}
	err = s.emailModel.PutRegistration(ctx, r.Msg.Verifier.Email, &email.Registration{
		User: userId,
	})
	if err != nil {
		s.logger.Error(fmt.Sprintf("email registration failure (orphaned profile left behind): %v", err), "endpoint", "register")
		return nil, err
	}
	return connect.NewResponse(&management.RegisterResponse{}), nil
}

func (s *Service) Get(ctx context.Context, r *connect.Request[management.GetRequest]) (*connect.Response[management.GetResponse], error) {
	claims, err := s.tokenCtrl.Verify(ctx, strings.TrimPrefix(r.Header().Get("Authorization"), "Bearer "))
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	profile, found, err := s.userModel.GetProfile(ctx, claims.Subject)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	} else if !found {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("user not found"))
	}

	return connect.NewResponse(&management.GetResponse{
		User: &manager.User{
			Id:          claims.Subject,
			Email:       claims.Email,
			Username:    profile.Username,
			Description: profile.Description,
			Leaderboard: profile.Leaderboard,
			CreatedAt:   profile.CreatedAt,
			Score:       profile.Score,
			Streak:      profile.Streak,
			MaxStreak:   profile.MaxStreak,
		},
	}), nil
}

func (s *Service) Update(ctx context.Context, r *connect.Request[management.UpdateRequest]) (*connect.Response[management.UpdateResponse], error) {
	claims, err := s.tokenCtrl.Verify(ctx, strings.TrimPrefix(r.Header().Get("Authorization"), "Bearer "))
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	err = s.userModel.UpdateProfile(ctx, claims.Subject, &user.Profile{
		Username:    r.Msg.User.Username,
		Description: r.Msg.User.Description,
		Leaderboard: r.Msg.User.Leaderboard,
	})
	if err != nil {
		s.logger.Warn(fmt.Sprintf("profile update failure: %v", err), "endpoint", "update")
		return nil, err
	}
	return connect.NewResponse(&management.UpdateResponse{}), nil
}

func (s *Service) Delete(ctx context.Context, r *connect.Request[management.DeleteRequest]) (*connect.Response[management.DeleteResponse], error) {
	claims, err := s.tokenCtrl.Verify(ctx, strings.TrimPrefix(r.Header().Get("Authorization"), "Bearer "))
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	verified, err := s.authCtrl.Authenticate(ctx, r.Msg.Verifier)
	if err != nil {
		return nil, err
	} else if !verified {
		return connect.NewResponse(&management.DeleteResponse{}), nil
	}

	// TODO transaction would be the super duper clean way here
	err = s.userModel.DeleteProfile(ctx, claims.Subject)
	if err != nil {
		s.logger.Warn(fmt.Sprintf("profile deletion failure: %v", err), "endpoint", "delete")
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	err = s.emailModel.DeleteRegistration(ctx, claims.Email)
	if err != nil {
		s.logger.Error(fmt.Sprintf("email deletion failure (orphaned email left behind): %v", err), "endpoint", "delete")
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&management.DeleteResponse{}), nil
}
