package timing

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/megakuul/zen/internal/auth"
	"github.com/megakuul/zen/internal/model/leaderboard"
	"github.com/megakuul/zen/internal/model/user"
	"github.com/megakuul/zen/internal/rating"
	"github.com/megakuul/zen/internal/token"
	"github.com/megakuul/zen/pkg/api/v1/scheduler/timing"
)

type Service struct {
	logger          *slog.Logger
	verificator     *token.Verificator
	authenticator   *auth.Authenticator
	userCtrl        *user.Controller
	leaderboardCtrl *leaderboard.Controller
	ratingAnchor    time.Duration
}

func New(logger *slog.Logger, verify *token.Verificator, auth *auth.Authenticator, user *user.Controller, leaderboard *leaderboard.Controller, ratingAnchor time.Duration) *Service {
	return &Service{
		logger:          logger,
		verificator:     verify,
		authenticator:   auth,
		userCtrl:        user,
		leaderboardCtrl: leaderboard,
		ratingAnchor:    ratingAnchor,
	}
}

func (s *Service) Start(ctx context.Context, r *connect.Request[timing.StartRequest]) (*connect.Response[timing.StartResponse], error) {
	claims, err := s.verificator.Verify(ctx, strings.TrimPrefix(r.Header().Get("Authorization"), "Bearer "))
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}
	event, found, err := s.userCtrl.GetEvent(ctx, claims.Subject, r.Msg.Id)
	if err != nil {
		return nil, err
	} else if !found {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("event does not exist"))
	} else if event.Immutable {
		// just a precheck to provide a userfriendly error (the check is also supplied as atomic operation in the update)
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("event already concluded"))
	}
	err = s.userCtrl.UpdateEventTimer(ctx, claims.Subject, r.Msg.Id, 
		time.Now(), time.Unix(event.StartTime, 0), event.RatingChange, event.RatingAlgorithm, false)
	if err != nil {
		return nil, err
	}
	return &connect.Response[timing.StartResponse]{
		Msg: &timing.StartResponse{},
	}, nil
}

func (s *Service) Stop(ctx context.Context, r *connect.Request[timing.StopRequest]) (*connect.Response[timing.StopResponse], error) {
	claims, err := s.verificator.Verify(ctx, strings.TrimPrefix(r.Header().Get("Authorization"), "Bearer "))
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	profile, found, err := s.userCtrl.GetProfile(ctx, claims.Subject)
	if err != nil {
		return nil, err
	} else if !found {
		// invalid accesstoken -> return unauthenticated to trigger re-authentication
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid access token"))
	}

	event, found, err := s.userCtrl.GetEvent(ctx, claims.Subject, r.Msg.Id)
	if err != nil {
		return nil, err
	} else if !found {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("event does not exist"))
	} else if event.Immutable {
		// just a precheck to provide a userfriendly error (the check is also supplied as atomic operation in the update)
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("event already concluded"))
	}

	timerStopTime := time.Now()
	algorithm, ratingChange := rating.CalculateRatingChange(
		time.Unix(event.StartTime, 0),
		time.Unix(event.StopTime, 0),
		time.Unix(event.TimerStartTime, 0),
		timerStopTime,
		s.ratingAnchor,
	)

	err = s.leaderboardCtrl.SendUpdate(ctx, &leaderboard.Update{
		UserId:       claims.Subject,
		Username:     profile.Username,
		Algorithm:    algorithm,
		RatingChange: ratingChange,
	})
	if err != nil {
		return nil, err
	}

	err = s.userCtrl.UpdateEventTimer(ctx, claims.Subject, r.Msg.Id,
		time.Unix(event.TimerStartTime, 0), timerStopTime, ratingChange, algorithm, true)
	if err != nil {
		return nil, err
	}
	return &connect.Response[timing.StopResponse]{
		Msg: &timing.StopResponse{},
	}, nil
}
