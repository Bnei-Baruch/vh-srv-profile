BEGIN;

ALTER TABLE "grant"
ADD COLUMN IF NOT EXISTS request_id INT NOT NULL,
ADD CONSTRAINT fk_grant_request_id FOREIGN KEY (request_id) REFERENCES request(id);

COMMIT;