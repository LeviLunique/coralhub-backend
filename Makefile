APP_ENV ?= development
COMPOSE ?= docker compose

.PHONY: build ci compose-down compose-up fmt fmt-check govulncheck lint run-api run-worker sqlc staticcheck test vet

fmt:
	gofmt -w ./cmd ./internal

fmt-check:
	test -z "$$(gofmt -l ./cmd ./internal)"

vet:
	go vet ./...

staticcheck:
	go run honnef.co/go/tools/cmd/staticcheck@2025.1.1 ./...

lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8 run

govulncheck:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

test:
	go test ./...

build:
	go build ./cmd/api
	go build ./cmd/worker

sqlc:
	sqlc generate

ci: fmt-check sqlc vet staticcheck lint govulncheck test build

run-api:
	APP_ENV=$(APP_ENV) go run ./cmd/api

run-worker:
	APP_ENV=$(APP_ENV) go run ./cmd/worker

compose-up:
	$(COMPOSE) up -d postgres minio minio-create-bucket

compose-down:
	$(COMPOSE) down
