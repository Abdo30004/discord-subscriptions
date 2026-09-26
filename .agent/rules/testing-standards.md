# Testing & Verification Standards

> **Directive**: This rule governs automated testing and build validation for all pull requests and agent edits.

---

## 1. Test Execution Commands

Agents must verify tests before finalizing any code edits:

```bash
# Test shared library
cd shared && go test -v ./... && cd ..

# Test microservices individually
cd microservices/auth-svc && go test -v ./... && cd ../..
cd microservices/catalog-svc && go test -v ./... && cd ../..
cd microservices/billing-svc && go test -v ./... && cd ../..
cd microservices/deploy-svc && go test -v ./... && cd ../..
cd microservices/monitor-svc && go test -v ./... && cd ../..

# Build check for all Go service executables
cd microservices/auth-svc && go build -o /dev/null ./cmd/api && cd ../..
cd microservices/catalog-svc && go build -o /dev/null ./cmd/api && cd ../..
cd microservices/billing-svc && go build -o /dev/null ./cmd/api && cd ../..
cd microservices/deploy-svc && go build -o /dev/null ./cmd/api && cd ../..
cd microservices/monitor-svc && go build -o /dev/null ./cmd/api && cd ../..

# Compile TypeScript Manager Bot
cd bots/manager-bot && npm run build && cd ../..

# Compile Next.js Frontend
cd frontend && npm run build && cd ..
```

---

## 2. Unit Testing Principles

1. **Domain Unit Tests**:
   - Write tests for business rules (e.g., promo code validation, subscription expiration check, pod naming format, token state transitions).
   - Domain unit tests must execute in < 10ms with zero network or database dependencies.
2. **Mocking Ports**:
   - Mock repositories and external messaging brokers using Go interfaces defined in `internal/core/ports`.
   - Never require a live PostgreSQL or RabbitMQ instance for standard unit tests (`go test -v ./...`).
3. **Table-Driven Tests**:
   - Use Go table-driven tests for multiple input cases (valid, invalid, edge cases).

---

## 3. Integration & Health Checks

When testing against Docker Compose:
- PostgreSQL must respond to `pg_isready -U postgres`.
- RabbitMQ must respond to `rabbitmq-diagnostics -q ping`.
- Vault must respond to `vault status` on `http://localhost:8200`.
- All microservices expose a `/health` endpoint returning `{"status": "ok"}`.
