BEGIN;

CREATE TYPE log_id_type AS ENUM ('user', 'tenant', 'membership');
ALTER TABLE log ADD COLUMN source_type log_id_type NOT NULL DEFAULT 'user';
ALTER TABLE log ADD COLUMN target_type log_id_type NOT NULL DEFAULT 'user';
ALTER TABLE log RENAME COLUMN user_id TO source_id;

COMMIT;