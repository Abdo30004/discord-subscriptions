# Multi-stage build for monitor-svc
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy shared library and service
COPY shared/ /app/shared/
COPY microservices/monitor-svc/ /app/microservices/monitor-svc/

WORKDIR /app/microservices/monitor-svc
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/monitor-svc ./cmd/api/main.go

# Production minimal container
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/bin/monitor-svc /app/monitor-svc

EXPOSE 8084

ENTRYPOINT ["/app/monitor-svc"]
