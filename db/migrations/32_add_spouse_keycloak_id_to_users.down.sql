BEGIN;

ALTER TABLE users
DROP CONSTRAINT fk_spouse_keycloak_id;

ALTER TABLE users
DROP COLUMN spouse_keycloak_id;

COMMIT;
