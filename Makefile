.PHONY: migrate
migrate:
	@migrate create -ext sql -dir database/migrations -format unix $(NAME)

test:
	go test -coverprofile coverage.out -failfast ./...
	go tool cover -func coverage.out | grep total

cover:
	go tool cover -html=coverage.outmak

build-financial-api:
	@echo "Compiling Financial API..."
	@CGO_ENABLED=0 go build -ldflags="-w -s" -o ./bin/financial ./cmd/main.go

start_without_api:
	docker-compose -f deployment/docker-compose.yml up -d mysql rabbitmq zipkin

start:
	docker-compose -f deployment/docker-compose.yml up --build -d

stop:
	docker-compose -f deployment/docker-compose.yml down