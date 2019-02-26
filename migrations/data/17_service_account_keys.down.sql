BEGIN;

DROP TABLE service_account_keys;
DROP INDEX users_email_key;
ALTER TABLE users ADD CONSTRAINT users_email_key UNIQUE (email);

COMMIT;
