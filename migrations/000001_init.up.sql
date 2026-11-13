-- Схема БД карго-приложения (Китай → Бишкек). См. docs/05-architecture.md
BEGIN;

CREATE TYPE order_status AS ENUM (
  'new','accepted','at_warehouse_cn','shipped','in_transit',
  'arrived_kg','ready_for_pickup','delivered','cancelled');

CREATE TYPE transport_type AS ENUM ('road','air');

CREATE TYPE event_type AS ENUM
  ('status_change','weight_set','track_added','track_removed','note','created','cancelled');

CREATE TYPE actor_type AS ENUM ('client','operator','system');
CREATE TYPE notif_status AS ENUM ('pending','sent','failed');
CREATE TYPE admin_role AS ENUM ('operator','admin');
CREATE TYPE order_source AS ENUM ('tma','admin');

-- Номера заказов: KG-YYMMDD-NNNN (порядковый из sequence)
CREATE SEQUENCE order_number_seq START 1;

CREATE TABLE clients (
  id            bigserial PRIMARY KEY,
  telegram_id   bigint UNIQUE NOT NULL,
  username      text,
  first_name    text NOT NULL DEFAULT '',
  last_name     text NOT NULL DEFAULT '',
  photo_url     text,
  phone         text,
  client_code   text UNIQUE NOT NULL,          -- личный код для опознания посылок на складе
  language      text NOT NULL DEFAULT 'ru',
  is_blocked    boolean NOT NULL DEFAULT false,
  created_at    timestamptz NOT NULL DEFAULT now(),
  updated_at    timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE admin_users (
  id            bigserial PRIMARY KEY,
  username      text UNIQUE NOT NULL,
  password_hash text NOT NULL,                 -- bcrypt
  display_name  text NOT NULL DEFAULT '',
  role          admin_role NOT NULL DEFAULT 'operator',
  is_active     boolean NOT NULL DEFAULT true,
  created_at    timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE tariffs (
  id                  bigserial PRIMARY KEY,
  transport_type      transport_type NOT NULL,
  price_per_kg        numeric(10,2) NOT NULL,
  currency            char(3) NOT NULL DEFAULT 'USD',
  volumetric_divisor  integer NOT NULL DEFAULT 5000,   -- объёмный вес = l*w*h(см)/divisor (v2)
  min_chargeable_kg   numeric(6,2) NOT NULL DEFAULT 1,
  is_active           boolean NOT NULL DEFAULT true,
  created_at          timestamptz NOT NULL DEFAULT now(),
  updated_at          timestamptz NOT NULL DEFAULT now(),
  UNIQUE (transport_type, is_active) DEFERRABLE INITIALLY DEFERRED
);

CREATE TABLE orders (
  id                  bigserial PRIMARY KEY,
  order_number        text UNIQUE NOT NULL,
  client_id           bigint NOT NULL REFERENCES clients(id),
  status              order_status NOT NULL DEFAULT 'new',
  transport_type      transport_type NOT NULL,
  contents            text NOT NULL,
  declared_weight_kg  numeric(8,2) NOT NULL,
  actual_weight_kg    numeric(8,2),
  chargeable_weight_kg numeric(8,2),
  estimated_cost      numeric(12,2),
  final_cost          numeric(12,2),
  currency            char(3) NOT NULL DEFAULT 'USD',
  comment             text,
  source              order_source NOT NULL DEFAULT 'tma',
  created_at          timestamptz NOT NULL DEFAULT now(),
  updated_at          timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX orders_client_idx ON orders (client_id, created_at DESC);
CREATE INDEX orders_status_idx ON orders (status, created_at DESC);

CREATE TABLE order_tracks (
  id              bigserial PRIMARY KEY,
  order_id        bigint NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  tracking_number text NOT NULL,
  created_at      timestamptz NOT NULL DEFAULT now(),
  UNIQUE (order_id, tracking_number)
);

CREATE TABLE order_events (
  id          bigserial PRIMARY KEY,
  order_id    bigint NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  event_type  event_type NOT NULL,
  status      order_status,                    -- для status_change / created / cancelled
  title       text NOT NULL,
  payload     jsonb NOT NULL DEFAULT '{}',
  actor_type  actor_type NOT NULL DEFAULT 'system',
  actor_id    bigint,
  created_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX order_events_order_idx ON order_events (order_id, created_at);

CREATE TABLE notifications (
  id          bigserial PRIMARY KEY,
  client_id   bigint NOT NULL REFERENCES clients(id),
  order_id    bigint REFERENCES orders(id),
  type        text NOT NULL,
  payload     jsonb NOT NULL DEFAULT '{}',
  status      notif_status NOT NULL DEFAULT 'pending',
  attempts    integer NOT NULL DEFAULT 0,
  sent_at     timestamptz,
  created_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX notifications_pending_idx ON notifications (status, created_at)
  WHERE status = 'pending';

CREATE TABLE settings (
  key        text PRIMARY KEY,
  value      jsonb NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now()
);

COMMIT;
