# Discord Bot Subscription & Turnkey Fleet Platform Documentation

Welcome to the comprehensive technical documentation for the **Discord Bot Subscription & Turnkey Fleet Management Platform**.

This documentation suite covers every layer of the platform: microservices, multi-subscription fleet tenancy, zero-setup turnkey bot provisioning, database schemas, event choreography, manager bot commands, frontend dashboard, and development operations.

---

## 🗺️ Documentation Directory

| Document | Description | Key Focus Areas |
| :--- | :--- | :--- |
| [**Architecture Overview**](./ARCHITECTURE.md) | High-level system design, topology, service boundaries | Clean Architecture, Sync vs Async, Vault boundary, K8s engine |
| [**Microservices Reference**](./MICROSERVICES.md) | Detailed specification of all 5 Go microservices | `auth-svc`, `catalog-svc`, `billing-svc`, `deploy-svc`, `monitor-svc`, REST APIs |
| [**Multi-Bot Fleet Tenancy**](./MULTI_BOT_FLEET.md) | Running multiple bots per Discord Guild | Pod collision prevention, instance labels, Discord select menus, dashboard tabs |
| [**Token Pool & 0-Setup Engine**](./TOKEN_POOL_ZERO_SETUP.md) | Turnkey 1-click deployment without developer bots | Pre-warmed token pool, Vault secret paths, persona customization |
| [**Database Schemas & ERD**](./DATABASE_SCHEMAS.md) | Database-per-service PostgreSQL relational models | Full Mermaid ER diagrams, table schemas, indexes, constraints |
| [**Event-Driven Specification**](./EVENT_SPECIFICATION.md) | RabbitMQ choreography & messaging protocol | Topic exchange `discord.events`, routing keys, JSON schemas, DLX |
| [**Manager Bot Manual**](./MANAGER_BOT.md) | Central Discord.js v14 slash command bot | `/status`, `/restart`, `/bot`, `/store`, `/redeem`, interactive dropdowns |
| [**Frontend Web Application**](./FRONTEND.md) | Next.js 16 Turbopack Web Dashboard & Store | `/store`, `/checkout`, `/dashboard`, `/admin`, `/setup`, state management |
| [**Kubernetes Production Guide**](./KUBERNETES.md) | Cluster manifests, namespaces, RBAC, and network isolation | `platform` & `discord-bots` namespaces, `deploy-svc-sa`, NetworkPolicy |
| [**Traefik Reverse Proxy Guide**](./REVERSE_PROXY_TRAEFIK.md) | Unified edge gateway, path routing, CORS, and rate limiting | Traefik v3, `:80` unified entrypoint, dashboard `:8090`, IngressRoute |
| [**Health Checks & Observability**](./HEALTH_CHECKS.md) | Standardized deep/shallow probes, metrics & readiness | `/health`, `/livez`, `/readyz`, Manager Bot `:8085`, Frontend, K8s probes |
| [**CI/CD & Changelog System**](./CI_CD_AND_CHANGELOG.md) | GitHub Actions CI/CD workflows and automated changelogs | `ci.yml`, `release.yml`, GHCR containers, `generate-changelog.mjs`, `cliff.toml` |
| [**Development & Runbook Guide**](./DEVELOPMENT_GUIDE.md) | Local environment, tests, builds, and doc sync | Docker Compose, `go.work`, TypeScript compilation, **Doc Sync Policy** |

---

## 🚀 System Architecture at a Glance

```mermaid
flowchart TB
    subgraph Clients [Clients & Interaction Channels]
        Browser["Web Dashboard & Store\n(Next.js 16 Turbopack :3000)"]
        DiscordClient["Discord Client / Guild Members\n(Discord Gateway)"]
    end

    subgraph Edge [Edge & Interfaces]
        ManagerBot["Discord Manager Bot\n(Node.js / Discord.js v14)"]
        APIEndpoints["Internal REST APIs\n(:8080 - :8084)"]
    end

    subgraph Services [Golang Microservices (Clean Architecture)]
        AuthSvc["auth-svc :8080\nOAuth2 & JWT Sessions"]
        CatalogSvc["catalog-svc :8081\nBot Templates & Plans"]
        BillingSvc["billing-svc :8082\nSubscriptions, Promos, Vouchers"]
        DeploySvc["deploy-svc :8083\nK8s Engine & Token Pool"]
        MonitorSvc["monitor-svc :8084\nHealth Poller & Telemetry"]
    end

    subgraph Infrastructure [Data, Security & Messaging]
        Postgres[("PostgreSQL 16\n(auth, catalog, billing, deploy, monitor DBs)")]
        RabbitMQ[["RabbitMQ 3.13\n(Exchange: discord.events)"]]
        Vault[("HashiCorp Vault 1.16\n(Encrypted Bot Tokens)")]
        K8s[("Kubernetes Cluster\n(Isolated Bot Pods: bot-{guildId}-{depShortId})")]
    end

    Browser --> APIEndpoints
    DiscordClient --> ManagerBot
    ManagerBot --> APIEndpoints

    APIEndpoints --> AuthSvc & CatalogSvc & BillingSvc & DeploySvc & MonitorSvc

    AuthSvc --- Postgres
    CatalogSvc --- Postgres
    BillingSvc --- Postgres
    DeploySvc --- Postgres
    MonitorSvc --- Postgres

    BillingSvc -->|subscription.created| RabbitMQ
    RabbitMQ -->|deployment.requested| DeploySvc
    DeploySvc -->|Provision Pod| K8s
    DeploySvc -->|Fetch / Store Token| Vault
    RabbitMQ -->|deployment.completed| MonitorSvc
    MonitorSvc -->|Poll /healthz| K8s
```

---

## 🔑 Core Platform Principles

1. **Zero-Friction Onboarding (0-Setup)**: Customers do not need Discord Developer Portal credentials, public keys, or API tokens. Pre-warmed tokens are automatically bound upon checkout.
2. **Multi-Bot Fleet Isolation**: A guild can deploy multiple bots without port, volume, or namespace conflicts. Pods are uniquely named `bot-{guildId}-{depShortId}`.
3. **Database-per-Service**: Decoupled persistence across 5 databases (`auth_db`, `catalog_db`, `billing_db`, `deploy_db`, `monitor_db`).
4. **Hardened Secret Management**: No bot tokens or OAuth secrets are ever stored in plain text. Everything is vaulted in HashiCorp Vault KV v2.
5. **Clean Architecture**: Go services strictly separate pure domain logic from database drivers and HTTP transports.

---

## 🔄 Self-Maintaining Documentation Directive

Whenever any engineer or AI agent modifies this codebase, the changes **MUST** be recorded across the relevant documents in this directory following the guidelines in [DEVELOPMENT_GUIDE.md](./DEVELOPMENT_GUIDE.md#documentation-synchronization-protocol) and [AGENTS.md](../AGENTS.md).
