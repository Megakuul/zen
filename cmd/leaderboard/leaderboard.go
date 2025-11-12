package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/megakuul/zen/internal/httplambda"
	"github.com/megakuul/zen/internal/server/v1/manager/leaderboard"
	"github.com/megakuul/zen/internal/server/v1/manager/management"
	"github.com/megakuul/zen/pkg/api/v1/manager/leaderboard/leaderboardconnect"
	"github.com/megakuul/zen/pkg/api/v1/manager/management/managementconnect"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Config struct {
	LeaderboardBucket       string `env:"LEADERBOARD_BUCKET"`
	LeaderboardBucketPrefix string `env:"LEADERBOARD_BUCKET_PREFIX"`
}

func main() {
	config := &Config{}
	if err := cleanenv.ReadEnv(config); err != nil {
		os.Stderr.WriteString(fmt.Sprintf(
			"cannot acquire env config: %v", err,
		))
		os.Exit(1)
	}
	mux := http.NewServeMux()
	mux.Handle(
		managementconnect.NewManagementServiceHandler(management.New()),
	)
	mux.Handle(
		leaderboardconnect.NewLeaderboardServiceHandler(leaderboard.New()),
	)
	lambda.Start(createHandler(mux))
}

func createHandler(mux *http.ServeMux) func(ctx context.Context, r events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	return func(ctx context.Context, r events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
		requestor := &httplambda.LambdaRequestor{}
		request, err := requestor.Request(ctx, r)
		if err != nil {
			return events.LambdaFunctionURLResponse{
				StatusCode: http.StatusBadRequest,
				Headers:    map[string]string{"Content-Type": "text/plain"},
				Body:       err.Error(),
			}, nil
		}
		responder := &httplambda.LambdaResponder{}
		handler, _ := mux.Handler(request)
		handler.ServeHTTP(responder, request)
		return responder.Response(), nil
	}
}
