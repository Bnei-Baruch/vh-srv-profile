BEGIN;

CREATE TABLE IF NOT EXISTS membership_automatic (
    id                              SERIAL PRIMARY KEY,
    order_id                        INT NOT NULL,
    payment_id                      INT NOT NULL,
    membership_id                   INT NOT NULL,
    created_at                      TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at                      TIMESTAMP WITH TIME ZONE DEFAULT now(),
    deleted_at                      TIMESTAMP WITH TIME ZONE DEFAULT null,
    CONSTRAINT fk_membership_id     FOREIGN KEY(membership_id) REFERENCES membership(id)
);

COMMIT;