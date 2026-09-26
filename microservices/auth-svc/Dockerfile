# Multi-stage build for auth-svc
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy shared library and service
COPY shared/ /app/shared/
COPY microservices/auth-svc/ /app/microservices/auth-svc/

WORKDIR /app/microservices/auth-svc
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/auth-svc ./cmd/api/main.go

# Production minimal container
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/bin/auth-svc /app/auth-svc

EXPOSE 8080

ENTRYPOINT ["/app/auth-svc"]
