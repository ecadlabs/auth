BEGIN;

CREATE TYPE tenant_type AS ENUM ('individual', 'organization');

CREATE TABLE tenants(
	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	name TEXT NOT NULL,
	added TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	modified TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	protected BOOLEAN NOT NULL DEFAULT FALSE,
	archived BOOLEAN NOT NULL DEFAULT FALSE,
	tenant_type tenant_type NOT NULL DEFAULT 'organization'
);

CREATE TYPE membership_type AS ENUM ('owner', 'member');
CREATE TYPE membership_status AS ENUM ('active', 'invited');

CREATE TABLE membership(
	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	user_id UUID REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
	tenant_id UUID REFERENCES tenants(id) ON UPDATE CASCADE ON DELETE CASCADE,
	added TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	modified TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	membership_type membership_type NOT NULL DEFAULT 'member',
	membership_status membership_status NOT NULL DEFAULT 'active',
    UNIQUE (user_id, tenant_id)
);

INSERT INTO tenants (name, tenant_type) 
SELECT email AS name, 'individual' as tenant_type FROM users;

INSERT INTO membership (user_id, tenant_id, membership_type)
SELECT users.id as user_id, tenants.id as tenant_id, 'owner' FROM users LEFT JOIN tenants ON users.email = tenants.name;

COMMIT;