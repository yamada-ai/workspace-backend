.PHONY: help gen gen-sqlc gen-openapi test lint run migrate-up migrate-down migrate-create

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

gen: gen-sqlc gen-openapi ## Generate all code (sqlc + OpenAPI)

gen-sqlc: ## Generate sqlc code
	@bash scripts/gen_sqlc.sh

gen-openapi: ## Generate OpenAPI code
	@bash scripts/gen_openapi.sh

test: ## Run tests
	go test -v -race -cover ./...

test-short: ## Run short tests (skip integration tests)
	go test -v -short ./...

lint: ## Run linter
	golangci-lint run

run: ## Run the application
	go run cmd/work-tracker/main.go

migrate-up: ## Run database migrations up
	migrate -path migrations -database "postgresql://localhost:5432/workspace?sslmode=disable" up

migrate-down: ## Run database migrations down
	migrate -path migrations -database "postgresql://localhost:5432/workspace?sslmode=disable" down

migrate-create: ## Create a new migration (usage: make migrate-create NAME=add_users)
	migrate create -ext sql -dir migrations -seq $(NAME)
