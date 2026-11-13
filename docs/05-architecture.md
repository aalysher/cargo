# 05. Архитектура и данные

## Общая схема

```
┌──────────────┐   initData (HMAC)   ┌───────────────────────────────┐
│  TMA (React) │ ──────────────────► │  Go API (chi + pgx) /api/v1   │
└──────────────┘  JWT-сессия         └──────────────┬────────────────┘
┌──────────────┐   JWT (login)                      │ outbox
│ Админка (React│ ──────────────────►              ▼
│ веб)         │                     PostgreSQL 16 + воркер → Bot API
└──────────────┘
```

- Один бинарь, два режима: `api` и `worker` (отправка уведомлений из outbox).
- Аутентификация клиента: `POST /auth/telegram` с initData → проверка HMAC-SHA256 по токену бота → JWT (30 дней).
- MVP-упрощение: статус ведётся на уровне заказа; сущность «посылка по треку» — v2 (нужна для консолидации).
- Деньги/вес в MVP: float64 по краям системы, numeric(12,2) в БД; при выходе на прод — decimal.

## Статусы и переходы

```
new ──► accepted ──► at_warehouse_cn ──► shipped ──► in_transit ──► arrived_kg
 │          │               │              (нужен actual_weight)      │
 ▼          ▼               ▼                                            ▼
cancelled cancelled     cancelled                                 ready_for_pickup ──► delivered
                                                                       │
                                                                  cancelled
```

- Переходы только вперёд, валидация в `internal/domain` (`CanTransition`).
- Любой переход → `order_events` (аудит) + `notifications` (outbox → бот).

| Статус | Кто ставит | Условие |
|---|---|---|
| new | клиент | заявка создана |
| accepted | оператор | подтверждена |
| at_warehouse_cn | оператор | получено на складе + вписан фактический вес |
| shipped | оператор | уехал из Китая |
| in_transit | оператор | транзит (можно пакетно) |
| arrived_kg | оператор | прибыл в Бишкек |
| ready_for_pickup | оператор | рассортирован |
| delivered | оператор | выдан + финальная стоимость |
| cancelled | клиент (new/accepted) или оператор | с reason |

## Модель данных

`clients 1—∞ orders 1—∞ (order_tracks, order_events, notifications)`; справочники `tariffs`, `settings`, `admin_users`.

Полная схема — `migrations/000001_init.up.sql` (enum-типы, индексы, sequence для номеров заказов `KG-YYMMDD-NNNN`).

## REST API /api/v1

### Клиентские (Bearer JWT после обмена initData)
| Метод | Путь | Назначение |
|---|---|---|
| POST | /auth/telegram | initData → JWT |
| POST | /auth/dev | dev-вход без Telegram (только APP_ENV=dev) |
| GET/PATCH | /me | профиль / телефон |
| GET | /warehouse-address | адрес склада + личный код |
| GET | /tariffs, /tariffs/estimate | тарифы, расчёт |
| POST | /orders | создать заявку |
| GET | /orders?scope=active\|delivering\|done\|cancelled\|all&cursor= | список |
| GET | /orders/{id} | деталь + таймлайн + треки |
| POST | /orders/{id}/cancel | отмена (new/accepted) |

### Админские (часть 2)
POST /admin/auth/login; GET /admin/orders; PATCH /admin/orders/{id} (вес/линия); POST /admin/orders/{id}/status; POST/DELETE /admin/orders/{id}/tracks; GET/PUT /admin/tariffs; GET /admin/clients; GET/PUT /admin/settings.

### Уведомления (outbox)
1. Сервис-экшен → в одной транзакции `order_events` + `notifications(status=pending)`.
2. Воркер (`cmd/worker`) опрашивает pending → Bot API sendMessage с inline-кнопкой `t.me/<bot>?startapp=order_<id>` → sent/failed (backoff, 5 попыток).

## Роли
| Роль | Где | Права |
|---|---|---|
| client | TMA | свои заказы (create/list/detail/cancel в new/accepted), профиль, тарифы, адрес склада |
| operator | админка | все заказы: статусы, вес, треки; клиенты (чтение); тарифы (чтение) |
| admin | админка | всё оператора + тарифы (правка), settings, admin_users |

## Структура Go-проекта
```
cargo/
├── cmd/api/main.go, cmd/worker/main.go
├── internal/
│   ├── config/          # env
│   ├── db/              # pgx pool + авто-миграции (golang-migrate, embed)
│   ├── auth/            # telegram.go (HMAC initData), jwt.go
│   ├── domain/          # статусы/переходы, pricing, ошибки
│   ├── repo/            # SQL через pgx (DBTX = pool|tx)
│   ├── service/         # бизнес-логика: orders, clients, tariffs
│   ├── httpapi/         # chi: router, middleware, dto, handlers
│   └── telegram/bot.go  # Bot API: sendMessage + deep-link кнопка
├── migrations/          # 000001_init, 000002_seed (embed)
├── web/tma, web/admin   # React (этапы 2 и 4)
├── docker-compose.yml, Dockerfile, Makefile
└── docs/, AGENTS.md, opencode.jsonc
```

Принципы: handlers тонкие → service держит правила (переходы, цена, события) → repo — SQL. Domain не зависит от HTTP/БД.

## Отклонения от плана (осознанные)
- SQL-запросы рукописные через pgx (вместо sqlc-кодогенерации) — меньше тулчейна для одного разработчика; интерфейс repo не меняется, если позже добавим sqlc.
- Миграции лежат в корне `migrations/` и применяются автоматически при старте `api` (embed через корневой пакет `migrations`).
