BEGIN;

ALTER TABLE membership
ADD COLUMN IF NOT EXISTS user_id uuid NOT NULL,
ADD CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES users(user_id);

COMMIT;