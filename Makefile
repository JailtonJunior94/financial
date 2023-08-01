.PHONY: migrate
migrate:
	@migrate create -ext sql -dir database/migrations -format unix $(NAME)

test:
	go test -coverprofile coverage.out -failfast ./...
	go tool cover -func coverage.out | grep total

cover:
	go tool cover -html=coverage.outmak
