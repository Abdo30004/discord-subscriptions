# Health Checks & Observability Specification

This document details the multi-tiered health checking, readiness verification, and liveness monitoring architecture implemented across all microservices, bots, frontend, and infrastructure components within the **Discord Bot Subscription Platform**.

---

## 1. Architecture & Probe Strategy

The platform adheres to Cloud-Native and Kubernetes best practices by separating health verification into two primary categories:

```mermaid
flowchart TD
    subgraph TrafficFlow [Traffic Routing & Orchestration]
        K8sKubelet["Kubernetes Kubelet / Docker Engine"]
        TraefikRouter["Traefik Reverse Proxy / K8s Ingress"]
    end

    subgraph Probes [Probe Endpoints]
        LivezProbe["Liveness Probe (/livez)\nProcess Alive, No Deadlocks"]
        ReadyzProbe["Readiness & Deep Health (/health, /readyz)\nDependencies Verified (DB, MQ, Vault)"]
    end

    subgraph ServiceCore [Microservice Core & Dependencies]
        ServiceProcess["Microservice Process"]
        PostgresDB[("PostgreSQL")]
        RabbitMQConn[["RabbitMQ"]]
        VaultCluster[("HashiCorp Vault")]
    end

    K8sKubelet -->|Polls every 10s| LivezProbe
    LivezProbe -->|Inspects process| ServiceProcess

    K8sKubelet -->|Polls every 5s| ReadyzProbe
    TraefikRouter -->|Routes only if Healthy| ReadyzProbe

    ReadyzProbe -->|db.PingContext()| PostgresDB
    ReadyzProbe -->|conn.IsConnected()| RabbitMQConn
    ReadyzProbe -->|vault.Ping()| VaultCluster
```

### Liveness vs. Readiness Probes
1. **Liveness (`/livez`)**:
   - **Purpose**: Verifies that the application process is running and responsive. Does not probe external downstream dependencies.
   - **Action on Failure**: If liveness fails continuously past threshold retries, the container orchestrator (Kubernetes or Docker) terminates and restarts the container pod.
2. **Readiness & Deep Health (`/health`, `/readyz`)**:
   - **Purpose**: Verifies that the microservice is operational **and** all mandatory dependencies (PostgreSQL database, RabbitMQ connection, HashiCorp Vault) are reachable and healthy.
   - **Action on Failure**: Returns `HTTP 503 Service Unavailable`. The container orchestrator temporarily removes the pod from ingress endpoints and service load balancing pools until dependencies recover, avoiding cascading client errors.

---

## 2. Standardized Go Health Protocol (`shared/health`)

All Go microservices use the unified `shared/health` package ([`shared/health/health.go`](../shared/health/health.go)).

### JSON Response Schema

#### Success Response (`200 OK`)
```json
{
  "status": "up",
  "service": "deploy-svc",
  "version": "1.0.0",
  "timestamp": "2026-09-26T02:40:00Z",
  "checks": {
    "database": {
      "status": "up",
      "latency": "1.12ms"
    },
    "rabbitmq": {
      "status": "up"
    },
    "vault": {
      "status": "up",
      "latency": "2.05ms"
    }
  }
}
```

#### Degraded / Failure Response (`503 Service Unavailable`)
```json
{
  "status": "down",
  "service": "deploy-svc",
  "version": "1.0.0",
  "timestamp": "2026-09-26T02:41:15Z",
  "checks": {
    "database": {
      "status": "up",
      "latency": "1.05ms"
    },
    "rabbitmq": {
      "status": "down",
      "message": "rabbitmq connection is closed"
    },
    "vault": {
      "status": "up",
      "latency": "1.98ms"
    }
  }
}
```

---

## 3. Service Health Matrix

| Service | Port | Endpoint(s) | Dependencies Checked | Failure Code |
| :--- | :--- | :--- | :--- | :--- |
| **`auth-svc`** | `8080` | `/health`, `/readyz`, `/livez` | `auth_db` (PostgreSQL) | `503` |
| **`catalog-svc`** | `8081` | `/health`, `/readyz`, `/livez` | `catalog_db` (PostgreSQL) | `503` |
| **`billing-svc`** | `8082` | `/health`, `/readyz`, `/livez` | `billing_db` (PostgreSQL), RabbitMQ AMQP | `503` |
| **`deploy-svc`** | `8083` | `/health`, `/readyz`, `/livez` | `deploy_db` (PostgreSQL), RabbitMQ AMQP, HashiCorp Vault | `503` |
| **`monitor-svc`** | `8084` | `/health`, `/readyz`, `/livez` | `monitor_db` (PostgreSQL), RabbitMQ AMQP | `503` |
| **`manager-bot`** | `8085` | `/health` | Discord Gateway WebSocket connection (`ready` / `standby`) | `503` |
| **`frontend`** | `3000` | `/api/health` | Next.js server runtime, process uptime | `503` |
| **`traefik`** | `8080` | `/ping` | Traefik entrypoint listener | Non-zero exit |
| **`postgres`** | `5432` | CLI exec | `pg_isready -U postgres` | Non-zero exit |
| **`rabbitmq`** | `5672` | CLI exec | `rabbitmq-diagnostics -q ping` | Non-zero exit |
| **`vault`** | `8200` | `/v1/sys/health` | HTTP API & `vault status` | Non-zero exit / >=500 |

