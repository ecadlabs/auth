BEGIN;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

ALTER TABLE log ADD COLUMN id UUID PRIMARY KEY DEFAULT uuid_generate_v4();

UPDATE log SET user_id = uuid_nil() WHERE user_id IS NULL;
UPDATE log SET target_id = uuid_nil() WHERE target_id IS NULL;
UPDATE log SET addr = '' WHERE addr IS NULL;

ALTER TABLE log
	ALTER COLUMN event SET NOT NULL,
	ALTER COLUMN addr SET NOT NULL,
	ALTER COLUMN user_id SET NOT NULL,
	ALTER COLUMN target_id SET NOT NULL;

COMMIT;
