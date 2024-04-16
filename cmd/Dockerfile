FROM golang:1.21.3 AS builder

WORKDIR /go/src/financial

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make build_financial_api

FROM alpine:3.16
WORKDIR /app/

COPY --from=builder /go/src/financial/bin .
COPY --from=builder /go/src/financial/cmd/.env .
COPY --from=builder /go/src/financial/database/migrations ./migrations/financial

EXPOSE 80
EXPOSE 443
ENTRYPOINT ["./financial"]