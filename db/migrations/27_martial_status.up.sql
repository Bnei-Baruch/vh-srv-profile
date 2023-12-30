BEGIN;

INSERT INTO marital_status_types
VALUES ('Divorced'),
       ('Widowed')
ON CONFLICT (name) DO NOTHING;

COMMIT;
