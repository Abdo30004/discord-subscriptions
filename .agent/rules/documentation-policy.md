# Documentation Synchronization Policy

> **Directive**: This rule is strictly binding on all AI agents operating in this repository.  
> Whenever any code, configuration, database table, or UI component is modified, added, or removed, the agent **MUST** update the relevant documentation in `./docs/` during the same turn.

---

## 1. Documentation Mapping Matrix

Use this matrix to identify which file(s) in `./docs/` must be updated when modifying code:

| What Changed in the Codebase | Documentation File(s) to Update | Sections to Update |
| :--- | :--- | :--- |
| **New or modified REST API endpoint** (`microservices/*/internal/adapters/handlers/http/`) | `docs/MICROSERVICES.md`<br>`docs/DEVELOPMENT_GUIDE.md` | Endpoint table, Request/Response JSON schemas, HTTP status codes |
| **Database schema, migration, seed, column** (`scripts/init-databases.sql` or repo queries) | `docs/DATABASE_SCHEMAS.md` | Mermaid ER diagrams, table column definitions, constraint lists |
| **RabbitMQ event struct or routing key** (`shared/events/*.go`) | `docs/EVENT_SPECIFICATION.md`<br>`docs/ARCHITECTURE.md` | Topic exchange routing table, JSON payload spec, sequence diagrams |
| **Discord slash command, button, or select menu** (`bots/manager-bot/src/commands/`) | `docs/MANAGER_BOT.md`<br>`docs/MULTI_BOT_FLEET.md` | Command table, Discord interaction walkthrough, options & permissions |
| **Turnkey token pool or 0-setup mechanics** (`deploy-svc`, Vault, or admin APIs) | `docs/TOKEN_POOL_ZERO_SETUP.md` | Token lifecycle state machine, Vault path conventions, API schemas |
| **Multi-subscription / fleet isolation logic** (instance labels, pod naming) | `docs/MULTI_BOT_FLEET.md` | Pod naming convention, disambiguation workflows, UI/Bot dropdown specs |
| **Frontend page, route, or checkout component** (`frontend/src/app/`, `frontend/src/lib/`) | `docs/FRONTEND.md` | Route table, Component hierarchy, state management, API client specs |
| **System architecture, topology, or infrastructure** (`docker-compose.yml`, `k8s/`) | `docs/ARCHITECTURE.md`<br>`docs/README.md` | System topology diagrams, port mappings, infrastructure inventory |
| **Environment variables, flags, secrets, or ports** (`.env.example`, `docker-compose.yml`) | `.env.example`<br>`docs/DEVELOPMENT_GUIDE.md` | Configuration reference, defaults, descriptions, and run commands |

---

## 2. Standard Update Workflow for Agents

When implementing a feature or fixing a bug:

1. **Step 1: Code Implementation & Verification**
   - Implement the necessary domain changes, repository methods, handlers, or UI components.
   - Run tests (`go test ./...` or `npm run build`) to ensure the code functions properly.

2. **Step 2: Identify Affected Documentation & Environment Configurations**
   - Check the mapping matrix above.
   - Example: If you added a new promo code validation field, you must update:
     - `docs/MICROSERVICES.md` (billing-svc endpoint table & payload)
     - `docs/DATABASE_SCHEMAS.md` (billing_db.promo_codes table columns)
     - `docs/FRONTEND.md` (checkout coupon input state)
   - If any new configuration or port was added, update `.env.example` immediately.

3. **Step 3: Edit Documentation & .env.example**
   - Modify the markdown file(s) in `./docs/`.
   - Update any Mermaid diagrams if service links, state machines, or ERDs were altered.
   - Keep `.env.example` fully synchronized with sensible development defaults.
   - Ensure markdown tables remain properly aligned.

4. **Step 4: Check Links & Consistency**
   - Verify that cross-document links between `./docs/` files remain valid.
   - Confirm that ports, environment variable names, and JSON keys match the exact code implementation.

---

## 3. Formatting Standards for `./docs/`

- **Diagrams**: All diagrams must use valid GitHub Flavored Markdown `mermaid` fenced blocks (`flowchart`, `sequenceDiagram`, `stateDiagram-v2`, `erDiagram`).
- **Endpoints**: Include HTTP Method, Path, Auth requirements, Request Body JSON, Response Body JSON, and Error codes.
- **Code Symbols**: Always format structs, fields, functions, and files in inline code blocks or clickable file links.

---

## 4. Mandatory `.env.example` Maintenance Rule

> [!IMPORTANT]
> **`.env.example` MUST BE KEPT 100% SYNCHRONIZED AT ALL TIMES.**  
> Whenever an agent adds, modifies, or retires an environment variable in any Go microservice, Manager Bot, Frontend, Docker Compose, or Traefik configuration:
> 1. Immediately edit `.env.example` in the root repository.
> 2. Organize the variable under its designated logical category (e.g. Database, RabbitMQ, Vault, Manager Bot, PayPal, Traefik).
> 3. Provide a working, safe development default or clear placeholder (e.g. `your_discord_client_id`).
> 4. Add an inline comment explaining what the variable does and valid choices.
