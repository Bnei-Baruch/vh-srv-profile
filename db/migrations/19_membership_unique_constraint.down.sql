BEGIN;

ALTER TABLE membership DROP CONSTRAINT uq_month_year_user_id;

COMMIT;