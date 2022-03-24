BEGIN;

CREATE TABLE IF NOT EXISTS request (
    request_name TEXT NOT NULL
    keycloak_id TEXT NOT NULL
    status TEXT NOT NULL
    request_note TEXT
    rejection_note TEXT
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT now(),
    CONSTRAINT fk_request_id FOREIGN KEY(keycloak_id) REFERENCES users(keycloak_id)
);

COMMIT;