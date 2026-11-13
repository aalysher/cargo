# Cargo — доставка из Китая в Бишкек

Клиентский интерфейс — **Telegram Mini App**, бэкенд — **Go + PostgreSQL**, админка — веб.

Документация продукта: `docs/01-market-analysis.md` → `docs/05-architecture.md`.

## Быстрый старт (бэкенд)

```sh
cp .env.example .env
docker compose up -d db      # Postgres 16
make dev                     # API на :8080 (миграции применяются автоматически)
make worker                  # воркер уведомлений (нужен BOT_TOKEN)
```

## Эндпоинты (dev)

```sh
curl localhost:${APP_PORT:-8085}/healthz

# dev-вход без Telegram (только APP_ENV=dev)
curl -s -X POST localhost:${APP_PORT:-8085}/api/v1/auth/dev \
  -H 'Content-Type: application/json' \
  -d '{"telegram_id":1001,"first_name":"Айгуль"}'
# → {"token":"...","profile":{...}}

TOKEN=...
curl -s localhost:${APP_PORT:-8085}/api/v1/tariffs -H "Authorization: Bearer $TOKEN"
curl -s "localhost:${APP_PORT:-8085}/api/v1/tariffs/estimate?transport=road&weight=4.5" -H "Authorization: Bearer $TOKEN"
curl -s localhost:${APP_PORT:-8085}/api/v1/warehouse-address -H "Authorization: Bearer $TOKEN"

curl -s -X POST localhost:${APP_PORT:-8085}/api/v1/orders \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"transport_type":"road","contents":"одежда","declared_weight_kg":4.5,"tracks":["7812345678"]}'
curl -s "localhost:${APP_PORT:-8085}/api/v1/orders?scope=all" -H "Authorization: Bearer $TOKEN"
curl -s localhost:${APP_PORT:-8085}/api/v1/orders/1 -H "Authorization: Bearer $TOKEN"
curl -s -X POST localhost:${APP_PORT:-8085}/api/v1/orders/1/cancel -H "Authorization: Bearer $TOKEN"
```

## Прод (план)

- `BOT_TOKEN` от @BotFather → уведомления ботом; `INIT_DATA_MAX_AGE_SECONDS=3600`.
- Валидация реального initData: `POST /api/v1/auth/telegram {"init_data": "..."}`.
- Deep link кнопки в сообщении: `t.me/<bot>?startapp=order_<id>` → открывает карточку заказа в TMA.

## Дорожная карта

| Часть | Содержание | Статус |
|---|---|---|
| 1 | Бэкенд-каркас: миграции, auth (Telegram initData), тарифы/калькулятор, заказы, outbox | ✅ |
| 2 | Админ-API: статусы, вес, треки, тарифы, settings + dev-тарифы | ⏳ |
| 3 | TMA: онбординг, главная, калькулятор, заявка, заказы, таймлайн | ⏳ |
| 4 | Админка (веб) + деплой | ⏳ |
