# CLAUDE.md

See [AGENTS.md](./AGENTS.md) for complete architecture specifications, directory maps, verification commands, and invariants.

## Key Rules
- Microservices follow Go Clean Architecture (`core/domain`, `core/ports`, `adapters/`).
- Database per service with PostgreSQL 16.
- Asynchronous choreography uses RabbitMQ topic exchange `discord.events`.
- Secrets stored in HashiCorp Vault under `secret/data/bots/{bot_id}`.
- Multi-subscription fleet isolation: `instance_label` + `bot-{guildId}-{depShortId}`.
- Mandatory rule: Always synchronize changes with `./docs/` documentation files!
- Git commits: Always follow Conventional Commits (`feat`, `fix`, `docs`, etc.) and progressive staging per `.agent/rules/git-workflow.md`.
