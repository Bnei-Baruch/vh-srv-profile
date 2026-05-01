-- Add properties JSONB to request, migrate months into it, then drop months column
ALTER TABLE request ADD COLUMN properties JSONB;

UPDATE request SET properties = jsonb_build_object('months', months) WHERE months IS NOT NULL;

ALTER TABLE request DROP COLUMN months;
