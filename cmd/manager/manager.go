package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/dchest/captcha"
	"github.com/ilyakaznacheev/cleanenv"
	captchastore "github.com/megakuul/zen/internal/captcha"
	"github.com/megakuul/zen/internal/httplambda"
	"github.com/megakuul/zen/internal/server/v1/manager/management"
	"github.com/megakuul/zen/pkg/api/v1/manager/leaderboard/leaderboardconnect"
	"github.com/megakuul/zen/pkg/api/v1/manager/management/managementconnect"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Config struct {
	Bucket string `env:"BUCKET"`
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	config := &Config{}
	if err := cleanenv.ReadEnv(config); err != nil {
		os.Stderr.WriteString(fmt.Sprintf(
			"cannot acquire env config: %v", err,
		))
		os.Exit(1)
	}

	awsConfig := aws.NewConfig()

	// no I'm not responsible for this global setCustomStore mess :<
	captcha.SetCustomStore(captchastore.New(
		s3.NewFromConfig(*awsConfig),
		logger, 2 * time.Second, 
		config.Bucket, "/captcha",
	))

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
		if err!=nil {
			return events.LambdaFunctionURLResponse{
				StatusCode: http.StatusBadRequest,
				Headers: map[string]string{"Content-Type": "text/plain"},
				Body: err.Error(),
			}, nil
		}
		responder := &httplambda.LambdaResponder{}
		handler, _ := mux.Handler(request)
		handler.ServeHTTP(responder, request)
		return responder.Response(), nil
	}
}
