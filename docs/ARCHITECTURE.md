# System Architecture & Technical Design

This document details the architectural topology, communication paradigms, security boundary, and orchestration mechanics powering the **Discord Bot Subscription Platform**.

---

## 1. High-Level Architecture Topology

The platform operates as a distributed system designed for high availability, multi-tenant security, and horizontal scalability.

```mermaid
flowchart TD
    subgraph Users ["End Users & Discord Servers"]
        GuildAdmin["Server Owner / Admin"]
        DiscordGuild["Discord Guild / Community"]
    end

    subgraph Presentation ["Presentation Layer"]
        WebUI["Next.js 16 Web Dashboard<br/>React 19 / Turbopack"]
        MgrBot["Central Discord Manager Bot<br/>Discord.js v14 Gateway"]
    end

    subgraph GatewayLayer ["API Gateway & Routing"]
        AuthProxy["Auth & Routing Gateway"]
    end

    subgraph Services ["Core Microservices (Go Clean Architecture)"]
        AuthSvc["auth-svc<br/>Port: 8080<br/>OAuth2 / JWT"]
        CatalogSvc["catalog-svc<br/>Port: 8081<br/>Plans & Pricing"]
        BillingSvc["billing-svc<br/>Port: 8082<br/>Subscriptions & Vouchers"]
        DeploySvc["deploy-svc<br/>Port: 8083<br/>K8s & Token Pool"]
        MonitorSvc["monitor-svc<br/>Port: 8084<br/>Telemetry & Watchdog"]
    end

    subgraph EventLayer ["Asynchronous Event Bus"]
        RabbitMQ[["RabbitMQ Topic Exchange<br/>discord.events"]]
    end

    subgraph Storage ["Databases (PostgreSQL 16)"]
        DBAuth[("auth_db")]
        DBCatalog[("catalog_db")]
        DBBilling[("billing_db")]
        DBDeploy[("deploy_db")]
        DBMonitor[("monitor_db")]
    end

    subgraph SecurityVault ["Secrets Security Perimeter"]
        Vault[("HashiCorp Vault 1.16<br/>KV v2 Engine")]
    end

    subgraph ContainerCompute ["Kubernetes Cluster / Bot Fleet"]
        Pod1["Pod: bot-11223344-a1b2<br/>Music Bot - Dedicated"]
        Pod2["Pod: bot-11223344-c3d4<br/>Music Bot - VIP Room"]
        Pod3["Pod: bot-99887766-e5f6<br/>Mod Bot - Defense"]
    end

    GuildAdmin -->|HTTPS Browser| WebUI
    DiscordGuild -->|Slash Commands| MgrBot
    WebUI --> GatewayLayer
    MgrBot --> GatewayLayer

    GatewayLayer --> AuthSvc & CatalogSvc & BillingSvc & DeploySvc & MonitorSvc

    AuthSvc --- DBAuth
    CatalogSvc --- DBCatalog
    BillingSvc --- DBBilling
    DeploySvc --- DBDeploy
    MonitorSvc --- DBMonitor

    BillingSvc -->|Publish Events| RabbitMQ
    RabbitMQ -->|Consume Events| DeploySvc & MonitorSvc

    DeploySvc -->|Injects Secrets| Vault
    DeploySvc -->|Orchestrates Deployments| ContainerCompute
    MonitorSvc -->|Polls Healthz| ContainerCompute
```

---

## 2. Microservice Layer & Clean Architecture

Each microservice is authored in Go 1.26+ and strictly enforces the **Ports and Adapters (Hexagonal / Clean Architecture)** design pattern. This decouples business logic from external frameworks, database drivers, and messaging transports.

```mermaid
classDiagram
    direction TB
    namespace CoreDomain {
        class Subscription {
            +string ID
            +string UserID
            +string GuildID
            +string PlanID
            +string BotType
            +string InstanceLabel
            +string Status
            +time ValidUntil
            +bool IsZeroSetup
            +IsValid() bool
            +IsActive() bool
        }
    }

    namespace Ports {
        class SubscriptionRepository {
            <<interface>>
            +Create(ctx, sub) error
            +GetByID(ctx, id) (*Subscription, error)
            +GetByGuildID(ctx, guildId) ([]*Subscription, error)
            +Update(ctx, sub) error
        }
        class EventPublisher {
            <<interface>>
            +PublishSubscriptionCreated(ctx, event) error
        }
    }

    namespace Adapters {
        class PostgresSubscriptionRepository {
            -sql.DB db
            +Create(ctx, sub) error
            +GetByID(ctx, id) (*Subscription, error)
        }
        class RabbitMQEventPublisher {
            -amqp.Channel channel
            +PublishSubscriptionCreated(ctx, event) error
        }
        class HTTPHandler {
            -SubscriptionRepository repo
            -EventPublisher publisher
            +CreateSubscription(w, r)
            +GetSubscription(w, r)
        }
    }

    SubscriptionRepository <|.. PostgresSubscriptionRepository : implements
    EventPublisher <|.. RabbitMQEventPublisher : implements
    HTTPHandler --> SubscriptionRepository : uses
    HTTPHandler --> EventPublisher : uses
    SubscriptionRepository --> Subscription : references
```

