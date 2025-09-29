package management

import (
	"context"

	"connectrpc.com/connect"
	"github.com/megakuul/zen/pkg/api/v1/manager/management"
)

type Management struct { }

func New() *Management {
	return &Management{

	}
}

func (s *Management) Get(ctx context.Context, r *connect.Request[management.GetRequest]) (*connect.Response[management.GetResponse], error) {
	return nil, nil
}


func (s *Management) Delete(ctx context.Context, r *connect.Request[management.DeleteRequest]) (*connect.Response[management.DeleteResponse], error) {
	return nil, nil
}
