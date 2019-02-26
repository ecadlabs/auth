BEGIN;

ALTER TABLE roles ADD COLUMN membership_id UUID;

UPDATE roles
    SET membership_id = membership.id
    FROM membership
    WHERE roles.user_id = membership.user_id AND roles.tenant_id = membership.tenant_id;

ALTER TABLE roles
    DROP COLUMN user_id,
    DROP COLUMN tenant_id,
    ADD CONSTRAINT roles_membership_id_fkey FOREIGN KEY (membership_id) REFERENCES membership(id) ON DELETE CASCADE ON UPDATE CASCADE,
    ADD CONSTRAINT roles_membership_id_role_key UNIQUE (membership_id, role);

COMMIT;
