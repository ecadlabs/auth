BEGIN;

CREATE TYPE account_type AS ENUM ('regular', 'service');

ALTER TABLE users ADD COLUMN account_type account_type NOT NULL DEFAULT 'regular';

CREATE TABLE service_account_ip(
    addr INET UNIQUE,
	user_id UUID REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
    added TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	modified TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

COMMIT;
