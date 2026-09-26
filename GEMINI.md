# GEMINI.md - Antigravity Agent Directive

This file configures Antigravity CLI and Gemini pairing agents for this repository.

## Project Context
- **Name**: Discord Bot Subscription Platform
- **Root Directory**: `C:\Users\kasep\Desktop\discord-subscriptions`
- **Architecture**: Go Clean Architecture microservices (`auth-svc`, `catalog-svc`, `billing-svc`, `deploy-svc`, `monitor-svc`), PostgreSQL 16 database-per-service, RabbitMQ event bus, HashiCorp Vault secrets, Next.js 16 (Turbopack) frontend, Discord.js v14 TypeScript Manager Bot.
- **Core Guidelines**: See [AGENTS.md](file:///C:/Users/kasep/Desktop/discord-subscriptions/AGENTS.md).

## Critical Directives
1. **Mandatory Documentation Sync**: Every change to an API, schema, event, bot command, or UI page MUST immediately be reflected in `./docs/` as defined in `AGENTS.md` and `.agent/rules/documentation-policy.md`.
2. **Clean Architecture Integrity**: In Go services, maintain the strict separation between `core/domain`, `core/ports`, and `adapters/`. Domain entities MUST NEVER import database drivers, HTTP frameworks, or external libraries.
3. **Multi-Subscription Invariant**: All subscriptions, deployments, and monitoring targets MUST preserve `instance_label` and use unique pod identifiers (`bot-{guildId}-{depShortId}`) to avoid fleet collisions.
4. **Zero-Setup Turnkey Safety**: User tokens and managed tokens MUST NEVER be stored in plain text in PostgreSQL. Always use HashiCorp Vault KV v2.
5. **Always Verify Builds**: Run tests across `go.work`, build TypeScript manager-bot (`npm run build`), and build Next.js frontend (`npm run build`) after code edits.
6. **Conventional Commits & Progressive Staging**: Format all commit messages according to `.agent/rules/git-workflow.md` (`feat`, `fix`, `docs`, `refactor`, etc.). Stage and commit changes progressively by architectural layer rather than creating monolithic commits.
