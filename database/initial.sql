BEGIN;

CREATE TABLE users
(
    user_id               uuid NOT NULL primary key,
    keycloak_id           TEXT NOT NULL,
    updated_at            timestamptz,
    created_at            timestamptz,
    deleted               bool    default false,
    first_name_latin      text NOT NULL,
    first_name_vernacular text,
    last_name_latin       text NOT NULL,
    last_name_vernacular  text,
    street_address        text,
    country               text,
    state_region          text,
    postal_code           text,
    gender                text NOT NULL,
    marital_status        text NOT NULL,
    date_of_birth         date,
    listening_language    text NOT NULL,
    reading_language      text NOT NULL,
    email_language        text NOT NULL,
    study_start_year      int  NOT NULL,
    study_framework       text NOT NULL,
    has_ten_group         boolean default false,
    wants_ten_group       boolean default false,
    name_of_ten_group     text
);

CREATE TABLE emails
(
    user_id       uuid REFERENCES users,
    email         text not null,
    primary_email bool default false,
    UNIQUE (user_id, email)
);

CREATE TYPE phone_type as ENUM ('WhatsApp', 'Telegram');

CREATE TABLE phone_numbers
(
    user_id      uuid REFERENCES users,
    phone_number bigint not null,
    type         phone_type,

    UNIQUE (user_id, phone_number)
);

CREATE TABLE languages
(
    user_id        uuid REFERENCES users,
    language       text not null,
    first_language bool,
    UNIQUE (user_id, language)
);

COMMIT;
