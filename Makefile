.PHONY: migrate
migrate:
	@migrate create -ext sql -dir database/migrations -format unix $(NAME)

test:
	go test -coverprofile coverage.out -failfast ./...
	go tool cover -func coverage.out | grep total

cover:
	go tool cover -html=coverage.out

build_financial_api:
	@echo "Compiling Financial API..."
	@CGO_ENABLED=0 go build -ldflags="-w -s" -o ./bin/financial ./cmd/main.go

start_without_api:
	docker-compose -f deployment/docker-compose.yml up -d mysql rabbitmq zipkin

start:
	docker-compose -f deployment/docker-compose.yml up --build -d

stop:
	docker-compose -f deployment/docker-compose.yml down

.PHONY: mockery
generate_mock:
	@mockery --dir=internal/domain/user/interfaces --name=UserRepository --filename=user_repository_mock.go --output=internal/infrastructure/user/repository/mock --outpkg=repositoryMock
	@mockery --dir=internal/domain/category/interfaces --name=CategoryRepository --filename=category_repository_mock.go --output=internal/infrastructure/category/repository/mock --outpkg=repositoryMock