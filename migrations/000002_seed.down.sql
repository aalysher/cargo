BEGIN;

DELETE FROM tariffs;
DELETE FROM settings WHERE key = 'warehouse_address';

COMMIT;
