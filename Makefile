.PHONY: help gen gen-sqlc gen-openapi test lint run migrate-up migrate-down migrate-create docker-up docker-down docker-logs dev-up build

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

gen: gen-sqlc gen-openapi ## Generate all code (sqlc + OpenAPI)

gen-sqlc: ## Generate sqlc code
	@bash scripts/gen_sqlc.sh

gen-openapi: ## Generate OpenAPI code
	@bash scripts/gen_openapi.sh

test: ## Run all tests (unit + integration)
	DATABASE_URL='postgres://postgres:postgres@localhost:5432/workspace_test?sslmode=disable' \
		go test -v -p 1 -parallel 1 -race -cover ./...

test-unit: ## Run only unit tests (fast, no DB required)
	go test -v -short -race -cover ./...

test-integration: ## Run only integration tests (requires DB)
	DATABASE_URL='postgres://postgres:postgres@localhost:5432/workspace_test?sslmode=disable' \
		go test -v -p 1 -race -cover -run Integration ./...

test-coverage: ## Run tests with coverage report
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run linter
	golangci-lint run

build: ## Build the application
	go build -o bin/work-tracker ./cmd/work-tracker

run: ## Run the application locally (auto-migrates DB)
	DATABASE_URL='postgres://postgres:postgres@localhost:5432/workspace?sslmode=disable' \
		go run cmd/work-tracker/main.go

migrate-up: ## Run database migrations up
	@bash scripts/migrate.sh

migrate-down: ## Run database migrations down
	migrate -path migrations -database "${DATABASE_URL:-postgres://postgres:postgres@localhost:5432/workspace?sslmode=disable}" down

migrate-create: ## Create a new migration (usage: make migrate-create NAME=add_users)
	migrate create -ext sql -dir migrations -seq $(NAME)

# Docker commands
docker-up: ## Start all services with Docker Compose
	docker compose -f infrastructure/compose.dev.yml up -d

docker-down: ## Stop all services
	docker compose -f infrastructure/compose.dev.yml down

docker-logs: ## Show logs from all services
	docker compose -f infrastructure/compose.dev.yml logs -f

docker-build: ## Build Docker image
	docker compose -f infrastructure/compose.dev.yml build

dev-up: ## Start development environment (DB + migrations)
	@bash scripts/dev_up.sh
