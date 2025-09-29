package leaderboard

import (
	"context"

	"connectrpc.com/connect"
	"github.com/megakuul/zen/pkg/api/v1/manager/leaderboard"
)

type Leaderboard struct { }

func New() *Leaderboard {
	return &Leaderboard{

	}
}

func (s *Leaderboard) Get(ctx context.Context, r *connect.Request[leaderboard.GetRequest]) (*connect.Response[leaderboard.GetResponse], error) {
	return nil, nil
}
