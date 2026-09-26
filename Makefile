.PHONY: up down restart up-dev down-dev up-prod down-prod test lint build build-bot build-frontend build-all validate-docker validate-k8s changelog changelog-check check-all

up:
	docker compose up -d

down:
	docker compose down

restart:
	docker compose restart

up-dev:
	docker compose -f docker-compose.dev.yml --env-file .env.dev up -d

down-dev:
	docker compose -f docker-compose.dev.yml --env-file .env.dev down

up-prod:
	docker compose -f docker-compose.prod.yml --env-file .env.prod up -d

down-prod:
	docker compose -f docker-compose.prod.yml --env-file .env.prod down

test:
	go test -v ./shared/... ./microservices/auth-svc/... ./microservices/catalog-svc/... ./microservices/billing-svc/... ./microservices/deploy-svc/... ./microservices/monitor-svc/...

lint:
	golangci-lint run ./...

build:
	go build ./shared/... ./microservices/catalog-svc/... ./microservices/deploy-svc/... ./microservices/monitor-svc/... ./microservices/billing-svc/... ./microservices/auth-svc/...

build-bot:
	cd bots/manager-bot && npm run build

build-frontend:
	cd frontend && npm run build

validate-docker:
	docker compose config --quiet
	docker compose -f docker-compose.dev.yml --env-file .env.dev config --quiet
	docker compose -f docker-compose.prod.yml --env-file .env.prod config --quiet

validate-k8s:
	kubectl kustomize k8s/

validate-diagrams:
	node scripts/validate-diagrams.mjs

changelog:
	node scripts/generate-changelog.mjs

changelog-check:
	node scripts/generate-changelog.mjs --check

build-all: build build-bot build-frontend

check-all: test build-all validate-docker validate-k8s validate-diagrams changelog-check



