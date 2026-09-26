# Changelog

All notable changes to the **Discord Bot Subscription & Turnkey Fleet Management Platform** will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project strictly adheres to [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/).

---

## [v1.0.0] - 2026-09-26

### 🚀 Features & Capabilities

- **ci**: add static analysis, vulnerability scanning, and e2e integration test suite ([f85ab25](https://github.com/Abdo30004/discord-subscriptions/commit/f85ab2538f52c88e65274f0f193cfda7f246d900))
- **frontend**: convert landing and store to server components with error boundaries ([6ffaf66](https://github.com/Abdo30004/discord-subscriptions/commit/6ffaf66f5b1befb02cbebfc1f6e090afc7db25b3))
- **manager-bot**: implement discord gateway sharding with master health aggregator ([c4f7d39](https://github.com/Abdo30004/discord-subscriptions/commit/c4f7d3980c7da7309188585e11f5917c17ad5d9e))
- **auth**: add silent session refresh with vault discord token rotation ([15845ed](https://github.com/Abdo30004/discord-subscriptions/commit/15845edd5fc64ef049f65e2ea7f9a32c44e89baa))
- **services**: wire prometheus metrics endpoint and outbound circuit breaking ([bb9e096](https://github.com/Abdo30004/discord-subscriptions/commit/bb9e0962fbfd3e01535212f627e98bfa05a6402f))
- **shared**: add prometheus telemetry, w3c trace context, and circuit breaker ([df2cda6](https://github.com/Abdo30004/discord-subscriptions/commit/df2cda65ca5785a37e9ed5334a8d48b18b5a3a75))
- **services**: integrate consumer deduplication and domain idempotency guards ([62f5cf7](https://github.com/Abdo30004/discord-subscriptions/commit/62f5cf713a62fbe51fbcc9eb48e0e6c01c34f31a))
- **shared**: add rabbitmq auto-reconnect, dlx/dlq retries, event versioning and deduplicator ([af896ff](https://github.com/Abdo30004/discord-subscriptions/commit/af896ff438315ede39c9ac6e596c0b4a74fc98d7))
- **frontend**: integrate httponly cookie sessions and pass auth tokens to admin APIs ([957f10d](https://github.com/Abdo30004/discord-subscriptions/commit/957f10d345ef71b358c3d5ae1a14908f0f59f889))
- **deploy**: enforce admin authentication on token pool endpoints ([0942fa4](https://github.com/Abdo30004/discord-subscriptions/commit/0942fa479f1b07ba89dc39b8dd06297fe9389d07))
- **billing**: protect admin grant, promo, and voucher endpoints with jwt validator ([9710efd](https://github.com/Abdo30004/discord-subscriptions/commit/9710efd58b1a5955a0b30f848fbb8f2efd8be953))
- **auth**: vault-backed discord oauth tokens with httponly session cookie ([8090e3c](https://github.com/Abdo30004/discord-subscriptions/commit/8090e3cc27af5845a320767a4dfac210f32b8a9e))
- **shared**: migrate oauth tokens to vault and add jwt claims validator ([daa0776](https://github.com/Abdo30004/discord-subscriptions/commit/daa07761c2d2e3c95392fe1bb57212ae73aa3c8d))
- **deploy**: inject SUPER_ADMIN_DISCORD_IDS into auth-svc docker compose environment ([f3e5479](https://github.com/Abdo30004/discord-subscriptions/commit/f3e5479cdc151f27456da86db6640d3b17a892e7))
- **frontend**: eradicate devLogin and add staff management panel to admin dashboard ([a9d6535](https://github.com/Abdo30004/discord-subscriptions/commit/a9d6535e1b47e7ddefbcbf74ccc85219851e4e7b))
- **auth**: implement super admin hierarchy and admin promotion endpoints ([eaecf51](https://github.com/Abdo30004/discord-subscriptions/commit/eaecf51706b0d87c7f631d7daa4c58b2007eaa48))
- **auth**: add admin columns to users table and SUPER_ADMIN_DISCORD_IDS to env template ([051aa59](https://github.com/Abdo30004/discord-subscriptions/commit/051aa59ccba60ab2d720b9dc26afb4e3d49a5211))
- **bot**: configure manager-bot container to route API requests through Traefik ([e7cd82e](https://github.com/Abdo30004/discord-subscriptions/commit/e7cd82e1e0e285429cb2b237c85002ec77263edb))
- **frontend**: enforce conditional navigation and strict route authentication guards ([f698d5d](https://github.com/Abdo30004/discord-subscriptions/commit/f698d5d9236c95ef7ce22ddc00caaf896e0f9421))
- **frontend**: fully connect application to backend apis and remove mock data ([28a0adc](https://github.com/Abdo30004/discord-subscriptions/commit/28a0adc7c5f53758b7c27567fcf67017751ecb5d))
- **catalog**: register /api/v1/bots route aliases alongside /api/v1/catalog/bots ([17cec58](https://github.com/Abdo30004/discord-subscriptions/commit/17cec58d7355edb0633b445c8e707c1df34bed88))
- **ci**: implement comprehensive github actions ci pipeline and release cd workflow ([4ac6c77](https://github.com/Abdo30004/discord-subscriptions/commit/4ac6c77ea81803eef9877fca9773c56947589ee3))
- **frontend**: build next.js 16 turbopack dashboard, subscription store, checkout, and admin panel ([bb0693b](https://github.com/Abdo30004/discord-subscriptions/commit/bb0693b5355e437919ccac1b21c4962c96919275))
- **manager-bot**: implement typescript discord.js v14 manager bot with interactive fleet controls ([e0ae3b3](https://github.com/Abdo30004/discord-subscriptions/commit/e0ae3b30f9513738b278bf139d216b074fad4f3e))
- **monitor**: implement multi-target guild health watchdog, latency telemetry, and alerts ([86837d7](https://github.com/Abdo30004/discord-subscriptions/commit/86837d73f68af2827e4dde0a41984462fa4b19cd))
- **deploy**: implement bot orchestrator, zero-setup token pool, and persona customization ([fbd4e9b](https://github.com/Abdo30004/discord-subscriptions/commit/fbd4e9b6cc5d7f77133d565aad900fb88810932d))
- **billing**: implement subscription lifecycle, promo codes, vouchers, and paypal webhooks ([1002eb0](https://github.com/Abdo30004/discord-subscriptions/commit/1002eb0c2854ac641a705cdaca04f7d91ecca1ae))
- **auth**: implement discord oauth2 authentication and jwt session microservice ([9d77ba2](https://github.com/Abdo30004/discord-subscriptions/commit/9d77ba23c3ebe0a9ddac58b0ba840cc77f413095))
- **catalog**: implement bot templates and subscription plans catalog microservice ([1348d14](https://github.com/Abdo30004/discord-subscriptions/commit/1348d14c6c1121ba89abbb6c3dbea20d3cf1995d))
- **shared**: add canonical domain events, amqp messaging, vault client, and health checker ([8aa978a](https://github.com/Abdo30004/discord-subscriptions/commit/8aa978aad617bc95907f4161957e4654a8d7dec2))

### 🐛 Bug Fixes & Resilience

- **go**: resolve static analysis linter warnings and upgrade e2e to go 1.26 ([046a1b1](https://github.com/Abdo30004/discord-subscriptions/commit/046a1b1aaa108e474181e59b440379f33ea1efa4))
- **docs**: resolve syntax errors across all markdown mermaid diagrams ([312204b](https://github.com/Abdo30004/discord-subscriptions/commit/312204bacc4b671b19bc503fa4b5476dc678c681))

### 🌐 Gateway, Traefik & Networking

- **gateway**: add traefik service-circuit-breaker middleware ([8fe3f9c](https://github.com/Abdo30004/discord-subscriptions/commit/8fe3f9c62d9f366e73292a2220ea7fb0131b2de6))
- **gateway**: route all services through traefik with zero host port exports ([04df03a](https://github.com/Abdo30004/discord-subscriptions/commit/04df03a4450c07d4014e81f0017e8983b10ac0bd))
- **gateway**: configure traefik v3 reverse proxy and docker compose orchestration ([133cb4e](https://github.com/Abdo30004/discord-subscriptions/commit/133cb4eaf0ca49bc8d2721d757b915ca4275feb0))

### ☁️ Kubernetes & Cloud Orchestration

- **k8s**: add pod disruption budgets, hpa scaling, and prometheus monitoring ([72b82a1](https://github.com/Abdo30004/discord-subscriptions/commit/72b82a165b0bf08c705c094d2f39ba773ae797bc))
- **k8s**: define cloud-native kubernetes manifests, rbac, network policies, and ingress ([998747c](https://github.com/Abdo30004/discord-subscriptions/commit/998747cd5528eaa786c6753e208a14409cbfe91f))

### 🏗️ Infrastructure & Persistence

- **infra**: provision isolated per-service postgres users and least-privilege db ownership ([6372dcc](https://github.com/Abdo30004/discord-subscriptions/commit/6372dcc25ba3ed90081fe7b666b0853f0d018f9c))
- **infra**: restrict traefik cors, secure dashboard, mount vault pvc, and template k8s secrets ([d7b365a](https://github.com/Abdo30004/discord-subscriptions/commit/d7b365acc7123e976e4a1024cee3617c47a6efc5))
- **infra**: integrate manager-bot into Traefik network mesh in Docker Compose ([aa79b2f](https://github.com/Abdo30004/discord-subscriptions/commit/aa79b2f112880ece49622bfdddcbf81f28caa332))
- **infra**: initialize repository scaffolding, multi-workspace go tooling, and database seeds ([a700cce](https://github.com/Abdo30004/discord-subscriptions/commit/a700cce8b14faf23cb01fdddfc3bd93766fcd24c))

### 📚 Documentation & Architecture Guides

- upgrade platform go baseline to 1.26 across architecture manuals ([f8cbe77](https://github.com/Abdo30004/discord-subscriptions/commit/f8cbe77613e898d69b279c0fe9bece96b7529d53))
- mark phase 4 completed and synchronize manager bot and ci documentation ([21fd123](https://github.com/Abdo30004/discord-subscriptions/commit/21fd12326b485bbc9dba9a8860a10aa9b11a45a4))
- add production operations runbook and update phase 3 resolutions ([41a895a](https://github.com/Abdo30004/discord-subscriptions/commit/41a895a3394ab87fad8f9e5bab4a4f0dc9298a39))
- update Phase 2 resilience and per-service database users specifications ([a2d2b85](https://github.com/Abdo30004/discord-subscriptions/commit/a2d2b85b77b3ac950f66c2f74dc5fcfdbf688782))
- update database schemas, api references, and security audit report ([ae4a11a](https://github.com/Abdo30004/discord-subscriptions/commit/ae4a11aede0525e2035d2cee5e3893d37cda178a))
- document super admin hierarchy, admin promotion endpoints, and staff panel ([8ddbf57](https://github.com/Abdo30004/discord-subscriptions/commit/8ddbf575d75fca4dbe46c364355bce32017bd280))
- update .env.example with comprehensive platform config and add env sync rule ([877f4bc](https://github.com/Abdo30004/discord-subscriptions/commit/877f4bc25e01683c1e75aa7cd727f2f056a6b8d1))
- **frontend**: update route guards and navigation visibility specification ([82f51de](https://github.com/Abdo30004/discord-subscriptions/commit/82f51de27a97738c0fe94614825305148bde231f))
- **frontend**: update architecture and route documentation for real api and auth flow ([892a402](https://github.com/Abdo30004/discord-subscriptions/commit/892a402c6aac8574356ca1c5c26fe87d90c034f5))
- specify zero host port exports and direct traefik web access ([82612d6](https://github.com/Abdo30004/discord-subscriptions/commit/82612d63888f0c492b420361724d1866a70175b8))
- provide comprehensive architectural specifications, runbooks, and agent rules ([5a19147](https://github.com/Abdo30004/discord-subscriptions/commit/5a19147acfd9ca24131b693f18ce341431c423a9))

### 🤖 CI/CD Automation

- **go**: upgrade go runner to 1.26 and configure multi-module golangci-lint ([a0798c8](https://github.com/Abdo30004/discord-subscriptions/commit/a0798c8bf163b2d0b1d38a774a775714bbcda116))
- modularize workflows with selective path triggers for targeted execution ([942c51c](https://github.com/Abdo30004/discord-subscriptions/commit/942c51ca956d38cee65ffced27ef4c59988d7f25))
- upgrade diagram validation job to node 24 for mermaid 12 compatibility ([80d9cc7](https://github.com/Abdo30004/discord-subscriptions/commit/80d9cc77820163caafa48ad9a5d5a82ea38b5dd4))
- add automated mermaid diagram validator script and ci job ([bae2a68](https://github.com/Abdo30004/discord-subscriptions/commit/bae2a6895ead1d5898a5ac855199ebd1b43c0581))
- configure multi-workspace go cache dependency pattern ([42900ba](https://github.com/Abdo30004/discord-subscriptions/commit/42900bab3969d4349e1312d6342fc25f35d3ca8e))

---

*Generated automatically with `scripts/generate-changelog.mjs` based on Conventional Commits.*
