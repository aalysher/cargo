# AGENTS.md — правила работы в этом репозитории

## Проект
Карго-компания Китай → Бишкек. Клиентский интерфейс — Telegram Mini App (`web/tma`), админка — веб (`web/admin`), бэкенд — Go API + Postgres. Этапы продукта описаны в `docs/01…05`.

## Стек
- Go 1.27 (chi, pgx/v5, golang-migrate, golang-jwt, bcrypt)
- PostgreSQL 16
- React + Vite + Tailwind (web/tma, web/admin)

## Команды
```sh
docker compose up -d db      # поднять Postgres
make dev                     # go run ./cmd/api (порт 8080, авто-миграции)
make worker                  # go run ./cmd/worker (outbox → Telegram)
make test                    # go test ./...
go vet ./...                 # перед коммитом
```

## Структура
- `cmd/api`, `cmd/worker` — точки входа.
- `internal/domain` — статусы/переходы, pricing, ошибки. Не зависит от HTTP и БД.
- `internal/repo` — SQL через pgx (`DBTX` = pool|tx). Запросы здесь, не в handlers.
- `internal/service` — бизнес-логика (переходы статусов, цена, события + outbox в одной транзакции).
- `internal/httpapi` — chi-роутер, middleware, dto, тонкие handlers.
- `migrations/` — `NNNNNN_name.{up,down}.sql`. Применённые файлы не редактировать; следующие номера — с 000003.
- `docs/` — продукт-документация по этапам (источник истины по фичам и UX).

## Конвенции
- Идентификаторы кода — английский; пользовательские строки (тексты статусов, уведомлений) — русский, в `internal/domain` (StatusTitle) и `internal/service`.
- Ошибки: `domain.ErrNotFound`, `domain.ErrInvalidTransition`, `domain.ValidationError{Detail}`; маппинг в HTTP — в `httpapi`.
- Изменение статуса заказа только через `service.ChangeStatus` (валидация перехода + событие + уведомление).
- Деньги/вес: float64 на краях, numeric в БД.
- Комментарии — по делу; каждый экспортированный тип/функция домена и сервиса — с комментарием.

## Инструменты
- MCP: playwright (браузер для тестов TMA/админки), context7 (документация библиотек) — `opencode.jsonc`.
- LSP включён (`lsp: true`): gopls для Go.

## Языки ответов пользователю
Общение с владельцем проекта — на русском.
