BEGIN;

CREATE TABLE IF NOT EXISTS notification (
    id                              SERIAL PRIMARY KEY,
    slub                            TEXT NOT NULL,
    content                         JSON,
    created_at                      TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at                      TIMESTAMP WITH TIME ZONE DEFAULT now(),
    deleted_at                      TIMESTAMP WITH TIME ZONE DEFAULT null
);

COMMIT;