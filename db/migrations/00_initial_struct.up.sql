BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS gender_types
(
    name text PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS marital_status_types
(
    name text PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS phone_number_types
(
    name text PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS country_list (
    "name" TEXT NOT NULL,
    "dial_code" TEXT NULL,
    "code" TEXT NOT NULL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS language_list
(
    code text PRIMARY KEY,
    name text
);

CREATE TABLE IF NOT EXISTS users
(
    user_id               uuid        NOT NULL PRIMARY KEY default gen_random_uuid(),
    keycloak_id           TEXT        NOT NULL UNIQUE,
    updated_at            timestamptz NOT NULL             default now(),
    created_at            timestamptz NOT NULL             default now(),
    deleted               bool                             default false,
    first_name_latin      text,
    first_name_vernacular text        NOT NULL,
    last_name_latin       text,
    last_name_vernacular  text        NOT NULL,
    street_address        text,
    country               text,
    state_region          text,
    postal_code           text,
    city                  text,
    gender                text REFERENCES gender_types (name),
    marital_status        text REFERENCES marital_status_types (name),
    date_of_birth         date,
    primary_email         text        NOT NULL,
    alternate_email_1     text,
    alternate_email_2     text,
    first_language        text REFERENCES language_list (code),
    other_language_1      text REFERENCES language_list (code),
    other_language_2      text REFERENCES language_list (code),
    other_language_3      text REFERENCES language_list (code),
    other_language_4      text REFERENCES language_list (code),
    listening_language    text REFERENCES language_list (code),
    reading_language      text REFERENCES language_list (code),
    email_language        text REFERENCES language_list (code),
    study_start_year      int,
    study_framework       text,
    has_ten_group         boolean,
    wants_ten_group       boolean,
    name_of_ten_group     text
);

CREATE TABLE IF NOT EXISTS phone_numbers
(
    user_id      uuid NOT NULL REFERENCES users (user_id),
    phone_number text NOT NULL,
    type         text NOT NULL REFERENCES phone_number_types (name),

    UNIQUE (user_id, type)
);

CREATE TABLE IF NOT EXISTS status (
    membership bool default false,
    membership_type text default 'inactive',
    ticket bool default false,
    convention bool default false,
    galaxy bool default false,
    user_id  uuid NOT NULL REFERENCES users (user_id)
);

CREATE TABLE IF NOT EXISTS request (
    id              SERIAL PRIMARY KEY,
    name            TEXT NOT NULL,
    keycloak_id     TEXT NOT NULL,
    status          TEXT NOT NULL,
    event_slug      TEXT,
    type            TEXT,
    request_note    TEXT,
    rejection_note  TEXT,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT now(),
    CONSTRAINT fk_request_id FOREIGN KEY(keycloak_id) REFERENCES users(keycloak_id)
);

COMMIT;