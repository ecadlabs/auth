BEGIN;

DROP TYPE IF EXISTS log_id_type;

ALTER TABLE log DROP COLUMN source_type;
ALTER TABLE log DROP COLUMN target_type;
ALTER TABLE log RENAME COLUMN source_id TO user_id;

COMMIT;