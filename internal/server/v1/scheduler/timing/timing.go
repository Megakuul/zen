package timing

import (
	"context"

	"connectrpc.com/connect"
	"github.com/megakuul/zen/pkg/api/v1/scheduler/timing"
)

type Timing struct { }

func New() *Timing {
	return &Timing{

	}
}

func (s *Timing) Start(ctx context.Context, r *connect.Request[timing.StartRequest]) (*connect.Response[timing.StartResponse], error) {
	return nil, nil
}


func (s *Timing) Stop(ctx context.Context, r *connect.Request[timing.StopRequest]) (*connect.Response[timing.StopResponse], error) {
	return nil, nil
}
