# Multi-stage build for billing-svc
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy shared library and service
COPY shared/ /app/shared/
COPY microservices/billing-svc/ /app/microservices/billing-svc/

WORKDIR /app/microservices/billing-svc
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/billing-svc ./cmd/api/main.go

# Production minimal container
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/bin/billing-svc /app/billing-svc

EXPOSE 8082

ENTRYPOINT ["/app/billing-svc"]
