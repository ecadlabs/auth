BEGIN;

ALTER TABLE roles DROP COLUMN tenant_id;

CREATE TABLE roles_temp (LIKE roles);

INSERT INTO roles_temp(user_id, role)
SELECT 
    DISTINCT ON (user_id, role) user_id,
    role
FROM roles;

DROP TABLE roles;

ALTER TABLE roles_temp 
RENAME TO roles; 

ALTER TABLE roles ADD PRIMARY KEY (user_id, role);

COMMIT;