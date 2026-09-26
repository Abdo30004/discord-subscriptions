---
name: docs-sync
description: >-
  Use this skill whenever any feature, API endpoint, database schema, event type, bot command, or UI page is added or modified in the Discord Bot Subscription Platform. Enforces automatic synchronization and updating of ./docs/ documentation files, Mermaid diagrams, API references, and schemas.
---

# Documentation Synchronization Skill

This skill ensures that all changes to the Discord Bot Subscription Platform codebase are accurately documented in the `./docs/` directory, keeping system specifications, architecture diagrams, and API references 100% up to date.

---

## 1. When to Activate This Skill

Activate this skill immediately after implementing any of the following:
- Adding or modifying a Go HTTP handler (`microservices/*/internal/adapters/handlers/http/`)
- Adding or modifying database tables, columns, or seeds (`scripts/init-databases.sql`)
- Adding or modifying domain events or RabbitMQ bindings (`shared/events/`)
- Adding or modifying Discord Slash Commands or UI components (`bots/manager-bot/src/commands/`)
- Adding or modifying Next.js routes, components, or client API callers (`frontend/src/`)
- Modifying Docker Compose service ports, volumes, or environment variables (`docker-compose.yml`)

---

## 2. Step-by-Step Documentation Update Protocol

### Step 1: Detect Code Delta
Inspect git diff or recent code modifications to list changed components:
1. Changed endpoints (e.g. `POST /api/v1/deployments/{id}/customize`)
2. Changed models/tables (e.g. `instance_label` column added to `subscriptions`, `deployments`, `monitoring_targets`)
3. Changed events (e.g. `DeploymentCustomizedEvent`)
4. Changed slash commands (e.g. `/bot avatar <url>`)

### Step 2: Cross-Reference Documentation Map

| Area | Target File | Action Required |
| :--- | :--- | :--- |
| **System Flow / Topology** | `docs/ARCHITECTURE.md` | Update Mermaid flowchart or sequence diagrams |
| **Microservice APIs** | `docs/MICROSERVICES.md` | Add/update endpoint table and JSON request/response schema |
| **Fleet / Multi-Tenancy** | `docs/MULTI_BOT_FLEET.md` | Update pod naming conventions or multi-bot UI flows |
| **Zero-Setup & Token Pool** | `docs/TOKEN_POOL_ZERO_SETUP.md` | Update token state machine or Vault secret paths |
| **PostgreSQL Databases** | `docs/DATABASE_SCHEMAS.md` | Update Mermaid `erDiagram` and table columns |
| **Event Choreography** | `docs/EVENT_SPECIFICATION.md` | Update routing keys and JSON event definitions |
| **Discord Manager Bot** | `docs/MANAGER_BOT.md` | Update slash command table and select menu interactions |
| **Frontend Web App** | `docs/FRONTEND.md` | Update route table, component props, and API calls |
| **Local Runbook / Envs** | `docs/DEVELOPMENT_GUIDE.md` | Update environment variables and test commands |

### Step 3: Mermaid Diagram Update Guidelines

When updating Mermaid diagrams:
- **Flowcharts**: Use `flowchart TD` or `flowchart LR`.
- **Sequence Diagrams**: Use `sequenceDiagram` with `autonumber` and clear participant labels.
- **Entity Relationship Diagrams**: Use `erDiagram`. Ensure cardinality (`||--o{`) and attributes match PostgreSQL schemas.
- **State Diagrams**: Use `stateDiagram-v2` for token or subscription status transitions.

### Step 4: Verification & Link Check
- Verify all relative file links in markdown are valid.
- Verify table headers and pipe separators align cleanly.
- Verify code blocks use correct syntax highlighters (`go`, `typescript`, `tsx`, `json`, `sql`, `mermaid`, `bash`).
