.PHONY: dev worker test vet build db-up db-down tidy admin admin-build dev-all

dev: ## запустить API (порт 8080, авто-миграции при старте)
	go run ./cmd/api

worker: ## запустить воркер уведомлений (outbox → Telegram)
	go run ./cmd/worker

tma: ## dev-сервер TMA (порт 5175)
	npm --prefix web/tma run dev

tma-build: ## сборка TMA (tsc + vite)
	npm --prefix web/tma run build

admin: ## dev-сервер админки (порт 5176)
	npm --prefix web/admin run dev

admin-build: ## сборка админки (tsc + vite)
	npm --prefix web/admin run build

dev-all: ## полный локальный стек: db/api/worker/admin + tunnel + фронты
	./scripts/dev.sh

db-up: ## поднять Postgres
	docker compose up -d db

db-down: ## остановить и удалить данные
	docker compose down -v

test:
	go test ./...

vet:
	go vet ./...

build:
	CGO_ENABLED=0 go build -o bin/api ./cmd/api
	CGO_ENABLED=0 go build -o bin/worker ./cmd/worker

tidy:
	go mod tidy
