package planning

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/megakuul/zen/internal/model/user"
	"github.com/megakuul/zen/internal/token"
	"github.com/megakuul/zen/pkg/api/v1/scheduler"
	"github.com/megakuul/zen/pkg/api/v1/scheduler/planning"
)

type Service struct {
	logger    *slog.Logger
	tokenCtrl *token.Controller
	userCtrl  *user.Model
}

func New(logger *slog.Logger, token *token.Controller, user *user.Model) *Service {
	return &Service{
		logger:    logger,
		tokenCtrl: token,
		userCtrl:  user,
	}
}

func (s *Service) Get(ctx context.Context, r *connect.Request[planning.GetRequest]) (*connect.Response[planning.GetResponse], error) {
	claims, err := s.tokenCtrl.Verify(ctx, strings.TrimPrefix(r.Header().Get("Authorization"), "Bearer "))
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}
	events, err := s.userCtrl.ListEvents(ctx, claims.Subject, time.Unix(r.Msg.Since, 0), time.Unix(r.Msg.Until, 0))
	if err != nil {
		return nil, err
	}
	resp := &connect.Response[planning.GetResponse]{
		Msg: &planning.GetResponse{Events: []*scheduler.Event{}},
	}
	for _, event := range events {
		resp.Msg.Events = append(resp.Msg.Events, &scheduler.Event{
			Id:              strconv.Itoa(int(event.StartTime)),
			Type:            scheduler.EventType(event.Type),
			Name:            event.Name,
			StartTime:       event.StartTime,
			StopTime:        event.StopTime,
			TimerStartTime:  event.TimerStartTime,
			TimerStopTime:   event.TimerStopTime,
			RatingChange:    event.RatingChange,
			RatingAlgorithm: event.RatingAlgorithm,
			Immutable:       event.Immutable,
			Description:     event.Description,
			MusicUrl:        event.MusicUrl,
		})
	}
	return resp, nil
}

func (s *Service) Upsert(ctx context.Context, r *connect.Request[planning.UpsertRequest]) (*connect.Response[planning.UpsertResponse], error) {
	claims, err := s.tokenCtrl.Verify(ctx, strings.TrimPrefix(r.Header().Get("Authorization"), "Bearer "))
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}
	err = s.userCtrl.PutEvent(ctx, claims.Subject, &user.Event{
		Type:            int64(r.Msg.Event.Type),
		Name:            r.Msg.Event.Name,
		StartTime:       r.Msg.Event.StartTime,
		StopTime:        r.Msg.Event.StopTime,
		TimerStartTime:  r.Msg.Event.TimerStartTime,
		TimerStopTime:   r.Msg.Event.TimerStopTime,
		RatingChange:    r.Msg.Event.RatingChange,
		RatingAlgorithm: r.Msg.Event.RatingAlgorithm,
		Immutable:       r.Msg.Event.Immutable,
		Description:     r.Msg.Event.Description,
		MusicUrl:        r.Msg.Event.MusicUrl,
	})
	if err != nil {
		return nil, err
	}
	// if the event is not new and has a mismatch between id (events are indexed by start time) and start_time
	// this means the item got moved; as primary keys are immutable, we just upsert the new event and delete the old one.
	if r.Msg.Event.Id != "" && r.Msg.Event.Id != strconv.Itoa(int(r.Msg.Event.StartTime)) {
		s.userCtrl.DeleteEvent(ctx, claims.Subject, r.Msg.Event.Id)
	}
	return &connect.Response[planning.UpsertResponse]{
		Msg: &planning.UpsertResponse{},
	}, nil
}

func (s *Service) Delete(ctx context.Context, r *connect.Request[planning.DeleteRequest]) (*connect.Response[planning.DeleteResponse], error) {
	claims, err := s.tokenCtrl.Verify(ctx, strings.TrimPrefix(r.Header().Get("Authorization"), "Bearer "))
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}
	err = s.userCtrl.DeleteEvent(ctx, claims.Subject, r.Msg.Id)
	if err != nil {
		return nil, err
	}
	return &connect.Response[planning.DeleteResponse]{
		Msg: &planning.DeleteResponse{},
	}, nil
}
