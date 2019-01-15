BEGIN;
ALTER TABLE users DROP COLUMN account_type;
DROP TABLE service_account_ip;
COMMIT;
