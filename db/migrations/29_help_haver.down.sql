BEGIN;

ALTER TABLE request
    ADD COLUMN event_slug text null;

ALTER TABLE "grant"
    ADD COLUMN amount integer not null,
    ADD COLUMN currency text not null,
    ADD COLUMN loaned integer not null,
    ADD COLUMN granted integer not null,
    ADD COLUMN repayed integer,
    ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE DEFAULT null,
    DROP COLUMN properties;
    
CREATE TABLE IF NOT EXISTS grant_membership (
    id                              SERIAL PRIMARY KEY,
    grant_id                        INT NOT NULL,
    user_id                         uuid NOT NULL,
    nb_months                       INT NOT NULL,
    months_used                     INT NOT NULL,
    months_left                     INT NOT NULL,
    created_at                      TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at                      TIMESTAMP WITH TIME ZONE DEFAULT now(),
    deleted_at                      TIMESTAMP WITH TIME ZONE DEFAULT null,
    CONSTRAINT fk_user_id           FOREIGN KEY(user_id) REFERENCES users(user_id),
    CONSTRAINT fk_grant_id          FOREIGN KEY(grant_id) REFERENCES "grant"(id)
);

COMMIT;
