OPEN           ?= open
BIN_DIR        := $(CURDIR)/bin
.PHONY: help setup run test test-integration mocks clean-mocks build swagger migrate docker-up docker-down open-swagger clean

## help: Show available targets
help:
	@echo "Usage: make <target>"
	@echo ""
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/  /'

## setup: Copy sample configs, download deps, start infra, run migrations
setup:
	test -f .env             || cp samples/.env.sample .env
	test -f docker-compose.yaml || cp samples/docker-compose.yaml.sample docker-compose.yaml
	go mod download
	go install github.com/swaggo/swag/cmd/swag@latest
	$(MAKE) docker-up
	$(MAKE) migrate

## run: Run the API server
run:
	go run ./cmd/subscriptions-api

## test: Run all tests
test:
	go test ./...

## test-integration: Run integration tests (requires Docker)
test-integration:
	RUN_INTEGRATION_TESTS=1 go test ./tests -run Integration -count=1

## mocks: Generate mockery mocks
mocks:
	go run github.com/vektra/mockery/v2@v2.53.4 --config api/mockery.yaml

## build: Build the API binary to bin/
build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/subscriptions-api ./cmd/subscriptions-api

## swagger: Generate Swagger docs
swagger:
	go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/subscriptions-api/main.go -o docs

## migrate: Apply database migrations
migrate:
	go run ./cmd/migrations -up

## docker-up: Start PostgreSQL via docker compose
docker-up:
	docker compose --env-file .env up -d postgres

## docker-down: Stop all docker compose services
docker-down:
	docker compose --env-file .env down

## open-swagger: Open Swagger UI in the browser
open-swagger:
	@PORT=$$(grep -E '^SERVER_PORT=' .env 2>/dev/null | cut -d= -f2); \
	$(OPEN) "http://localhost:$${PORT:-1323}/swagger/index.html"

## clean: Remove build artifacts
clean:
	$(MAKE) clean-mocks
	rm -rf $(BIN_DIR)
	rm .env docker-compose.yaml
	rm -f mocks/*.go
