# Development, Operations & Documentation Maintenance Guide

This guide provides step-by-step instructions for setting up the local development environment, executing automated test suites, building production artifacts, and strictly adhering to the **Self-Maintaining Documentation Protocol**.

---

## 1. Prerequisites & Tooling

Ensure the following tools are installed on your host system:
- **Go**: Version 1.26 or newer
- **Node.js**: Version 20.x or newer (npm v10+)
- **Docker & Docker Compose**: Docker Desktop or Docker Engine v24+
- **PostgreSQL Client (Optional)**: `psql` for database inspection

---

## 2. Local Environment Setup

### 2.1 Configuration
Copy the sample environment file to `.env`:
```bash
cp .env.example .env
```
Ensure `SUPER_ADMIN_DISCORD_IDS` is populated with your Discord Snowflake ID (comma-separated if multiple administrators) to grant permanent Super Admin permissions upon signing in with Discord.

### 2.2 Development vs Production Docker Compose Environments

The platform provides isolated Compose files and environment configurations for development and production modes:

| Mode | Compose File | Environment File | Key Characteristics | NPM Command | Make Target |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Development** | `docker-compose.dev.yml` | `.env.dev` | `ENV=development`, `LOG_LEVEL=debug`, `PAYPAL_MODE=sandbox`, `ENABLE_SHARDING=false`, public dashboard on port `8090` | `npm run compose:dev` | `make up-dev` |
| **Production** | `docker-compose.prod.yml` | `.env.prod` | `ENV=production`, `LOG_LEVEL=info`, JSON structured logs, `PAYPAL_MODE=live`, `ENABLE_SHARDING=true` (Discord sharding cluster), dynamic DB passwords, loopback-only dashboard `127.0.0.1:8090`, CPU & memory resource limits, `restart: always` | `npm run compose:prod` | `make up-prod` |

#### Running in Development Mode
```powershell
# Start dev stack
npm run compose:dev
# or: docker compose -f docker-compose.dev.yml --env-file .env.dev up -d

# Check health using PowerShell
docker compose -f docker-compose.dev.yml --env-file .env.dev ps --format json | ConvertFrom-Json | Select-Object Service, State, Health | Format-Table -AutoSize

# Stop dev stack
npm run compose:dev:down
```

#### Running in Production Mode
```powershell
# Start production stack
npm run compose:prod
# or: docker compose -f docker-compose.prod.yml --env-file .env.prod up -d

# Check health using PowerShell
docker compose -f docker-compose.prod.yml --env-file .env.prod ps --format json | ConvertFrom-Json | Select-Object Service, State, Health | Format-Table -AutoSize

# Verify endpoints with PowerShell
Invoke-RestMethod -Uri "http://localhost/api/health" -TimeoutSec 5
Invoke-RestMethod -Uri "http://localhost/api/v1/catalog/bots" -TimeoutSec 5

# Stop production stack
npm run compose:prod:down
```

