# AGENTS.md - Agent Operating Guidelines & Architecture Manual

> **Scope**: This repository contains the **Discord Bot Subscription & Turnkey Fleet Management Platform**.  
> All AI coding assistants (Antigravity, Claude Code, Cursor, Copilot, Codex, etc.) operating in this codebase MUST read, respect, and strictly adhere to the guidelines, invariants, and policies documented here.

---

## 1. System Overview & Technology Stack

The platform allows Discord server owners to purchase, provision, manage, and monitor single-tenant and shared Discord bots with optional **Zero-Setup Turnkey Delivery** (pre-provisioned bot tokens from an admin-managed token pool, 0-friction 1-click deployment) and **Multi-Subscription Fleet Tenancy** (running multiple distinct bots or multiple instances of the same bot type inside a single Discord Guild).

```mermaid
flowchart TB
    subgraph Clients [Clients & Interfaces]
        Frontend["Next.js 16 Dashboard & Store\n(Port 3000)"]
        ManagerBot["TypeScript Discord Manager Bot\n(Discord Gateway)"]
    end

    subgraph CoreServices [Golang Microservices (Clean Architecture)]
        AuthSvc["auth-svc :8080\n(OAuth2, JWT, Users)"]
        CatalogSvc["catalog-svc :8081\n(Templates, Plans, Pricing)"]
        BillingSvc["billing-svc :8082\n(Subscriptions, Promos, Vouchers)"]
        DeploySvc["deploy-svc :8083\n(K8s Orchestrator, Token Pool)"]
        MonitorSvc["monitor-svc :8084\n(Health Poller, Telemetry)"]
    end

    subgraph Infrastructure [Data & Messaging Infrastructure]
        Postgres[("PostgreSQL 16\n(Database-per-Service)")]
        RabbitMQ[["RabbitMQ 3.13\n(Topic Exchange: discord.events)"]]
        Vault[("HashiCorp Vault 1.16\n(Bot Tokens & Secrets)")]
        K8sCluster[("Kubernetes Cluster\n(Isolated Bot Pods)")]
    end

    Frontend --> AuthSvc & CatalogSvc & BillingSvc & DeploySvc & MonitorSvc
    ManagerBot --> CatalogSvc & BillingSvc & DeploySvc & MonitorSvc

    BillingSvc -->|Publishes Events| RabbitMQ
    RabbitMQ -->|Consumes Events| DeploySvc & MonitorSvc
    DeploySvc -->|Provisions Pods| K8sCluster
    DeploySvc -->|Encrypted Storage| Vault
    MonitorSvc -->|Polls Health| K8sCluster

    AuthSvc --- Postgres
    CatalogSvc --- Postgres
    BillingSvc --- Postgres
    DeploySvc --- Postgres
    MonitorSvc --- Postgres
```

### Technology Breakdown

| Component | Tech Stack | Location | Role |
| :--- | :--- | :--- | :--- |
| **Microservices** | Go 1.23+, Clean Architecture | `microservices/` | Decoupled domain services (`auth`, `catalog`, `billing`, `deploy`, `monitor`) |
| **Shared Library** | Go 1.23+ Module | `shared/` | Canonical events, messaging, database, vault client, error handling |
| **Manager Bot** | Node.js 20+, TypeScript, Discord.js v14 | `bots/manager-bot/` | Central Discord slash command interface with interactive select menus |
| **Frontend** | Next.js 16.3.6 (Turbopack), React 19, Tailwind CSS | `frontend/` | Web dashboard, subscription store, checkout, fleet viewer, admin panel |
| **Databases** | PostgreSQL 16 Alpine | `scripts/init-databases.sql` | 5 isolated databases (`auth_db`, `catalog_db`, `billing_db`, `deploy_db`, `monitor_db`) |
| **Broker** | RabbitMQ 3.13 (Management) | AMQP :5672, UI :15672 | Asynchronous choreography via topic exchange `discord.events` |
| **Vault** | HashiCorp Vault 1.16 | KV v2 :8200 | Sensitive Discord bot tokens stored under `secret/data/bots/{bot_id}` |
| **Orchestration**| Kubernetes / Docker Compose | `docker-compose.yml`, `k8s/` | Pod isolation, lifecycle management, PVC storage |

