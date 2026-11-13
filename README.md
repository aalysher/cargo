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

## Демо-доступ админки (часть 2)

Сезонные демо-операторы сидируются миграцией `000003` — **в prod сменить пароли**:

```sh
docker exec cargo-db-1 psql -U cargo -d cargo \
  -c "UPDATE admin_users SET password_hash='<bcrypt-хэш>' WHERE username IN ('admin','operator');"
```

| Логин | Пароль | Роль | Права |
|---|---|---|---|
| `admin` | `admin123` | admin | всё + правка тарифов и settings |
| `operator` | `operator123` | operator | заказы, треки, клиенты (чтение) |

## Эндпоинты админ-API

```sh
AT=$(curl -s -X POST localhost:${APP_PORT:-8085}/api/v1/admin/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin123"}' | python3 -c "import json,sys; print(json.load(sys.stdin)['token'])")

curl -s "localhost:${APP_PORT:-8085}/api/v1/admin/orders?status=new&q=KG" -H "Authorization: Bearer $AT"
curl -s localhost:${APP_PORT:-8085}/api/v1/admin/orders/2 -H "Authorization: Bearer $AT"          # деталь + таймлайн
curl -s -X PATCH localhost:${APP_PORT:-8085}/api/v1/admin/orders/2 \
  -H "Authorization: Bearer $AT" -H 'Content-Type: application/json' \
  -d '{"actual_weight_kg":2.4}'                                                                    # вес → пересчёт цены
curl -s -X POST localhost:${APP_PORT:-8085}/api/v1/admin/orders/2/status \
  -H "Authorization: Bearer $AT" -H 'Content-Type: application/json' \
  -d '{"status":"accepted","note":"забираем"}'                                                     # переход статуса
curl -s -X POST localhost:${APP_PORT:-8085}/api/v1/admin/orders/2/tracks \
  -H "Authorization: Bearer $AT" -H 'Content-Type: application/json' -d '{"tracking_number":"LBX123"}'
curl -s localhost:${APP_PORT:-8085}/api/v1/admin/tariffs -H "Authorization: Bearer $AT"            # operator: только чтение
curl -s -X PATCH localhost:${APP_PORT:-8085}/api/v1/admin/tariffs/1 \
  -H "Authorization: Bearer $AT" -H 'Content-Type: application/json' -d '{"price_per_kg":2.95}'
curl -s "localhost:${APP_PORT:-8085}/api/v1/admin/clients?q=SSZ" -H "Authorization: Bearer $AT"
curl -s localhost:${APP_PORT:-8085}/api/v1/admin/settings -H "Authorization: Bearer $AT"
```

## Админка (веб, часть 4)

```sh
npm --prefix web/admin install
npm --prefix web/admin run dev      # http://localhost:5176
npm --prefix web/admin run build    # tsc + vite → ~78 КБ gzip
```

- JWT-логин: `POST /api/v1/admin/auth/login` → токен (TTL 12ч) + профиль (роль).
- Роль `operator` — заказы, треки, клиенты (чтение); роль `admin` — плюс правка тарифов и настроек.
- Стек: React 19 + Vite 6 + Tailwind 3; в dev режиме допускается любой из демо-операторов.
- Сборка — `npm run build` (TypeScript strict, 78 КБ gzip); dev-сервер — порт 5176.

## Telegram Mini App (web/tma)

```sh
npm --prefix web/tma install
npm --prefix web/tma run dev      # http://localhost:5175
```

- Вне Telegram (браузер) вход автоматический через `POST /auth/dev` — личность хранится в localStorage.
- Внутри Telegram: `initData` → `POST /auth/telegram`, тема и MainButton/BackButton подключены.
- Боевой адрес API задаётся `VITE_API_URL` (по умолчанию `http://localhost:8085/api/v1`); сборка — `npm run build` (~83 КБ gzip).
- Deep link из бота `t.me/<bot>?startapp=order_<id>` открывает карточку заказа.

## Деплой (Docker Compose)

```sh
docker compose up -d --build   # поднимает db + api + worker + admin
docker compose logs -f api
```

- API — http://localhost:8085, админка (nginx + статика) — http://localhost:8081, Postgres — :5432.
- Миграции применяются автоматически при старте api (embed + golang-migrate).
- В prod: `APP_ENV=prod`, реальный `BOT_TOKEN`, `INIT_DATA_MAX_AGE_SECONDS=3600`, свой `JWT_SECRET`.

## Прод (план)

- `BOT_TOKEN` от @BotFather → уведомления ботом; `INIT_DATA_MAX_AGE_SECONDS=3600`.
- Валидация реального initData: `POST /api/v1/auth/telegram {"init_data": "..."}`.
- Deep link кнопки в сообщении: `t.me/<bot>?startapp=order_<id>` → открывает карточку заказа в TMA.

## Дорожная карта

| Часть | Содержание | Статус |
|---|---|---|
| 1 | Бэкенд-каркас: миграции, auth (Telegram initData), тарифы/калькулятор, заказы, outbox | ✅ |
| 2 | Админ-API: login, статусы, вес, треки, тарифы, settings, клиенты | ✅ |
| 3 | TMA: онбординг, главная, калькулятор, заявка, заказы, таймлайн | ✅ |
| 4 | Админка (веб) + деплой | ✅ (веб-админка готова; деплой — часть 5) |
