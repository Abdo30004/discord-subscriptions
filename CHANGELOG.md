# Changelog

All notable changes to the **Discord Bot Subscription & Turnkey Fleet Management Platform** will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project strictly adheres to [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/).

---

## [v1.0.0] - 2026-09-26

### 🚀 Features & Capabilities

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

- **docs**: resolve syntax errors across all markdown mermaid diagrams ([312204b](https://github.com/Abdo30004/discord-subscriptions/commit/312204bacc4b671b19bc503fa4b5476dc678c681))

### 🌐 Gateway, Traefik & Networking

- **gateway**: route all services through traefik with zero host port exports ([04df03a](https://github.com/Abdo30004/discord-subscriptions/commit/04df03a4450c07d4014e81f0017e8983b10ac0bd))
- **gateway**: configure traefik v3 reverse proxy and docker compose orchestration ([133cb4e](https://github.com/Abdo30004/discord-subscriptions/commit/133cb4eaf0ca49bc8d2721d757b915ca4275feb0))

### ☁️ Kubernetes & Cloud Orchestration

- **k8s**: define cloud-native kubernetes manifests, rbac, network policies, and ingress ([998747c](https://github.com/Abdo30004/discord-subscriptions/commit/998747cd5528eaa786c6753e208a14409cbfe91f))

### 🏗️ Infrastructure & Persistence

- **infra**: initialize repository scaffolding, multi-workspace go tooling, and database seeds ([a700cce](https://github.com/Abdo30004/discord-subscriptions/commit/a700cce8b14faf23cb01fdddfc3bd93766fcd24c))

### 📚 Documentation & Architecture Guides

- **frontend**: update route guards and navigation visibility specification ([82f51de](https://github.com/Abdo30004/discord-subscriptions/commit/82f51de27a97738c0fe94614825305148bde231f))
- **frontend**: update architecture and route documentation for real api and auth flow ([892a402](https://github.com/Abdo30004/discord-subscriptions/commit/892a402c6aac8574356ca1c5c26fe87d90c034f5))
- specify zero host port exports and direct traefik web access ([82612d6](https://github.com/Abdo30004/discord-subscriptions/commit/82612d63888f0c492b420361724d1866a70175b8))
- provide comprehensive architectural specifications, runbooks, and agent rules ([5a19147](https://github.com/Abdo30004/discord-subscriptions/commit/5a19147acfd9ca24131b693f18ce341431c423a9))

### 🤖 CI/CD Automation

- modularize workflows with selective path triggers for targeted execution ([942c51c](https://github.com/Abdo30004/discord-subscriptions/commit/942c51ca956d38cee65ffced27ef4c59988d7f25))
- upgrade diagram validation job to node 24 for mermaid 12 compatibility ([80d9cc7](https://github.com/Abdo30004/discord-subscriptions/commit/80d9cc77820163caafa48ad9a5d5a82ea38b5dd4))
- add automated mermaid diagram validator script and ci job ([bae2a68](https://github.com/Abdo30004/discord-subscriptions/commit/bae2a6895ead1d5898a5ac855199ebd1b43c0581))
- configure multi-workspace go cache dependency pattern ([42900ba](https://github.com/Abdo30004/discord-subscriptions/commit/42900bab3969d4349e1312d6342fc25f35d3ca8e))

---

*Generated automatically with `scripts/generate-changelog.mjs` based on Conventional Commits.*
