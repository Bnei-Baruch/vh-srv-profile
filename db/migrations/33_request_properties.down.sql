-- Restore months column from properties, then drop properties
ALTER TABLE request ADD COLUMN months INTEGER;

UPDATE request SET months = (properties->>'months')::integer WHERE properties IS NOT NULL AND properties ? 'months';

ALTER TABLE request DROP COLUMN properties;
