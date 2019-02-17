BEGIN;

CREATE TABLE service_account_keys(
    id UUID NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    membership_id UUID NOT NULL REFERENCES membership(id) ON DELETE CASCADE ON UPDATE CASCADE,
    added TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

ALTER TABLE users DROP CONSTRAINT users_email_key;
CREATE UNIQUE INDEX users_email_key ON users(email) WHERE account_type = 'regular';

COMMIT;
