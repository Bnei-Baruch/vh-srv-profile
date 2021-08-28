BEGIN;

CREATE TABLE status (
    membership bool default false,
    membership_type text default 'inactive',
    ticket bool default false,
    convention bool default false,
    galaxy bool default false,
    user_id  uuid NOT NULL REFERENCES users (user_id)
);

COMMIT;
