-- Стартовые данные. Реальные значения оператор правит через админку.
BEGIN;

-- Адрес склада в Китае (плейсхолдер — заменить на реальные данные компании)
INSERT INTO settings (key, value) VALUES
('warehouse_address', '{"contact":"Cargo Company — приём посылок","address":"广东省 广州市 白云区, склад Cargo Company, получатель: <ваш личный код>","phone":"+86 000 0000 0000","city_cn":"Гуанчжоу"}'::jsonb)
ON CONFLICT (key) DO NOTHING;

-- Демо-тарифы (ориентир по рынку Бишкека: авто ~$2.4–2.9/кг)
-- UNIQUE(transport_type, is_active) deferred — не может быть арбитром ON CONFLICT, поэтому WHERE NOT EXISTS
INSERT INTO tariffs (transport_type, price_per_kg, currency, volumetric_divisor, min_chargeable_kg, is_active)
SELECT 'road', 2.90, 'USD', 5000, 1.00, true
WHERE NOT EXISTS (SELECT 1 FROM tariffs WHERE transport_type = 'road' AND is_active);

INSERT INTO tariffs (transport_type, price_per_kg, currency, volumetric_divisor, min_chargeable_kg, is_active)
SELECT 'air', 6.50, 'USD', 6000, 1.00, true
WHERE NOT EXISTS (SELECT 1 FROM tariffs WHERE transport_type = 'air' AND is_active);

COMMIT;
