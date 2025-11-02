package leaderboard

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Board struct {
	Year       string   `json:"year"`
	Week       string   `json:"week"`
	Algorithms []string `json:"algorithms"`

	Entries BoardEntry `json:"entries"`
}

type BoardEntry struct {
	UserId   string  `json:"user_id"`
	Username string  `json:"username"`
	Rating   float64 `json:"rating"`
}

func (c *Controller) GetBoard(ctx context.Context, week time.Time) (*Board, bool, error) {
}

func (c *Controller) PutBoard(ctx context.Context, week time.Time, board *Board) error {
}
