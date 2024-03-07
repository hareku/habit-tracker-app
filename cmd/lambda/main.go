package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/guregu/dynamo"
	"github.com/hareku/habit-tracker-app/pkg/habit"
)

//go:embed .secrets/habittrackerapp-cred.json
var googleCred []byte

//go:embed .secrets/csrf-token.key
var csrfKey []byte

var handler *httpadapter.HandlerAdapter

func init() {
	fa, err := habit.NewFirebaseAuthenticator(googleCred)
	if err != nil {
		panic(fmt.Errorf("init firebase authenticator: %w", err))
	}

	secure, err := strconv.ParseBool(os.Getenv("SECURE"))
	if err != nil {
		panic(fmt.Errorf("parse str as bool: %w", err))
	}
	log.Printf("[config] secure is %+v", secure)

	var endpoint *string
	if e := os.Getenv("AWS_ENDPOINT"); e != "" {
		endpoint = aws.String(e)
		log.Printf("[config] aws endpoint is %q", *endpoint)
	}
	db := dynamo.New(session.New(), &aws.Config{
		Region:   aws.String("ap-northeast-1"),
		Endpoint: endpoint,
	})
	repo := &habit.DynamoRepository{
		DB:    db,
		Table: db.Table("HabitTrackerApp"),
	}

	handler = httpadapter.New(habit.NewHTTPHandler(&habit.NewHTTPHandlerInput{
		AuthMiddleware: habit.NewAuthMiddleware(fa),
		CSRFMiddleware: habit.NewCSRFMiddleware(csrfKey, secure),
		Authenticator:  fa,
		Repository:     repo,
		Secure:         secure,
	}))
}

func main() {
	lambda.Start(handler.ProxyWithContext)
}
