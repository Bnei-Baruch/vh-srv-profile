BEGIN;
 
ALTER TABLE "grant"
DROP COLUMN IF EXISTS user_id,
DROP COLUMN IF EXISTS cancelled_at;

ALTER TABLE grant_membership
ADD COLUMN IF NOT EXISTS user_id uuid NOT NULL;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_grant_membership_user_id') THEN
        ALTER TABLE grant_membership ADD CONSTRAINT fk_grant_membership_user_id FOREIGN KEY(user_id) REFERENCES users(user_id);
    END IF;
END$$;

COMMIT;