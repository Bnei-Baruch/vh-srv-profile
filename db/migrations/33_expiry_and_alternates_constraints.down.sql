BEGIN;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_alternate_email_2_not_empty;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_alternate_email_1_not_empty;
ALTER TABLE membership DROP CONSTRAINT IF EXISTS membership_expiry_matches_type;
COMMIT;
