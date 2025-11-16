package leaderboard

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/megakuul/zen/internal/model/leaderboard"
	"github.com/megakuul/zen/internal/model/rating"
)

type Service struct {
	logger      *slog.Logger
	boardModel  *leaderboard.Model
	ratingModel *rating.Model
}

func New(logger *slog.Logger, board *leaderboard.Model, rating *rating.Model) *Service {
	return &Service{
		logger:      logger,
		boardModel:  board,
		ratingModel: rating,
	}
}

func (s *Service) Process(ctx context.Context, r events.SQSEvent) error {
	s.logger.Debug(fmt.Sprintf("processing %d events from sqs...", len(r.Records)))
	board, found, err := s.boardModel.GetBoard(ctx, time.Now())
	if err != nil {
		s.logger.Error(fmt.Sprintf("critical error: failed to read board: %v", err))
		return fmt.Errorf("cannot lookup leaderboard: %v", err)
	} else if !found {
		year, week := time.Now().ISOWeek()
		s.logger.Info(fmt.Sprintf("creating new weekly board for %d-%d", year, week))
		board = &leaderboard.Board{
			Year:       strconv.Itoa(year),
			Week:       strconv.Itoa(week),
			Algorithms: map[string]int64{},
			Entries:    map[string]leaderboard.BoardEntry{},
		}
	}
	for _, message := range r.Records {
		update, err := s.ratingModel.ParseUpdate(message.Body)
		if err != nil {
			s.logger.Error(fmt.Sprintf("critical error: failed to read message '%s': %v", message.MessageId, err))
			return fmt.Errorf("failed to parse update: %v", err)
		}
		board.Algorithms[update.Algorithm] = time.Now().Unix()
		entry, ok := board.Entries[update.UserId]
		if !ok {
			entry = leaderboard.BoardEntry{
				UserId:   update.UserId,
				Username: update.Username,
				Streak:   update.Streak,
				Rating: map[int64]float64{
					update.Time.Unix(): update.RatingChange,
				},
			}
		} else {
			entry.Streak = update.Streak
			// by using a rating map here instead of +=, an update is idempotent.
			// (important for cases where the lambda fails and the message is reprocessed).
			entry.Rating[update.Time.Unix()] = update.RatingChange
		}
		board.Entries[update.UserId] = entry
	}
	err = s.boardModel.PutBoard(ctx, time.Now(), board)
	if err != nil {
		s.logger.Error(fmt.Sprintf("failure while inserting updated board: %v", err))
		return fmt.Errorf("failed to insert updated board: %v", err)
	}
	return nil
}
