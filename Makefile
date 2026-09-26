.PHONY: up down restart test lint build build-bot build-frontend build-all validate-docker validate-k8s check-all

up:
	docker compose up -d

down:
	docker compose down

restart:
	docker compose restart

test:
	go test -v ./shared/... ./microservices/...

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

build-all: build build-bot build-frontend

check-all: test build-all validate-docker validate-k8s

