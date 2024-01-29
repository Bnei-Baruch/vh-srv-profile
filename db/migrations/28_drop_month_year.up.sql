BEGIN;

ALTER TABLE membership
    DROP CONSTRAINT IF EXISTS uq_month_year_user_id,
    DROP COLUMN month,
    DROP COLUMN year;

COMMIT;
