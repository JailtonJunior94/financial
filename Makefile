generate_dotenv:
	@echo "Generating .env file..."
	@cp cmd/.env.example cmd/.env

golint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0
	golangci-lint run ./...

.PHONY: migrate
migrate:
	@migrate create -ext sql -dir database/migrations -format unix $(NAME)

.PHONY: mockery
generate_mock:
	@mockery --dir=internal/user/domain/interfaces --name=UserRepository --filename=user_repository_mock.go --output=internal/user/infrastructure/repositories/mock --outpkg=repositoryMock
	@mockery --dir=internal/category/domain/interfaces --name=CategoryRepository --filename=category_repository_mock.go --output=internal/category/infrastructure/repositories/mock --outpkg=repositoryMock
	@mockery --dir=internal/budget/domain/interfaces --name=BudgetRepository --filename=budget_repository_mock.go --output=internal/budget/infrastructure/repositories/mock --outpkg=repositoryMock

test:
	go test -coverprofile coverage.out -failfast ./...
	go tool cover -func coverage.out | grep total

cover:
	go tool cover -html=coverage.out

build_financial_api:
	@echo "Compiling Financial API..."
	@CGO_ENABLED=0 go build -ldflags="-w -s" -o ./bin/financial ./cmd/main.go

start_docker_without_api:
	docker compose -f deployment/docker-compose.yml up --build -d financial_migration mysql rabbitmq grafana prometheus otel_collector jaeger 

start_docker:
	docker compose -f deployment/docker-compose.yml up --build -d

stop_docker:
	docker compose -f deployment/docker-compose.yml down