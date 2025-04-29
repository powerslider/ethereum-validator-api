# This loads all the env vars from the passed envfile (which defaults to .env)
# and exports them so that the child processes of make (like shell scripts or Go builds)
# can see these env vars.
envfile ?= .env
-include $(envfile)
ifneq ("$(wildcard $(envfile))","")
	export $(shell sed 's/=.*//' $(envfile))
endif

GOLANGCI_VERSION:=2.1.5
PROJECT_NAME:=ethereum-validator-api
SWAG_VERSION:=v1.16.4

.PHONY: install
install:
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v${GOLANGCI_VERSION}
	# Install Swag tool for Swagger API documentation generation.
	@go install github.com/swaggo/swag/cmd/swag@${SWAG_VERSION}

.PHONY: all
all: clean init lint test

.PHONY: init
init:
	@cp .env.dist .env

.PHONY: lint
lint:
	@echo ">>> Performing golang code linting..."
	@golangci-lint run --config=.golangci.yml

.PHONY: test
test:
	@echo ">>> Running Unit Tests..."
	@go test -v -race ./...

.PHONY: cover-test
cover-test:
	@echo ">>> Running Tests with Coverage..."
	@go test -v -race ./... -coverprofile=coverage.out -covermode=atomic
	@go tool cover -html=coverage.out

.PHONY: run-server
run-server:
	@echo ">>> Running ${PROJECT_NAME} API server..."
	@go run ./cmd/server/main.go

.PHONY: compose-up
compose-up:
	@docker compose --env-file .env up --build -d

.PHONY: compose-down
compose-down:
	@docker compose down --volumes --remove-orphans

.PHONY: api-docs
api-docs:
	@echo ">>> Generating Swagger API documentation..."
	@swag init --generalInfo cmd/server/main.go

.PHONY: clean
clean:
	@echo ">>> Removing .env file..."
	@rm -rf .env
