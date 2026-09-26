---
name: git-commit-ops
description: >-
  Use this skill to execute git operations, stage changes by logical architectural units, format Conventional Commit messages, verify clean test & build status before committing, and manage GitHub remotes.
---

# Git Commit Operations & Conventional Commits Skill

This skill provides step-by-step procedures for preparing, staging, validating, and committing changes to the **Discord Bot Subscription Platform** git repository following the **Conventional Commits** standard and progressive staging invariants.

---

## 1. When to Activate This Skill

Activate this skill whenever:
- Initializing or structuring git version control for the platform
- Committing a completed feature, bugfix, documentation update, or refactoring
- Splitting a large changeset into progressive, atomic commits
- Creating or managing GitHub repositories using `gh` CLI
- Verifying staging cleanliness before concluding an agent session

---

## 2. Pre-Commit Validation Checklist

Before staging any commit, agents MUST run and verify the validation suite:

```powershell
# 1. Run all Go tests
go test -v ./shared/... ./microservices/...

# 2. Build all Go microservices
go build ./shared/... ./microservices/auth-svc/... ./microservices/catalog-svc/... ./microservices/billing-svc/... ./microservices/deploy-svc/... ./microservices/monitor-svc/...

# 3. Build Manager Bot (TypeScript)
npm --prefix bots/manager-bot run build

# 4. Build Frontend (Next.js 16)
npm --prefix frontend run build

# 5. Validate Docker Compose syntax
docker compose config --quiet

# 6. Validate Kubernetes manifests
kubectl kustomize k8s/
```

---

## 3. Progressive Staging Strategy

Never execute a blanket `git add .` that mixes disparate domain layers. Group changes logically:

1. **Scaffolding & Tooling**: Base configurations, `.gitignore`, `.env.example`, `go.work`, `Makefile`, DB init scripts.
2. **Shared Libraries**: Canonical events, messaging clients, Vault integration, health check package.
3. **Core Microservices**: Specific domain microservice changes (`catalog-svc`, `auth-svc`, `billing-svc`, `deploy-svc`, `monitor-svc`).
4. **Bots & Clients**: Manager Bot (`bots/manager-bot/`).
5. **Frontend Web UI**: Next.js dashboard, checkout, components (`frontend/`).
6. **Infrastructure & Gateway**: `docker-compose.yml`, Traefik reverse proxy configs (`traefik/`).
7. **Cloud Orchestration**: Kubernetes manifests (`k8s/`).
8. **Documentation & Agent Configs**: `./docs/`, `.agent/`, `AGENTS.md`, `GEMINI.md`, `CLAUDE.md`.

---

## 4. Conventional Commit Template

```text
<type>(<scope>): <short imperative description>

<detailed body explaining motivation, architectural choices, and context>

[optional footer, e.g. Closes #123]
```

### Types & Examples:
- `feat(billing): add paypal subscription checkout and voucher redemption`
- `fix(deploy): prevent container name collision in multi-bot guild fleet`
- `docs: add health checks and observability specification`
- `chore: initialize repository scaffolding and multi-workspace go tooling`
- `refactor(shared): unify deep health checks across all microservices`
