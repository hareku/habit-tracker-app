package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/hareku/habit-tracker-app/pkg/habit"
)

//go:embed .secrets/habittrackerapp-cred.json
var googleCred []byte

//go:embed .secrets/csrf-token.key
var csrfKey []byte

var handler *httpadapter.HandlerAdapter

func init() {
	ctx := context.Background()

	fa, err := habit.NewFirebaseAuthenticator(googleCred)
	if err != nil {
		panic(fmt.Errorf("init firebase authenticator: %w", err))
	}

	secure, err := strconv.ParseBool(os.Getenv("SECURE"))
	if err != nil {
		panic(fmt.Errorf("parse str as bool: %w", err))
	}
	log.Printf("[config] secure is %+v", secure)

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(fmt.Errorf("load aws config: %w", err))
	}
	cfg.Region = "ap-northeast-1"
	if e := os.Getenv("AWS_ENDPOINT"); e != "" {
		cfg.BaseEndpoint = aws.String(e)
		log.Printf("[config] aws endpoint is %q", e)
	}

	handler = httpadapter.New(habit.NewHTTPHandler(&habit.NewHTTPHandlerInput{
		AuthMiddleware: habit.NewAuthMiddleware(fa),
		CSRFMiddleware: habit.NewCSRFMiddleware(csrfKey, secure),
		Authenticator:  fa,
		Repository: &habit.DynamoRepository{
			Client:    dynamodb.NewFromConfig(cfg),
			TableName: "HabitTrackerApp",
		},
		Secure: secure,
	}))
}

func main() {
	lambda.Start(handler.ProxyWithContext)
}