### 2.3 Initializing Databases & Passwords
- **Development**: The PostgreSQL container initializes schemas from [`scripts/init-databases.sql`](file:///C:/Users/kasep/Desktop/discord-subscriptions/scripts/init-databases.sql).
- **Production**: In addition to `init-databases.sql`, [`scripts/set-passwords.sh`](file:///C:/Users/kasep/Desktop/discord-subscriptions/scripts/set-passwords.sh) is mounted as an entrypoint script to dynamically apply distinct, secure passwords from `.env.prod` to `auth_user`, `catalog_user`, `billing_user`, `deploy_user`, and `monitor_user`.

To manually reset or re-seed the databases:
```bash
docker compose down -v
docker compose up -d postgres
```

### 2.4 Service Ports Summary

| Component | Host Port | Management / Protocol |
| :--- | :--- | :--- |
| **PostgreSQL** | `5432` | Relational Database (5 DBs) |
| **RabbitMQ AMQP** | `5672` | AMQP Protocol |
| **RabbitMQ Management** | `15672` | Web Dashboard (`guest` / `guest`) |
| **HashiCorp Vault** | `8200` | Vault HTTP API (`token: root`) |
| **auth-svc** | `8080` | REST API |
| **catalog-svc** | `8081` | REST API |
| **billing-svc** | `8082` | REST API |
| **deploy-svc** | `8083` | REST API |
| **monitor-svc** | `8084` | REST API |
| **manager-bot** | `8085` | Health Probe (`/health`) |
| **frontend** | `3000` | Next.js Web Dashboard |
| **Traefik Gateway** | `80` | Unified Edge Gateway & Reverse Proxy |
| **Traefik Dashboard** | `8090` | Traefik Web UI & Traffic Visualizer |

### 2.5 Deploying to Kubernetes (Minikube / Production)
To deploy the entire production infrastructure, microservices, frontend, and RBAC to a Kubernetes cluster using Kustomize:
```bash
# 1. Preview generated manifests
kubectl kustomize k8s/

# 2. Apply all resources (namespaces, RBAC, storage, microservices, ingress)
kubectl apply -k k8s/

# 3. Check status of all pods in the platform namespace
kubectl -n platform get pods -w
```
For detailed explanations of namespaces, RBAC privileges, and network isolation policies, see [KUBERNETES.md](./KUBERNETES.md).

---

## 3. Automated Testing & Build Validation

Before committing code or submitting pull requests, run the complete verification suite:

### 3.1 Go Unit Tests
Run all unit tests across all Go workspace modules:
```powershell
go test -v ./shared/...
go test -v ./microservices/auth-svc/...
go test -v ./microservices/catalog-svc/...
go test -v ./microservices/billing-svc/...
go test -v ./microservices/deploy-svc/...
go test -v ./microservices/monitor-svc/...
```

### 3.2 Compilation Checks (Go Executables)
Ensure that all microservice entrypoints compile without syntax or import errors:
```powershell
go build -o /dev/null ./microservices/auth-svc/cmd/api
go build -o /dev/null ./microservices/catalog-svc/cmd/api
go build -o /dev/null ./microservices/billing-svc/cmd/api
go build -o /dev/null ./microservices/deploy-svc/cmd/api
go build -o /dev/null ./microservices/monitor-svc/cmd/api
```

### 3.3 Manager Bot Build (TypeScript)
```powershell
cd bots/manager-bot
npm run build
cd ../..
```

### 3.4 Frontend Build (Next.js 16 Turbopack)
```powershell
cd frontend
npm run build
cd ..
```

### 3.5 Changelog Generation & Synchronization
Verify or regenerate `CHANGELOG.md` following Conventional Commits:
```powershell
# Update CHANGELOG.md from git commits
npm run changelog
# Or verify CHANGELOG.md is up-to-date
npm run changelog:check
```

---

## 4. Documentation Synchronization Protocol

> [!IMPORTANT]
> **MANDATORY POLICY FOR DEVELOPERS AND AI AGENTS**  
> Every modification to the codebase MUST be accompanied by updates to the corresponding documentation in `./docs/` during the same session.

### Step-by-Step Sync Checklist:

```text
[ ] 1. Did you add/modify a REST endpoint?
       --> Update docs/MICROSERVICES.md (add endpoint, request/response JSON, headers).
       --> If applicable, update docs/DEVELOPMENT_GUIDE.md port table.

[ ] 2. Did you add/modify a database table, column, constraint, or seed?
       --> Update scripts/init-databases.sql.
       --> Update docs/DATABASE_SCHEMAS.md (Mermaid ER diagram & column table).

[ ] 3. Did you add/modify a RabbitMQ event struct or routing key?
       --> Update shared/events/*.go.
       --> Update docs/EVENT_SPECIFICATION.md (routing table and JSON envelope).
       --> Update docs/ARCHITECTURE.md if inter-service flow changed.

[ ] 4. Did you add/modify a Discord slash command or UI component?
       --> Update docs/MANAGER_BOT.md (command table and interaction sequence).
       --> If it involves multi-bot fleet selection, update docs/MULTI_BOT_FLEET.md.

[ ] 5. Did you add/modify a turnkey token pool or zero-setup feature?
       --> Update docs/TOKEN_POOL_ZERO_SETUP.md (state machine, Vault path, admin API).

[ ] 6. Did you add/modify a Next.js frontend page or API client function?
       --> Update docs/FRONTEND.md (route specification and state description).

[ ] 7. Did you change Docker Compose or infrastructure dependencies?
       --> Update docker-compose.yml.
       --> Update docs/ARCHITECTURE.md and docs/DEVELOPMENT_GUIDE.md.

[ ] 8. Did you add/modify/retire any environment variable, secret, or port?
       --> Update .env.example (categorized with defaults and descriptions).
       --> Update docs/DEVELOPMENT_GUIDE.md environment reference table.
```

---

## 5. Antigravity Agent Skills Quick Reference

When interacting via AI coding assistants, you can invoke or recommend the following project skills:
- **`discord-fleet-ops`**: Executes build steps, launches Docker Compose stacks, seeds database tables, and inspects token pools.
- **`docs-sync`**: Performs automated audits of modified code against `./docs/` to ensure 100% documentation coverage and diagram accuracy.
