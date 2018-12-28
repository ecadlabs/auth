BEGIN;

ALTER TABLE roles ADD COLUMN tenant_id UUID REFERENCES tenants(id) ON UPDATE CASCADE ON DELETE CASCADE;

/* Set all existing roles under the root tenant */
UPDATE roles SET tenant_id = (SELECT id FROM tenants WHERE name = 'root' AND protected = TRUE LIMIT 1);

COMMIT;