### Layer Responsibilities
1. **Core Domain (`internal/core/domain`)**:
   - Represents the enterprise business rules and entities.
   - Completely free of external package dependencies (no `net/http`, no `database/sql`, no ORMs).
2. **Ports (`internal/core/ports`)**:
   - Interfaces defining inbound operations (services/use cases) and outbound operations (repositories, event publishers, external APIs).
3. **Adapters (`internal/adapters`)**:
   - Primary (Driving) Adapters: REST HTTP handlers decoding JSON requests and invoking domain use cases.
   - Secondary (Driven) Adapters: PostgreSQL database repositories executing parameterized SQL, and RabbitMQ publishers/subscribers.

---

## 3. Communication Patterns: Sync vs Async

| Interaction | Protocol | Mechanism | Rationale |
| :--- | :--- | :--- | :--- |
| **User Navigation & Dashboard Queries** | Synchronous REST | HTTP / JSON | Immediate feedback required for UI rendering, catalog browsing, and status queries. |
| **Manager Bot Commands** | Synchronous REST | HTTP / JSON | Discord slash command interactions have a 3-second reply window requiring direct responses. |
| **Checkout & Subscription Lifecycle** | Asynchronous Event-Driven | AMQP over RabbitMQ | Guarantees eventual consistency, decoupling payment processing from Kubernetes pod scheduling. |
| **Watchdog Health Notifications** | Asynchronous Event-Driven | AMQP over RabbitMQ | Alerts and restart requests are dispatched asynchronously without blocking polling loops. |

---

## 4. End-to-End Subscription & Deployment Sequence

The sequence diagram below illustrates the exact lifecycle from user checkout to running Kubernetes bot pod:

```mermaid
sequenceDiagram
    autonumber
    actor User as Server Owner
    participant Web as Web Dashboard
    participant Billing as billing-svc
    participant Broker as RabbitMQ (discord.events)
    participant Deploy as deploy-svc
    participant Vault as HashiCorp Vault
    participant K8s as Kubernetes Cluster
    participant Monitor as monitor-svc

    User->>Web: Selects Plan, Enters Guild ID & "VIP Room" Label
    User->>Web: Confirms 0-Setup Turnkey Delivery
    Web->>Billing: POST /api/v1/subscriptions
    Billing->>Billing: Verify Promo / Voucher & Record Subscription
    Billing->>Broker: Publish "subscription.created" Event
    Billing-->>Web: HTTP 201 Created (Subscription Activated)

    Broker->>Deploy: Consume "subscription.created"
    Deploy->>Deploy: Check is_zero_setup flag (true)
    Deploy->>Deploy: Reserve pre-warmed token from token_pool
    Deploy->>Vault: Fetch Bot Token & Client Secret
    Deploy->>K8s: Create K8s Deployment & Service<br/>Pod Name: bot-11223344-a1b2c3d4
    K8s-->>Deploy: Pod Status: Running
    Deploy->>Broker: Publish "deployment.completed" Event

    Broker->>Monitor: Consume "deployment.completed"
    Monitor->>Monitor: Register Monitoring Target (instance_label: "VIP Room")
    Monitor->>K8s: Poll GET /healthz every 15s
    K8s-->>Monitor: HTTP 200 OK (latency: 12ms)
```

---

## 5. Security Perimeter & HashiCorp Vault Integration

Discord Bot Tokens grant privileged access to Discord servers and must be defended against database breaches and code exposure:

```mermaid
flowchart LR
    subgraph PublicBoundary [Public / DMZ]
        DashboardClient["Next.js Web Browser"]
    end

    subgraph AppBoundary [Microservice Internal Network]
        DeploySvc["deploy-svc"]
        BillingSvc["billing-svc"]
    end

    subgraph SafeBoundary [Isolated Vault Perimeter]
        VaultServer["HashiCorp Vault 1.16\n(AppRole / Token Auth)"]
        VaultKV["KV v2 Storage Engine\nsecret/data/bots/{bot_id}"]
    end

    subgraph PodExecution [Kubernetes Workloads]
        BotContainer["Discord Bot Container\n(Environment Variable Injected)"]
    end

    DashboardClient -.->|Never Sees Plain Token| DeploySvc
    DeploySvc -->|HTTPS /vault/v1/data| VaultServer
    VaultServer --- VaultKV
    DeploySvc -->|Injects via Secret Volume / Env| BotContainer
```

### Security Rules:
1. **Never Stored in PostgreSQL**: Discord bot tokens and OAuth client secrets are strictly forbidden from being stored in plaintext in any PostgreSQL database.
2. **Path Convention**:
   - Turnkey Pool Tokens: `secret/data/bots/pool/{pool_id}`
   - Customer-Owned Tokens: `secret/data/bots/{deployment_id}`
3. **Quarantine Mechanism**: Compromised or corrupted tokens are flagged as `quarantined` in `deploy_db.token_pool` and locked from reuse.
