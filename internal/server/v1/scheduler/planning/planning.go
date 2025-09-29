package planning

import (
	"context"

	"connectrpc.com/connect"
	"github.com/megakuul/zen/pkg/api/v1/scheduler/planning"
)

type Planning struct { }

func New() *Planning {
	return &Planning{

	}
}

func (s *Planning) Get(ctx context.Context, r *connect.Request[planning.GetRequest]) (*connect.Response[planning.GetResponse], error) {
	return nil, nil
}


func (s *Planning) Upsert(ctx context.Context, r *connect.Request[planning.UpsertRequest]) (*connect.Response[planning.UpsertResponse], error) {
	return nil, nil
}


func (s *Planning) Delete(ctx context.Context, r *connect.Request[planning.DeleteRequest]) (*connect.Response[planning.DeleteResponse], error) {
	return nil, nil
}