---

## 2. Core Architectural Invariants

Whenever generating or modifying code, you MUST preserve the following invariants:

### A. Database-per-Service Isolation
- Each microservice owns its private PostgreSQL database (`auth_db`, `catalog_db`, `billing_db`, `deploy_db`, `monitor_db`).
- **NO CROSS-DATABASE SQL QUERIES OR FOREIGN KEYS**.
- Cross-service data sharing must occur via synchronous REST APIs or asynchronous RabbitMQ events.

### B. Clean Architecture in Go Services
Every Go microservice under `microservices/<service-name>` follows strict Clean Architecture:
```text
microservices/<service-name>/
├── cmd/api/main.go          # Application bootstrap & dependency injection
└── internal/
    ├── core/
    │   ├── domain/          # Pure entities, domain models, zero external dependencies
    │   └── ports/           # Interface definitions (repositories, handlers, services)
    ├── adapters/
    │   ├── handlers/http/   # REST API controllers & router registration
    │   └── repository/      # SQL database persistence (database/sql, lib/pq)
    └── config/              # Environment variable loading & validation
```

### C. Multi-Subscription & Fleet Isolation
A single Discord Guild can hold **multiple active subscriptions** (e.g. 2 Music bots, 1 Moderation bot, 1 RPG bot).
1. **Instance Labels**: All subscription, deployment, and monitoring records require an `instance_label` (defaults to `"Default"`, `"Lobby Music"`, `"VIP Room"`, etc.).
2. **K8s Deployment Naming**: Pod names MUST avoid collisions across multiple instances within the same guild.  
   Format: `bot-{guildId}-{depShortId}` (e.g., `bot-1122334455-a1b2c3d4`).
3. **Monitoring Targets**: `monitor_db.monitoring_targets` uses `bot_id VARCHAR(64) UNIQUE` (keyed to the deployment ID) and stores `instance_label` to distinguish alerts.
4. **Interactive Disambiguation**:
   - **Discord Manager Bot**: When multiple bots exist in a guild, `/status`, `/restart`, `/bot name`, and `/bot avatar` MUST present an interactive Discord `StringSelectMenuBuilder` dropdown for user selection.
   - **Web Dashboard**: Displays tabbed fleet switchers allowing users to inspect metrics and status for each bot individually.

### D. Zero-Setup Turnkey Delivery & Token Pools
1. Customers who do not want to create Discord developer bots can choose **Zero-Setup Turnkey Delivery**.
2. Pre-warmed tokens are stored in `deploy_db.token_pool` in status `available`.
3. When purchased, `deploy-svc` transitions the token to `assigned`, fetches the token from HashiCorp Vault (`secret/data/bots/pool/{pool_id}`), and spins up the bot pod immediately.
4. Users can customize their turnkey bot's persona (username & avatar) via:
   - Manager Bot commands: `/bot name <new_name>` and `/bot avatar <url>`
   - REST API: `POST /api/v1/deployments/{id}/customize`
   - Web Dashboard: Persona customization panel.

---

## 3. Mandatory Documentation Synchronization Policy

> [!IMPORTANT]
> **DOCUMENTATION MAINTENANCE IS MANDATORY FOR ALL AGENTS.**
> Any agent modifying, refactoring, or extending platform functionality MUST keep `./docs/` synchronized with code changes in the same turn or session.

### Agent Documentation Checklist

