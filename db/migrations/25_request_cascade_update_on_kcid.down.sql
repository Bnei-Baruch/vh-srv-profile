BEGIN;

ALTER TABLE request
DROP CONSTRAINT IF EXISTS fk_request_id;

ALTER TABLE request
ADD CONSTRAINT fk_request_id FOREIGN KEY(keycloak_id) REFERENCES users(keycloak_id);

COMMIT;
