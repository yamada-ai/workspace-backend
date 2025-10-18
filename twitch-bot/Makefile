.PHONY: help install gen-client run test lint

help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

install: ## Install Python dependencies
	pip install -r requirements.txt

gen-client: ## Generate OpenAPI client from spec
	../scripts/gen_openapi_client_python.sh

run: ## Run the twitch bot locally
	python -m app.main

test: ## Run tests (when implemented)
	pytest tests/

lint: ## Run linters
	ruff check app/
	mypy app/