Whenever you implement or alter:
1. **A REST API endpoint or parameter**: Update [MICROSERVICES.md](file:///C:/Users/kasep/Desktop/discord-subscriptions/docs/MICROSERVICES.md) and [DEVELOPMENT_GUIDE.md](file:///C:/Users/kasep/Desktop/discord-subscriptions/docs/DEVELOPMENT_GUIDE.md).
2. **A Database table, column, constraint, or seed**: Update [DATABASE_SCHEMAS.md](file:///C:/Users/kasep/Desktop/discord-subscriptions/docs/DATABASE_SCHEMAS.md) and the SQL script `scripts/init-databases.sql`.
3. **A RabbitMQ Event, topic, or payload struct**: Update [EVENT_SPECIFICATION.md](file:///C:/Users/kasep/Desktop/discord-subscriptions/docs/EVENT_SPECIFICATION.md) and `shared/events/`.
4. **A Discord slash command or UI flow**: Update [MANAGER_BOT.md](file:///C:/Users/kasep/Desktop/discord-subscriptions/docs/MANAGER_BOT.md).
5. **A Frontend page, route, or component**: Update [FRONTEND.md](file:///C:/Users/kasep/Desktop/discord-subscriptions/docs/FRONTEND.md).
6. **A Token Pool or 0-Setup rule**: Update [TOKEN_POOL_ZERO_SETUP.md](file:///C:/Users/kasep/Desktop/discord-subscriptions/docs/TOKEN_POOL_ZERO_SETUP.md).
7. **A Fleet or Multi-Subscription rule**: Update [MULTI_BOT_FLEET.md](file:///C:/Users/kasep/Desktop/discord-subscriptions/docs/MULTI_BOT_FLEET.md).
8. **An architectural pattern or service link**: Update [ARCHITECTURE.md](file:///C:/Users/kasep/Desktop/discord-subscriptions/docs/ARCHITECTURE.md).

---

## 4. Verification & Testing Commands

Before concluding any development task, agents must run the following validation suite:

```bash
# 1. Run all Go tests across the entire workspace
go test -v ./shared/...
go test -v ./microservices/auth-svc/...
go test -v ./microservices/catalog-svc/...
go test -v ./microservices/billing-svc/...
go test -v ./microservices/deploy-svc/...
go test -v ./microservices/monitor-svc/...

# 2. Compile all Go service binaries
cd microservices/auth-svc && go build -o /dev/null ./cmd/api && cd ../..
cd microservices/catalog-svc && go build -o /dev/null ./cmd/api && cd ../..
cd microservices/billing-svc && go build -o /dev/null ./cmd/api && cd ../..
cd microservices/deploy-svc && go build -o /dev/null ./cmd/api && cd ../..
cd microservices/monitor-svc && go build -o /dev/null ./cmd/api && cd ../..

# 3. Build Manager Bot (TypeScript)
cd bots/manager-bot && npm run build && cd ../..

# 4. Build Frontend (Next.js Turbopack)
cd frontend && npm run build && cd ..
```

---

## 5. Skills Available for Agents

Agents should invoke the specialized project skills:
- **`discord-fleet-ops`**: Runbooks for spinning up Docker Compose, seeding databases, running migrations, executing test suites, and debugging bot deployments.
- **`docs-sync`**: Runbooks for cross-checking code changes against `./docs/` and maintaining Mermaid charts and API specifications.
- **`git-commit-ops`**: Procedures for progressive staging, Conventional Commits formatting, pre-commit validation, and GitHub repository synchronization.

---

## 6. Conventional Commits & Progressive Git Workflow

All agents committing code to this repository MUST strictly follow the Conventional Commits specification:

```text
<type>(<scope>): <subject>

<body>
```

### Commit Types
- `feat`: New feature or capability (e.g. `feat(billing): add paypal subscription checkout`)
- `fix`: Bug fix (e.g. `fix(deploy): prevent pod collision on duplicate guild deployments`)
- `docs`: Documentation updates (e.g. `docs: add comprehensive health check specification`)
- `refactor`: Refactoring without behavior change (e.g. `refactor(shared): unify deep health checks`)
- `test`: Adding or correcting tests (e.g. `test(catalog): add unit tests for subscription tier validation`)
- `build`: Dependency upgrades or build tool changes (e.g. `build: upgrade next.js to 16.3.6 LTS`)
- `chore`: Tooling, scaffolding, git configuration (e.g. `chore: initialize repository scaffolding`)

### Progressive Staging Invariant
- **DO NOT** create monolithic commits spanning unrelated architectural boundaries.
- Stage and commit progressively by architectural layer:
  1. Base tooling & shared libraries
  2. Individual microservices
  3. Bots and frontend clients
  4. Infrastructure & orchestrators (Traefik, K8s)
  5. Documentation & agent configurations
- Run full pre-commit verification (`go test`, builds, `docker compose config`, `kubectl kustomize`) before committing.

