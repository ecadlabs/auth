BEGIN;

ALTER TABLE roles ADD COLUMN tenant_id UUID REFERENCES tenants(id) ON UPDATE CASCADE ON DELETE CASCADE;

UPDATE roles SET tenant_id = (SELECT tenants.id FROM tenants LEFT JOIN users ON users.email = tenants.name WHERE users.id = roles.user_id);

ALTER TABLE roles DROP CONSTRAINT roles_pkey;
ALTER TABLE roles ADD PRIMARY KEY (tenant_id, user_id, role);

COMMIT;