---
name: discord-fleet-ops
description: >-
  Use this skill when developing, testing, deploying, or troubleshooting the Discord Bot Subscription Platform microservices, manager bot, or frontend. Provides step-by-step procedures for local environment startup, Go multi-workspace testing, TypeScript builds, database migrations, Vault token inspection, and fleet debugging.
---

# Discord Fleet Operations & Developer Skill

This skill provides operational procedures, build commands, and debugging runbooks for the **Discord Bot Subscription & Turnkey Fleet Management Platform**.

---

## 1. Environment & Service Topology

| Component | Port | Technology | Database / Backend |
| :--- | :--- | :--- | :--- |
| **auth-svc** | `:8080` | Go 1.23 REST API | `auth_db` |
| **catalog-svc** | `:8081` | Go 1.23 REST API | `catalog_db` |
| **billing-svc** | `:8082` | Go 1.23 REST API | `billing_db` + RabbitMQ |
| **deploy-svc** | `:8083` | Go 1.23 REST API | `deploy_db` + Vault + K8s |
| **monitor-svc** | `:8084` | Go 1.23 REST API | `monitor_db` + RabbitMQ |
| **manager-bot** | Gateway | TypeScript (Discord.js v14) | REST Client |
| **frontend** | `:3000` | Next.js 16.3.6 (Turbopack) | REST Client |
| **PostgreSQL** | `:5432` | PostgreSQL 16 Alpine | 5 isolated databases |
| **RabbitMQ** | `:5672` (AMQP), `:15672` (UI) | RabbitMQ 3.13 Alpine | Topic: `discord.events` |
| **HashiCorp Vault** | `:8200` | Vault 1.16 | KV v2 engine |

---

## 2. Fast Build & Test Procedure

Run this standard command sequence to verify the integrity of the entire codebase:

### Go Microservices & Shared Library
```powershell
# In repository root
go test -v ./shared/...
go test -v ./microservices/auth-svc/...
go test -v ./microservices/catalog-svc/...
go test -v ./microservices/billing-svc/...
go test -v ./microservices/deploy-svc/...
go test -v ./microservices/monitor-svc/...
```

### TypeScript Manager Bot Build
```powershell
# Compile TypeScript to dist/
cd bots/manager-bot
npm run build
cd ../..
```

### Next.js Frontend Build
```powershell
# Compile Next.js 16 app with Turbopack
cd frontend
npm run build
cd ..
```

---

## 3. Docker Compose Local Infrastructure

To spin up the entire platform locally using Docker Compose:

```powershell
# Start all containers in the background
docker compose up -d

# Check health of database, broker, and vault
docker compose ps

# Tail logs of a specific service
docker compose logs -f deploy-svc
docker compose logs -f billing-svc
docker compose logs -f manager-bot
```

### Resetting Databases
If schemas in `scripts/init-databases.sql` are updated and you need to wipe and re-seed the local PostgreSQL volume:
```powershell
docker compose down -v
docker compose up -d postgres
```

---

## 4. HashiCorp Vault Bot Token Administration

All sensitive bot tokens are stored in Vault under `secret/data/bots/{bot_id}` or `secret/data/bots/pool/{pool_id}`.

### Inspecting Pre-Warmed Turnkey Token Pool
```bash
# Check status in deploy_db
docker exec -it discord_subscriptions_postgres psql -U postgres -d deploy_db -c "SELECT id, bot_type, status, assigned_guild_id FROM token_pool;"
```

### Adding New Managed Bot Tokens to Pool via API
```bash
curl -X POST http://localhost:8083/api/v1/token-pool/tokens \
  -H "Content-Type: application/json" \
  -d '{
    "bot_type": "music",
    "client_id": "131234567890123499",
    "discord_token": "YOUR_DISCORD_BOT_TOKEN_HERE"
  }'
```

---

## 5. Multi-Subscription Fleet Verification

To verify that multiple bots in the same guild do not collide:

1. **Verify Pod Naming**: Inspect `deploy_db.deployments` to confirm the format `bot-{guildId}-{depShortId}`:
   ```sql
   SELECT id, guild_id, bot_type, instance_label, k8s_deployment_name, status FROM deployments;
   ```
2. **Verify Monitoring Separation**: Inspect `monitor_db.monitoring_targets`:
   ```sql
   SELECT id, bot_id, guild_id, instance_label, current_status, last_checked_at FROM monitoring_targets;
   ```
3. **Verify Manager Bot Disambiguation**:
   In Discord, trigger `/status` or `/restart` in a guild containing 2 or more bots. The bot will reply with a dropdown (`StringSelectMenuBuilder`) listing all instances by label and status.
