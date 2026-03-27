APP_ENV ?= development
COMPOSE ?= docker compose

.PHONY: build compose-down compose-up fmt run-api run-worker sqlc test vet

fmt:
	gofmt -w ./cmd ./internal

vet:
	go vet ./...

test:
	go test ./...

build:
	go build ./cmd/api
	go build ./cmd/worker

sqlc:
	sqlc generate

run-api:
	APP_ENV=$(APP_ENV) go run ./cmd/api

run-worker:
	APP_ENV=$(APP_ENV) go run ./cmd/worker

compose-up:
	$(COMPOSE) up -d postgres minio minio-create-bucket

compose-down:
	$(COMPOSE) down
