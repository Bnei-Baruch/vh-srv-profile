BEGIN;

ALTER TABLE request
ADD event_slug TEXT,
ADD type TEXT;

COMMIT;