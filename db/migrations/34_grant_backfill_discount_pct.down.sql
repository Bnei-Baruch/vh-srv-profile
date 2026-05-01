BEGIN;

UPDATE "grant"
SET properties = properties - 'discount_pct';

COMMIT;
