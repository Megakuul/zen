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

func (c *Controller) SendUpdate(ctx context.Context, update *Update) error {
}

func (c *Controller) ReadUpdates(ctx context.Context) ([]*Update, error) {
}
