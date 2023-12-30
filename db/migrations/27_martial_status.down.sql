BEGIN;

DELETE
FROM marital_status_types
WHERE name IN ('Divorced', 'Widowed');

COMMIT;
