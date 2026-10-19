.PHONY: help setup docker-up docker-down docker-logs run build fmt fmt-check vet lint check test test-unit test-integration clean migrate-up migrate-down migrate-create migrate-version migrate-force migrate-drop

help: ## Show this help message
	@echo "Available commands:"
	@echo "  make setup           - Install dependencies"
	@echo "  make docker-up       - Start Docker services"
	@echo "  make docker-down     - Stop Docker services"
	@echo "  make docker-logs     - View Docker logs"
	@echo "  make run             - Run the API server (./cmd/server)"
	@echo "  make run-demo        - Run the bootstrap demo (main.go)"
	@echo "  make build           - Build the application"
	@echo "  make fmt             - Format code (gofmt -s + goimports)"
	@echo "  make fmt-check       - Fail if code is not formatted"
	@echo "  make vet             - Run go vet"
	@echo "  make lint            - Run golangci-lint (extended ruleset)"
	@echo "  make check           - fmt-check + vet + lint (quality gate)"
	@echo "  make test            - Run unit tests"
	@echo "  make test-unit       - Run unit tests (no external services)"
	@echo "  make test-integration - Run real integration tests (needs live PostgreSQL)"
	@echo "  make clean           - Clean build files"
	@echo ""
	@echo "Migration commands:"
	@echo "  make migrate-up      - Run all pending migrations"
	@echo "  make migrate-down    - Rollback last migration"
	@echo "  make migrate-create  - Create new migration (name=migration_name)"
	@echo "  make migrate-version - Show current migration version"
	@echo "  make migrate-force   - Force migration version (version=N)"
	@echo "  make migrate-drop    - Drop all tables (dangerous!)"

setup: ## Install dependencies
	go mod download
	go mod tidy

docker-up: ## Start Docker services
	docker-compose up -d
	@echo "Waiting for services to be ready..."
	@timeout /t 10 /nobreak >nul
	@echo "Services started! Access RabbitMQ Management UI at http://localhost:15672"

docker-down: ## Stop Docker services
	docker-compose down

docker-logs: ## View Docker logs
	docker-compose logs -f

docker-restart: ## Restart Docker services
	docker-compose restart

docker-ps: ## Show Docker services status
	docker-compose ps

run: ## Run the API server
	go run ./cmd/server

run-demo: ## Run the bootstrap demo (command main.go)
	go run main.go

run-examples: ## Run examples
	go run examples/basic_usage.go

build: ## Build the application
	go build -o bin/server ./cmd/server

fmt: ## Format code (gofmt -s + goimports)
	gofmt -s -w .
	goimports -w -local golang .

fmt-check: ## Fail if the code is not formatted
	@test -z "$$(gofmt -s -l . | grep -v '^$$')" || (echo "gofmt -s needed:"; gofmt -s -l .; exit 1)
	@test -z "$$(goimports -l -local golang .)" || (echo "goimports needed:"; goimports -l -local golang .; exit 1)

vet: ## Run go vet
	go vet ./...

lint: ## Run golangci-lint (extended ruleset)
	golangci-lint run ./...

check: fmt-check vet lint ## Quality gate: formatting + vet + lint

test: ## Run unit tests
	go test ./...

test-unit: ## Run unit tests (no external services)
	go test ./...

test-integration: ## Run real integration tests (needs live PostgreSQL; use TEST_DB_* env to override)
	go test -tags=integration -count=1 ./internal/app/...

clean: ## Clean build files
	if exist bin rmdir /s /q bin
	go clean

# Migration commands
migrate-up: ## Run all pending migrations
	go run cmd/migrate.go -action=up

migrate-down: ## Rollback last migration
	go run cmd/migrate.go -action=down -steps=1

migrate-create: ## Create new migration (usage: make migrate-create name=create_users)
	@if not defined name (echo Error: Please specify migration name: make migrate-create name=migration_name && exit /b 1)
	go run cmd/migrate.go -action=create -name=$(name)

migrate-version: ## Show current migration version
	go run cmd/migrate.go -action=version

migrate-force: ## Force migration version (usage: make migrate-force version=1)
	@if not defined version (echo Error: Please specify version: make migrate-force version=N && exit /b 1)
	go run cmd/migrate.go -action=force -version=$(version)

migrate-drop: ## Drop all tables (dangerous!)
	go run cmd/migrate.go -action=drop

env: ## Copy .env.example to .env
	if not exist .env copy .env.example .env

