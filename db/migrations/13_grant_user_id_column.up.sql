BEGIN;

ALTER TABLE "grant"
ADD COLUMN IF NOT EXISTS user_id uuid NOT NULL,
ADD COLUMN IF NOT EXISTS cancelled_at TIMESTAMP WITH TIME ZONE DEFAULT null;

ALTER TABLE grant_membership
DROP COLUMN IF EXISTS user_id;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_grant_user_id') THEN
        ALTER TABLE "grant" ADD CONSTRAINT fk_grant_user_id FOREIGN KEY(user_id) REFERENCES users(user_id);
    END IF;
END$$;

COMMIT;