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

.PHONY: test
test:
	@echo "Running tests..."
	go test -count=1 -race -covermode=atomic -coverprofile=coverage.out ./...

cover:
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out

start_minimal:
	docker compose -f deployment/docker-compose.yml up --build -d financial_migration cockroachdb rabbitmq otel-lgtm
	
start_docker:
	docker compose -f deployment/docker-compose.yml up --build -d

stop_docker:
	docker compose -f deployment/docker-compose.yml down