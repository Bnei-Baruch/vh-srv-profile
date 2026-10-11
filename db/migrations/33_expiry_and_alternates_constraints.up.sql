BEGIN;

-- membership.expiry: only manual, helphaver and special memberships have a
-- fixed expiry; evaluation computes none for automatic, new and cancelled.
-- These rows kept an earlier type's date, and the invalidator re-evaluated
-- the past-dated ones on every run. 2,200 rows at the time of writing:
-- automatic 1,097, new 796, cancelled 307. 0 rows of the other three types
-- have no expiry.
UPDATE membership SET expiry = NULL
 WHERE type IN ('automatic', 'new', 'cancelled') AND expiry IS NOT NULL;
ALTER TABLE membership ADD CONSTRAINT membership_expiry_matches_type
  CHECK ((type IN ('manual', 'helphaver', 'special')) = (expiry IS NOT NULL));

-- users alternate emails: absent is NULL, never ''. 23 and 22 rows.
UPDATE users SET alternate_email_1 = NULL WHERE alternate_email_1 = '';
UPDATE users SET alternate_email_2 = NULL WHERE alternate_email_2 = '';
ALTER TABLE users ADD CONSTRAINT users_alternate_email_1_not_empty CHECK (alternate_email_1 <> '');
ALTER TABLE users ADD CONSTRAINT users_alternate_email_2_not_empty CHECK (alternate_email_2 <> '');

COMMIT;
