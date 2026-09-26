# Git Workflow & Conventional Commits Policy

> **Scope**: All AI coding agents and human contributors committing changes to this repository MUST strictly follow the Conventional Commits specification and the Progressive Commits Protocol.

---

## 1. Conventional Commits Standard

All commit messages MUST adhere to the following format:

```text
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

### 1.1 Allowed Types

| Type | Description | Example |
| :--- | :--- | :--- |
| `feat` | A new feature or capability | `feat(billing): add paypal recurring subscription webhook handler` |
| `fix` | A bug fix | `fix(deploy): prevent pod collision on duplicate guild deployments` |
| `docs` | Documentation-only changes | `docs: add comprehensive health check specification` |
| `refactor` | Code change that neither fixes a bug nor adds a feature | `refactor(shared): extract rabbitmq reconnect logic into shared client` |
| `perf` | Code change that improves performance | `perf(frontend): optimize fleet dashboard bot card rendering` |
| `test` | Adding missing tests or correcting existing tests | `test(catalog): add unit tests for subscription tier validation` |
| `build` | Changes that affect build system, docker, or external dependencies | `build: upgrade next.js to 16.3.6 LTS` |
| `ci` | Changes to CI/CD workflows or scripts | `ci: add github actions workflow for multi-service tests` |
| `chore` | Maintenance tasks, configs, gitignore, or tooling | `chore: configure traefik reverse proxy and unified ingress` |
| `revert` | Reverts a previous commit | `revert: revert feat(billing) voucher discount calculation` |

### 1.2 Recognized Scopes

- `auth`: `microservices/auth-svc/`
- `catalog`: `microservices/catalog-svc/`
- `billing`: `microservices/billing-svc/`
- `deploy`: `microservices/deploy-svc/`
- `monitor`: `microservices/monitor-svc/`
- `shared`: `shared/`
- `manager-bot`: `bots/manager-bot/`
- `frontend`: `frontend/`
- `gateway`: `traefik/`
- `k8s`: `k8s/`
- `infra`: `docker-compose.yml`, databases, Vault, RabbitMQ
- `docs`: `./docs/`
- `agents`: `.agent/`, `AGENTS.md`, `GEMINI.md`, `CLAUDE.md`, `.cursorrules`

---

## 2. Progressive Commit Invariants

1. **Atomic & Progressive Staging**:
   - DO NOT stage all changed files simultaneously using `git add .` if changes span multiple decoupled architectural layers or features.
   - Stage and commit files by domain layer (e.g. `shared/` first, followed by microservice core, then bots/frontend, then infrastructure/k8s, then documentation).
2. **Never Commit Broken Code**:
   - Every commit in the repository history must compile cleanly (`go test`, `go build`, `npm run build`).
3. **No Unversioned Bloat**:
   - Never commit `.env`, `node_modules/`, `.next/`, `dist/`, binary `.exe` files, or temporary test artifacts. Always verify `git status` against `.gitignore`.
4. **Synchronized Documentation**:
   - Accompanying documentation changes in `./docs/` should either be included in the feature commit or in a distinct `docs:` commit immediately following it.
