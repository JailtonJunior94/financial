.PHONY: help
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ============================================================================
# ENVIRONMENT SETUP
# ============================================================================

.PHONY: env
env: ## Create .env file from example (used by both local and Docker)
	@echo "📝 Creating environment file..."
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "✅ .env created from .env.example"; \
		echo ""; \
		echo "📝 Note: .env is used for both:"; \
		echo "  • Local development (localhost)"; \
		echo "  • Docker (hosts overridden automatically)"; \
	else \
		echo "✓ .env already exists"; \
	fi

.PHONY: env-migrate
env-migrate: ## Migrate from old .env structure to new (cleanup duplicates)
	@echo "🔄 Migrating to new .env structure..."
	@echo ""
	@echo "Old structure (deprecated):"
	@echo "  • .env.example"
	@echo "  • cmd/.env.example + cmd/.env"
	@echo "  • deployment/.env.docker + deployment/.env"
	@echo ""
	@echo "New structure (current):"
	@echo "  • .env.example (single template)"
	@echo "  • .env (single config)"
	@echo ""
	@if [ -f cmd/.env ] || [ -f cmd/.env.example ] || [ -f deployment/.env ] || [ -f deployment/.env.docker ]; then \
		echo "⚠️  Found old files. Removing duplicates..."; \
		rm -f cmd/.env.example cmd/.env; \
		rm -f deployment/.env.docker deployment/.env; \
		echo "✅ Old files removed"; \
		echo ""; \
		if [ ! -f .env ]; then \
			cp .env.example .env; \
			echo "✅ Created new .env from template"; \
		else \
			echo "✓ .env already exists (keeping it)"; \
		fi; \
		echo ""; \
		echo "✅ Migration complete!"; \
	else \
		echo "✅ Already using new structure. No migration needed."; \
	fi

# ============================================================================
# DOCUMENTATION
# ============================================================================

.PHONY: docs-generate
docs-generate: ## Generate Swagger/OpenAPI docs from source annotations (requires swag)
	@echo "📚 Generating Swagger documentation..."
	@if ! command -v swag &> /dev/null; then \
		echo "📦 Installing swag CLI..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	@swag init \
		-g cmd/main.go \
		--output docs \
		--parseInternal \
		--generatedTime false
	@echo "✅ Docs generated: docs/swagger.json | docs/swagger.yaml | docs/docs.go"
	@echo "📖 Serve via: http://localhost:8080/swagger/index.html (após adicionar rota swagger UI)"

# ============================================================================
# BUILD
# ============================================================================

.PHONY: build
build: ## Build the application binary
	@echo "🔨 Compiling Financial API..."
	@CGO_ENABLED=0 go build -ldflags="-w -s" -o ./bin/financial ./cmd/main.go
	@echo "✅ Binary created: ./bin/financial"

# ============================================================================
# INFRASTRUCTURE (for local development with VSCode debug)
# ============================================================================

.PHONY: infra-up
infra-up: env ## Start infrastructure only (DB, RabbitMQ, Observability)
	@echo "🚀 Starting infrastructure services..."
	@cd deployment && docker-compose up -d cockroachdb rabbitmq otel-lgtm
	@echo "⏳ Waiting for services to be healthy..."
	@sleep 5
	@echo ""
	@echo "✅ Infrastructure ready!"
	@echo ""
	@echo "📊 Services available:"
	@echo "  • Database:    http://localhost:8081 (CockroachDB UI)"
	@echo "  • RabbitMQ:    http://localhost:15672 (guest / rabbitmq@dev)"
	@echo "  • Grafana:     http://localhost:3000 (admin / admin@dev)"
	@echo ""
	@echo "🐛 Now you can run VSCode debug (F5) or use 'make run-api'"

.PHONY: infra-down
infra-down: ## Stop infrastructure services
	@echo "🛑 Stopping infrastructure services..."
	@cd deployment && docker-compose stop cockroachdb rabbitmq otel-lgtm
	@echo "✅ Infrastructure stopped"

.PHONY: infra-logs
infra-logs: ## Show infrastructure logs
	@cd deployment && docker-compose logs -f cockroachdb rabbitmq otel-lgtm

.PHONY: infra-restart
infra-restart: infra-down infra-up ## Restart infrastructure services

.PHONY: infra-clean
infra-clean: ## Stop and remove infrastructure (includes volumes)
	@echo "🧹 Cleaning infrastructure..."
	@cd deployment && docker-compose down -v
	@echo "✅ Infrastructure cleaned (volumes removed)"