---

## 4. Manager Bot Health Check (`manager-bot`)

The Discord Manager Bot runs a lightweight internal HTTP server on port `:8085` managed in [`bots/manager-bot/src/index.ts`](../bots/manager-bot/src/index.ts).

### Endpoints & Responses
- **URL**: `GET http://127.0.0.1:8085/health`
- **Connected State (`200 OK`)**:
  ```json
  {
    "status": "up",
    "service": "manager-bot",
    "gatewayStatus": "ready",
    "uptime": 1420.5,
    "pingMs": 42,
    "guilds": 15
  }
  ```
- **Standby Mode (Development without `MANAGER_BOT_TOKEN`, `200 OK`)**:
  ```json
  {
    "status": "standby",
    "service": "manager-bot",
    "message": "MANAGER_BOT_TOKEN not configured (offline mode)"
  }
  ```
- **Disconnected / Gateway Error (`503 Service Unavailable`)**:
  ```json
  {
    "status": "down",
    "service": "manager-bot",
    "gatewayStatus": "disconnected",
    "message": "Discord gateway client not ready"
  }
  ```

---

## 5. Next.js Frontend Health Check (`frontend`)

The Next.js 16 application exposes an API route handler in [`frontend/src/app/api/health/route.ts`](../frontend/src/app/api/health/route.ts).

### Endpoint & Response
- **URL**: `GET http://127.0.0.1:3000/api/health`
- **Response (`200 OK`)**:
  ```json
  {
    "status": "up",
    "service": "frontend",
    "framework": "nextjs-16",
    "timestamp": "2026-09-26T02:40:00.000Z",
    "uptime": 86400
  }
  ```

---

## 6. Docker Compose Orchestration & Dependency Gating

In [`docker-compose.yml`](../docker-compose.yml), health checks are defined on every service. Downstream containers wait for upstream dependencies using `condition: service_healthy`:

```yaml
  deploy-svc:
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://127.0.0.1:8083/health"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 5s
    depends_on:
      postgres:
        condition: service_healthy
      rabbitmq:
        condition: service_healthy
      vault:
        condition: service_healthy

  manager-bot:
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://127.0.0.1:8085/health"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 10s
    depends_on:
      catalog-svc:
        condition: service_healthy
      billing-svc:
        condition: service_healthy
      deploy-svc:
        condition: service_healthy
      monitor-svc:
        condition: service_healthy

  frontend:
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://127.0.0.1:3000/api/health"]
      interval: 15s
      timeout: 5s
      retries: 5
      start_period: 15s
    depends_on:
      auth-svc:
        condition: service_healthy
      catalog-svc:
        condition: service_healthy
      billing-svc:
        condition: service_healthy
      deploy-svc:
        condition: service_healthy
      monitor-svc:
        condition: service_healthy

  traefik:
    healthcheck:
      test: ["CMD", "traefik", "healthcheck", "--ping"]
      interval: 10s
      timeout: 5s
      retries: 3
      start_period: 5s
    depends_on:
      auth-svc:
        condition: service_healthy
      catalog-svc:
        condition: service_healthy
      billing-svc:
        condition: service_healthy
      deploy-svc:
        condition: service_healthy
      monitor-svc:
        condition: service_healthy
      frontend:
        condition: service_healthy
```

---

## 7. Kubernetes Probes Configuration

All Kubernetes Deployments in [`k8s/`](../k8s/) define tailored `readinessProbe` and `livenessProbe` specs:

```yaml
# Example: Deploy-Svc Kubernetes Probes
readinessProbe:
  httpGet:
    path: /health
    port: 8083
  initialDelaySeconds: 5
  periodSeconds: 5
livenessProbe:
  httpGet:
    path: /health
    port: 8083
  initialDelaySeconds: 15
  periodSeconds: 10

# Example: Manager Bot Kubernetes Probes
ports:
  - name: health
    containerPort: 8085
readinessProbe:
  httpGet:
    path: /health
    port: 8085
  initialDelaySeconds: 10
  periodSeconds: 10
livenessProbe:
  httpGet:
    path: /health
    port: 8085
  initialDelaySeconds: 20
  periodSeconds: 15

# Example: Frontend Kubernetes Probes
readinessProbe:
  httpGet:
    path: /api/health
    port: 3000
  initialDelaySeconds: 10
  periodSeconds: 5
livenessProbe:
  httpGet:
    path: /api/health
    port: 3000
  initialDelaySeconds: 20
  periodSeconds: 15
```

---

## 8. Verification & Diagnostics Runbook

To test health checks across running containers:

```bash
# 1. Microservice Deep Health
curl -s http://localhost:8080/health | jq .
curl -s http://localhost:8081/health | jq .
curl -s http://localhost:8082/health | jq .
curl -s http://localhost:8083/health | jq .
curl -s http://localhost:8084/health | jq .

# 2. Microservice Shallow Liveness
curl -s http://localhost:8080/livez | jq .

# 3. Manager Bot Health
curl -s http://localhost:8085/health | jq .

# 4. Frontend Health
curl -s http://localhost:3000/api/health | jq .

# 5. Traefik Ping
curl -s http://localhost:8090/ping

# 6. Docker Container Health Status
docker compose ps
```
