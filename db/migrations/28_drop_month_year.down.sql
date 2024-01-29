BEGIN;

ALTER TABLE membership
    ADD COLUMN month INT NULL,
    ADD COLUMN year  INT NULL;

UPDATE membership
set month=date_part('month', CURRENT_DATE),
    year=date_part('year', CURRENT_DATE);

ALTER TABLE membership
    ALTER COLUMN month SET NOT NULL,
    ALTER COLUMN year SET NOT NULL;

COMMIT;
