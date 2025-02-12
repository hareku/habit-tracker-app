package main

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/hareku/habit-tracker-app/internal/api"
	"github.com/hareku/habit-tracker-app/internal/applog"
	"github.com/hareku/habit-tracker-app/internal/auth"
	"github.com/hareku/habit-tracker-app/internal/repository"
)

var (
	//go:embed .secrets/*
	secretsDir embed.FS
	// handler is the main handler for the lambda function
	handler *httpadapter.HandlerAdapter
)

func init() {
	slog.SetDefault(slog.New(
		applog.NewContextValueLogHandler(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelInfo,
		})),
	))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	handler, err = newHandler(ctx)
	if err != nil {
		slog.ErrorContext(ctx, fmt.Errorf("init handler: %w", err).Error())
		os.Exit(1)
	}
}

func newHandler(ctx context.Context) (*httpadapter.HandlerAdapter, error) {
	googleCred, err := secretsDir.ReadFile(".secrets/habittrackerapp-cred.json")
	if err != nil {
		return nil, fmt.Errorf("open google cred: %w", err)
	}

	fa, err := auth.NewFirebaseAuthenticator(googleCred)
	if err != nil {
		return nil, fmt.Errorf("init firebase authenticator: %w", err)
	}

	secure, err := strconv.ParseBool(os.Getenv("SECURE"))
	if err != nil {
		return nil, fmt.Errorf("parse str as bool: %w", err)
	}
	slog.Info("Loaded SECURE env", slog.Bool("secure", secure))

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}
	cfg.Region = "ap-northeast-1"
	if e := os.Getenv("AWS_ENDPOINT"); e != "" {
		cfg.BaseEndpoint = aws.String(e)
		slog.Info("Loaded AWS_ENDPOINT env", slog.String("endpoint", e))
	}

	csrfKey, err := secretsDir.ReadFile(".secrets/csrf-token.key")
	if err != nil {
		return nil, fmt.Errorf("open csrf key: %w", err)
	}

	return httpadapter.New(api.NewHTTPHandler(&api.NewHTTPHandlerInput{
		AuthMiddleware: api.NewAuthMiddleware(fa),
		CSRFMiddleware: api.NewCSRFMiddleware(csrfKey, secure),
		Authenticator:  fa,
		Repository: &repository.DynamoRepository{
			Client:    dynamodb.NewFromConfig(cfg),
			TableName: "HabitTrackerApp",
		},
		Secure: secure,
	})), nil
}

func main() {
	lambda.Start(handler.ProxyWithContext)
}
