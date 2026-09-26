# Multi-stage build for deploy-svc
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy shared library and service
COPY shared/ /app/shared/
COPY microservices/deploy-svc/ /app/microservices/deploy-svc/

WORKDIR /app/microservices/deploy-svc
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/deploy-svc ./cmd/api/main.go

# Production minimal container
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/bin/deploy-svc /app/deploy-svc

EXPOSE 8083

ENTRYPOINT ["/app/deploy-svc"]
