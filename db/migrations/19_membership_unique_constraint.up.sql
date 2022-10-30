BEGIN;

ALTER TABLE membership 
ADD CONSTRAINT uq_month_year_user_id UNIQUE (user_id, month, year);

COMMIT;