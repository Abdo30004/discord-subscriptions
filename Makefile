.PHONY: up down restart test lint build build-bot build-frontend build-all validate-docker validate-k8s changelog changelog-check check-all

up:
	docker compose up -d

down:
	docker compose down

restart:
	docker compose restart

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

validate-k8s:
	kubectl kustomize k8s/

changelog:
	node scripts/generate-changelog.mjs

changelog-check:
	node scripts/generate-changelog.mjs --check

build-all: build build-bot build-frontend

check-all: test build-all validate-docker validate-k8s changelog-check


