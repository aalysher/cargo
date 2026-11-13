-- Демо-операторы (пароли заменить в prod через UPDATE admin_users SET password_hash = ...)
BEGIN;

INSERT INTO admin_users (username, password_hash, display_name, role) VALUES
('admin',    '$2a$10$hC1G7oDu9ERjn.ra8S4anOmMkf9j5l2//t.2sflalRVYJo7qhuhMK', 'Демо-админ',    'admin'),
('operator', '$2a$10$VlIKJKtAD/EdA2oTEkEMLuzpPMFh84y28OEhuG/C4t6tn3JVYqP6u', 'Демо-оператор', 'operator')
ON CONFLICT (username) DO NOTHING;

COMMIT;
