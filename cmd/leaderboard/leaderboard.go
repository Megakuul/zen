package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	leaderboardmodel "github.com/megakuul/zen/internal/model/leaderboard"
	"github.com/megakuul/zen/internal/model/rating"
	"github.com/megakuul/zen/internal/server/v1/leaderboard"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type Config struct {
	LeaderboardQueue        string `env:"LEADERBOARD_QUEUE"`
	LeaderboardBucket       string `env:"LEADERBOARD_BUCKET"`
	LeaderboardBucketPrefix string `env:"LEADERBOARD_BUCKET_PREFIX"`
}

func main() {
	config := &Config{}
	if err := cleanenv.ReadEnv(config); err != nil {
		fmt.Fprintf(os.Stderr, "cannot acquire env config: %v", err)
		os.Exit(1)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))

	awsCfg := aws.NewConfig()
	s3Client := s3.NewFromConfig(*awsCfg)
	sqsClient := sqs.NewFromConfig(*awsCfg)

	boardModel := leaderboardmodel.New(s3Client, config.LeaderboardBucket, config.LeaderboardBucketPrefix)
	ratingModel := rating.New(sqsClient, config.LeaderboardQueue)
	service := leaderboard.New(logger, boardModel, ratingModel)

	lambda.Start(service.Process)
}
