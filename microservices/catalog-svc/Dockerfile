# Multi-stage build for catalog-svc
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy shared library and service
COPY shared/ /app/shared/
COPY microservices/catalog-svc/ /app/microservices/catalog-svc/

WORKDIR /app/microservices/catalog-svc
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/catalog-svc ./cmd/api/main.go

# Production minimal container
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/bin/catalog-svc /app/catalog-svc

EXPOSE 8081

ENTRYPOINT ["/app/catalog-svc"]
