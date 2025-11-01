package authentication

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamotypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"

	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/megakuul/zen/pkg/api/v1/manager/authentication"
)

type Authentication struct {
	dynamoClient *dynamodb.Client
	dynamoTable  string
	sesClient    *ses.Client
	emailSender  string
}

func New() *Authentication {
	return &Authentication{}
}

func (s *Authentication) Get(ctx context.Context, r *connect.Request[authentication.GetRequest]) (*connect.Response[authentication.GetResponse], error) {
	return nil, nil
}

func (s *Authentication) Revoke(ctx context.Context, r *connect.Request[authentication.RevokeRequest]) (*connect.Response[authentication.RevokeResponse], error) {
	return nil, nil
}
