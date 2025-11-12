package leaderboard

import (
	"context"
	"time"
)

type Update struct {
	UserId       string  `json:"user_id"`
	Username     string  `json:"username"`
	Algorithm    string  `json:"algorithm"`
	RatingChange float64 `json:"rating_change"`
}

func (m *Model) SendUpdate(ctx context.Context, update *Update) error {
}

func (m *Model) ReadUpdates(ctx context.Context) ([]*Update, error) {
}
