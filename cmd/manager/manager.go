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
	"github.com/matelang/jwt-go-aws-kms/v2/jwtkms"
	"github.com/megakuul/zen/internal/auth"
	captchastore "github.com/megakuul/zen/internal/captcha"
	"github.com/megakuul/zen/internal/httplambda"
	"github.com/megakuul/zen/internal/model/email"
	"github.com/megakuul/zen/internal/model/user"
	"github.com/megakuul/zen/internal/server/v1/manager/authentication"
	"github.com/megakuul/zen/internal/server/v1/manager/management"
	"github.com/megakuul/zen/internal/token"
	"github.com/megakuul/zen/pkg/api/v1/manager/authentication/authenticationconnect"
	"github.com/megakuul/zen/pkg/api/v1/manager/management/managementconnect"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ses"
)

type Config struct {
	Table               string `env:"TABLE"`
	TokenIssuer         string `env:"TOKEN_ISSUER"`
	TokenKmsKeyId       string `env:"TOKEN_KMS_KEY_ID"`
	AuthMailSender      string `env:"AUTH_MAIL_SENDER"`
	CaptchaBucket       string `env:"CAPTCHA_BUCKET"`
	CaptchaBucketPrefix string `env:"CAPTCHA_BUCKET_PREFIX"`
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
		fmt.Fprintf(os.Stderr, "cannot load aws config: %v", err)
		os.Exit(1)
	}
	dynamoClient := dynamodb.NewFromConfig(awsCfg)
	kmsClient := kms.NewFromConfig(awsCfg)
	s3Client := s3.NewFromConfig(awsCfg)
	sesClient := ses.NewFromConfig(awsCfg)

	emailModel := email.New(dynamoClient, cfg.Table)
	userModel := user.New(dynamoClient, cfg.Table)
	tokenCtrl := token.New(cfg.TokenIssuer, jwtkms.NewKMSConfig(kmsClient, cfg.TokenKmsKeyId, false))
	authCtrl := auth.New(emailModel, sesClient, cfg.AuthMailSender)

	mux := http.NewServeMux()
	mux.Handle(
		authenticationconnect.NewAuthenticationServiceHandler(authentication.New(logger, tokenCtrl, authCtrl, emailModel)),
	)
	mux.Handle(
		managementconnect.NewManagementServiceHandler(management.New(logger, tokenCtrl, authCtrl, userModel, emailModel)),
	)

	// no I'm not responsible for this global setCustomStore mess :<
	captcha.SetCustomStore(captchastore.New(
		s3Client,
		logger, 2*time.Second,
		cfg.CaptchaBucket, cfg.CaptchaBucketPrefix,
	))

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
