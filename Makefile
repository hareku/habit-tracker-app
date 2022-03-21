.PHONY: build

build:
	sam build

prepare:
	go run ./cmd/generate-key cmd/lambda/.secrets/csrf-token.key
	docker-compose up -d --no-recreate
	aws dynamodb create-table --cli-input-json file://dynamodb.json --endpoint-url http://localhost:8000

dev:
	docker-compose up -d --no-recreate
	sam build
	sam local start-api --env-vars local-env.json --docker-network habit-tracker-app_default

deploy:
	sam build
	sam deploy
