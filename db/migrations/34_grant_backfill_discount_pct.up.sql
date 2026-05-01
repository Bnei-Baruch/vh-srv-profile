BEGIN;

UPDATE "grant"
SET properties = jsonb_set(properties, '{discount_pct}', '100'::jsonb)
WHERE NOT (properties ? 'discount_pct');

COMMIT;
