FROM golang:1.26.0 AS builder

WORKDIR /go/src/financial

ENV USER=appuser 
ENV UID=1000

RUN adduser \    
    --disabled-password \    
    --gecos "" \    
    --home "/nonexistent" \    
    --shell "/sbin/nologin" \    
    --no-create-home \    
    --uid "${UID}" \    
    "${USER}"

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make build

FROM alpine:3.22.1
WORKDIR /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /go/src/financial/bin .
COPY --from=builder /go/src/financial/.env .
COPY --from=builder /go/src/financial/database/migrations ./migrations/financial

# Fix permissions for appuser
RUN chown -R 1000:1000 /app

EXPOSE 80
EXPOSE 443
USER appuser:appuser
ENTRYPOINT ["./financial"]