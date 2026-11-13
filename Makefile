.PHONY: dev worker test vet build db-up db-down tidy

dev: ## запустить API (порт 8080, авто-миграции при старте)
	go run ./cmd/api

worker: ## запустить воркер уведомлений (outbox → Telegram)
	go run ./cmd/worker

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