# ============================================================================
# FULL APPLICATION (Docker Compose)
# ============================================================================

.PHONY: app-up
app-up: env ## Start complete application stack (Migration + API + Consumer + Worker + Infra)
	@echo "🚀 Starting complete application stack..."
	@cd deployment && docker-compose up --build -d
	@echo "⏳ Waiting for services to be ready..."
	@sleep 10
	@echo ""
	@echo "✅ Application stack is running!"
	@echo ""
	@echo "📊 Services available:"
	@echo "  • API:         http://localhost:8080"
	@echo "  • Database:    http://localhost:8081 (CockroachDB UI)"
	@echo "  • RabbitMQ:    http://localhost:15672 (guest / rabbitmq@dev)"
	@echo "  • Grafana:     http://localhost:3000 (admin / admin@dev)"
	@echo ""
	@echo "📝 Check logs with: make app-logs"
	@echo "🛑 Stop with: make app-down"

.PHONY: app-down
app-down: ## Stop complete application stack
	@echo "🛑 Stopping application stack..."
	@cd deployment && docker-compose down
	@echo "✅ Application stopped"

.PHONY: app-restart
app-restart: ## Restart complete application stack
	@echo "🔄 Restarting application stack..."
	@cd deployment && docker-compose restart financial_api financial_consumer financial_worker
	@echo "✅ Application services restarted"

.PHONY: app-logs
app-logs: ## Show application logs (API, Consumer, Worker)
	@cd deployment && docker-compose logs -f financial_api financial_consumer financial_worker

.PHONY: app-logs-all
app-logs-all: ## Show all logs (including infrastructure)
	@cd deployment && docker-compose logs -f

.PHONY: app-rebuild
app-rebuild: ## Rebuild and restart application services
	@echo "🔨 Rebuilding application..."
	@cd deployment && docker-compose up --build -d financial_api financial_consumer financial_worker
	@echo "✅ Application rebuilt and restarted"

.PHONY: app-clean
app-clean: ## Stop and remove everything (includes volumes)
	@echo "🧹 Cleaning everything..."
	@cd deployment && docker-compose down -v
	@echo "✅ Everything cleaned (volumes removed)"

# ============================================================================
# DATABASE MIGRATIONS
# ============================================================================

.PHONY: migrate-create
migrate-create: ## Create a new migration (usage: make migrate-create NAME=create_users_table)
	@if [ -z "$(NAME)" ]; then \
		echo "❌ Error: NAME is required"; \
		echo "Usage: make migrate-create NAME=create_users_table"; \
		exit 1; \
	fi
	@echo "📝 Creating migration: $(NAME)"
	@migrate create -ext sql -dir database/migrations -format unix $(NAME)
	@echo "✅ Migration files created"

.PHONY: migrate-up
migrate-up: env ## Run migrations (local - requires infra-up)
	@echo "🔄 Running migrations..."
	@go run ./cmd/main.go migrate
	@echo "✅ Migrations completed"

.PHONY: migrate-docker
migrate-docker: ## Run migrations in Docker
	@echo "🔄 Running migrations in Docker..."
	@cd deployment && docker-compose up financial_migration
	@echo "✅ Migrations completed"

# ============================================================================
# LOCAL DEVELOPMENT (with infrastructure running)
# ============================================================================

.PHONY: run-api
run-api: env ## Run API locally (requires infra-up)
	@echo "🚀 Starting API locally..."
	@go run ./cmd/main.go api

.PHONY: run-consumer
run-consumer: env ## Run Consumer locally (requires infra-up)
	@echo "🚀 Starting Consumer locally..."
	@go run ./cmd/main.go consumer

.PHONY: run-worker
run-worker: env ## Run Worker locally (requires infra-up)
	@echo "🚀 Starting Worker locally..."
	@go run ./cmd/main.go worker

# ============================================================================
# TESTING
# ============================================================================

.PHONY: test
test: ## Run unit tests
	@echo "🧪 Running unit tests..."
	@go test -short -count=1 -race -covermode=atomic -coverprofile=coverage-unit.out ./...
	@go tool cover -func=coverage-unit.out | grep total

.PHONY: test-integration
test-integration: ## Run integration tests
	@echo "🧪 Running integration tests..."
	@go test -tags=integration -count=1 -race -covermode=atomic -coverprofile=coverage-integration.out ./...

