BEGIN;

DELETE FROM admin_users WHERE username IN ('admin', 'operator');

COMMIT;
