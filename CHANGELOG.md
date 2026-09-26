# Changelog

All notable changes to the **Discord Bot Subscription & Turnkey Fleet Management Platform** will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project strictly adheres to [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/).

---

## [v1.0.0] - 2026-09-26

### 🚀 Features & Capabilities

- **changelog**: add automated conventional commits changelog generator and tooling ([aa06c7d](https://github.com/Abdo30004/discord-subscriptions/commit/aa06c7de10c6ef0a4e70dc614727cb9bd260b80b))
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

- **changelog**: ignore changelog and release meta-commits during generation ([59e35d2](https://github.com/Abdo30004/discord-subscriptions/commit/59e35d23d38c94df680dd15050cc518c4ab172be))

### 🌐 Gateway, Traefik & Networking

- **gateway**: configure traefik v3 reverse proxy and docker compose orchestration ([133cb4e](https://github.com/Abdo30004/discord-subscriptions/commit/133cb4eaf0ca49bc8d2721d757b915ca4275feb0))

### ☁️ Kubernetes & Cloud Orchestration

- **k8s**: define cloud-native kubernetes manifests, rbac, network policies, and ingress ([998747c](https://github.com/Abdo30004/discord-subscriptions/commit/998747cd5528eaa786c6753e208a14409cbfe91f))

### 🏗️ Infrastructure & Persistence

- **infra**: initialize repository scaffolding, multi-workspace go tooling, and database seeds ([a700cce](https://github.com/Abdo30004/discord-subscriptions/commit/a700cce8b14faf23cb01fdddfc3bd93766fcd24c))

### 📚 Documentation & Architecture Guides

- document ci/cd pipelines, automated changelog system, and agent rules ([52d54f3](https://github.com/Abdo30004/discord-subscriptions/commit/52d54f386e6bbc2b264fc1990641c3012cfd119a))
- provide comprehensive architectural specifications, runbooks, and agent rules ([5a19147](https://github.com/Abdo30004/discord-subscriptions/commit/5a19147acfd9ca24131b693f18ce341431c423a9))

---

*Generated automatically with `scripts/generate-changelog.mjs` based on Conventional Commits.*
