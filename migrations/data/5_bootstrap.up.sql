BEGIN;

CREATE TABLE bootstrap(
	val BOOLEAN NOT NULL DEFAULT FALSE
);

INSERT INTO bootstrap (val) VALUES (FALSE);

COMMIT;