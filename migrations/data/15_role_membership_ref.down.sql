ALTER TABLE roles
    ADD COLUMN user_id UUID REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
    ADD COLUMN tenant_id UUID REFERENCES tenants(id) ON UPDATE CASCADE ON DELETE CASCADE,
    DROP COLUMN membership_id;