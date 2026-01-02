dotenv:
	@echo "Generating .env file..."
	@cp cmd/.env.example cmd/.env

build:
	@echo "Compiling Financial API..."
	@CGO_ENABLED=0 go build -ldflags="-w -s" -o ./bin/financial ./cmd/main.go

.PHONY: migrate
migrate:
	@migrate create -ext sql -dir database/migrations -format unix $(NAME)

lint:
	@echo "Running linter..."
	@echo "Installing golangci-lint..."
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.2.1
	GOGC=20 golangci-lint run --config .golangci.yml ./...

.PHONY: mockery
mocks:
	@echo "Generating mocks..."
	go install github.com/vektra/mockery/v3@v3.5.0
	mockery

.PHONY: test-unit
test-unit:
	@echo "Running unit tests..."
	@go test -short -count=1 -race -covermode=atomic -coverprofile=coverage-unit.out ./...
	@go tool cover -func=coverage-unit.out | grep total

.PHONY: test-integration
test-integration:
	@echo "Running integration tests..."
	@go test -tags=integration -count=1 -race -covermode=atomic -coverprofile=coverage-integration.out ./...

.PHONY: test-all
test-all: test-unit test-integration
	@echo "All tests completed successfully"

.PHONY: test
test: test-unit
	@echo "Unit tests completed"

.PHONY: cover-html
cover-html:
	@echo "Generating HTML coverage reports..."
	@go tool cover -html=coverage-unit.out -o coverage-unit.html
	@test -f coverage-integration.out && go tool cover -html=coverage-integration.out -o coverage-integration.html || echo "No integration coverage found"
	@echo "Coverage reports: coverage-unit.html, coverage-integration.html"

.PHONY: cover
cover: cover-html
	@echo "Opening unit test coverage..."
	@go tool cover -html=coverage-unit.out

.PHONY: check
check: lint test-all
	@echo "All checks passed successfully"

start_minimal:
	docker compose -f deployment/docker-compose.yml up --build -d cockroachdb rabbitmq otel-lgtm
	
start_docker:
	docker compose -f deployment/docker-compose.yml up --build -d

stop_docker:
	docker compose -f deployment/docker-compose.yml down