.PHONY: test-all
test-all: test test-integration ## Run all tests
	@echo "✅ All tests completed"

.PHONY: test-cover
test-cover: ## Run tests and show coverage in browser
	@echo "🧪 Running tests with coverage..."
	@go test -short -count=1 -race -covermode=atomic -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out

.PHONY: test-cover-html
test-cover-html: ## Generate HTML coverage reports
	@echo "📊 Generating HTML coverage reports..."
	@go tool cover -html=coverage-unit.out -o coverage-unit.html
	@test -f coverage-integration.out && go tool cover -html=coverage-integration.out -o coverage-integration.html || echo "No integration coverage found"
	@echo "✅ Coverage reports: coverage-unit.html, coverage-integration.html"

# ============================================================================
# CODE QUALITY
# ============================================================================

.PHONY: lint
lint: ## Run linter
	@echo "🔍 Running linter..."
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "📦 Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.2.1; \
	fi
	@GOGC=20 golangci-lint run --config .golangci.yml ./...
	@echo "✅ Linting completed"

.PHONY: fmt
fmt: ## Format code
	@echo "💅 Formatting code..."
	@go fmt ./...
	@echo "✅ Code formatted"

.PHONY: vet
vet: ## Run go vet
	@echo "🔍 Running go vet..."
	@go vet ./...
	@echo "✅ Vet completed"

.PHONY: mocks
mocks: ## Generate mocks
	@echo "🎭 Generating mocks..."
	@if ! command -v mockery &> /dev/null; then \
		echo "📦 Installing mockery..."; \
		go install github.com/vektra/mockery/v3@v3.5.0; \
	fi
	@mockery
	@echo "✅ Mocks generated"

.PHONY: check
check: fmt vet lint test-all ## Run all checks (format, vet, lint, tests)
	@echo "✅ All checks passed!"

# ============================================================================
# HEALTH CHECKS
# ============================================================================

.PHONY: health
health: ## Check application health
	@echo "🏥 Checking application health..."
	@curl -f http://localhost:8080/health || echo "❌ API is not responding"
	@echo ""
	@curl -f http://localhost:15672/api/health/checks/alarms || echo "❌ RabbitMQ is not responding"
	@echo ""
	@curl -f http://localhost:3000/api/health || echo "❌ Grafana is not responding"

.PHONY: status
status: ## Show Docker services status
	@echo "📊 Services status:"
	@cd deployment && docker-compose ps

# ============================================================================
# UTILITIES
# ============================================================================

.PHONY: logs-api
logs-api: ## Show API logs
	@cd deployment && docker-compose logs -f financial_api

.PHONY: logs-consumer
logs-consumer: ## Show Consumer logs
	@cd deployment && docker-compose logs -f financial_consumer

.PHONY: logs-worker
logs-worker: ## Show Worker logs
	@cd deployment && docker-compose logs -f financial_worker

.PHONY: logs-db
logs-db: ## Show database logs
	@cd deployment && docker-compose logs -f cockroachdb

.PHONY: logs-rabbitmq
logs-rabbitmq: ## Show RabbitMQ logs
	@cd deployment && docker-compose logs -f rabbitmq

.PHONY: shell-api
shell-api: ## Open shell in API container
	@cd deployment && docker-compose exec financial_api sh

.PHONY: shell-db
shell-db: ## Open CockroachDB SQL shell
	@cd deployment && docker-compose exec cockroachdb cockroach sql --insecure

# ============================================================================
# QUICK START WORKFLOWS
# ============================================================================

.PHONY: dev-setup
dev-setup: env infra-up migrate-up ## Complete local dev setup (env + infra + migrations)
	@echo ""
	@echo "✅ Development environment ready!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Open VSCode and press F5 to debug"
	@echo "  2. Or run: make run-api"
	@echo ""

.PHONY: dev-start
dev-start: infra-up ## Quick start: infrastructure for VSCode debug
	@echo ""
	@echo "✅ Ready for development!"
	@echo "Press F5 in VSCode to start debugging"

.PHONY: docker-start
docker-start: app-up ## Quick start: complete Docker stack

.PHONY: clean-all
clean-all: app-clean ## Clean everything (stops containers and removes volumes)
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf bin coverage*.out coverage*.html
	@echo "✅ Everything cleaned!"

# ============================================================================
# DEFAULT TARGET
# ============================================================================

.DEFAULT_GOAL := help
