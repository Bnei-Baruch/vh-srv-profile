BEGIN;

ALTER TABLE users
ADD COLUMN spouse_keycloak_id TEXT;

ALTER TABLE users
ADD CONSTRAINT fk_spouse_keycloak_id
FOREIGN KEY (spouse_keycloak_id)
REFERENCES users(keycloak_id);

COMMIT;
