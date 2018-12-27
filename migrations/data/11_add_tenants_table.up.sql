BEGIN;

CREATE TABLE tenants(
	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	name TEXT,
	added TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	modified TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	protected BOOLEAN NOT NULL DEFAULT FALSE,
	archived BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TYPE membership_type AS ENUM ('owner', 'member');

CREATE TABLE membership(
	user_id UUID REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
	tenant_id UUID REFERENCES tenants(id) ON UPDATE CASCADE ON DELETE CASCADE,
	added TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	modified TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	mem_type membership_type NOT NULL DEFAULT 'member',
    PRIMARY KEY (user_id, tenant_id)
);

INSERT INTO tenants (name, protected) VALUES ('root', TRUE);

/* Add all previous users to the root tenant */
INSERT INTO membership (user_id, tenant_id)
SELECT id AS user_id, (SELECT id AS tenant_id FROM tenants WHERE name like 'root' LIMIT 1) FROM users;

COMMIT;