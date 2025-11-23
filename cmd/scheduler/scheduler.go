package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/matelang/jwt-go-aws-kms/v2/jwtkms"
	"github.com/megakuul/zen/internal/httplambda"
	"github.com/megakuul/zen/internal/model/rating"
	"github.com/megakuul/zen/internal/model/user"
	"github.com/megakuul/zen/internal/server/v1/scheduler/planning"
	"github.com/megakuul/zen/internal/server/v1/scheduler/timing"
	"github.com/megakuul/zen/internal/token"
	"github.com/megakuul/zen/pkg/api/v1/scheduler/planning/planningconnect"
	"github.com/megakuul/zen/pkg/api/v1/scheduler/timing/timingconnect"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type Config struct {
	Table            string        `env:"TABLE"`
	TokenIssuer      string        `env:"TOKEN_ISSUER"`
	TokenKmsKeyId    string        `env:"TOKEN_KMS_KEY_ID"`
	LeaderboardQueue string        `env:"LEADERBOARD_QUEUE"`
	RatingAnchor     time.Duration `env:"RATING_ANCHOR" env-default:"2m"`
}

func main() {
	cfg := &Config{}
	if err := cleanenv.ReadEnv(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "cannot acquire env config: %v", err)
		os.Exit(1)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))

	awsCfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot load aws default config: %v", err)
		os.Exit(1)
	}
	dynamoClient := dynamodb.NewFromConfig(awsCfg)
	sqsClient := sqs.NewFromConfig(awsCfg)
	kmsClient := kms.NewFromConfig(awsCfg)

	userModel := user.New(dynamoClient, cfg.Table)
	ratingModel := rating.New(sqsClient, cfg.LeaderboardQueue)
	tokenCtrl := token.New(cfg.TokenIssuer, jwtkms.NewKMSConfig(kmsClient, cfg.TokenKmsKeyId, false))

	mux := http.NewServeMux()
	mux.Handle(
		planningconnect.NewPlanningServiceHandler(planning.New(logger, tokenCtrl, userModel)),
	)
	mux.Handle(
		timingconnect.NewTimingServiceHandler(timing.New(logger, tokenCtrl, userModel, ratingModel, cfg.RatingAnchor)),
	)
	lambda.Start(createHandler(mux))
}

func createHandler(mux *http.ServeMux) func(ctx context.Context, r events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	return func(ctx context.Context, r events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
		requestor := httplambda.NewRequestor()
		request, err := requestor.Request(ctx, r)
		if err != nil {
			return events.LambdaFunctionURLResponse{
				StatusCode: http.StatusBadRequest,
				Headers:    map[string]string{"Content-Type": "text/plain"},
				Body:       err.Error(),
			}, nil
		}
		responder := httplambda.NewResponder()
		handler, _ := mux.Handler(request)
		handler.ServeHTTP(responder, request)
		return responder.Response(), nil
	}
}
