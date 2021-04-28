-- public.users_id_seq definition

-- DROP SEQUENCE public.users_id_seq;

CREATE SEQUENCE public.users_id_seq
	INCREMENT BY 1
	MINVALUE 1
	MAXVALUE 9223372036854775807
	CACHE 1
	NO CYCLE;

-- Table: public.users

-- DROP TABLE public.users;

CREATE TABLE public.users
(
    id bigint NOT NULL DEFAULT nextval('users_id_seq'::regclass),
    created timestamp without time zone NOT NULL,
    updated timestamp without time zone NOT NULL,
    active boolean,
    roles integer[],
    first_name text COLLATE pg_catalog."default" NOT NULL,
    last_name text COLLATE pg_catalog."default" NOT NULL,
    phone text COLLATE pg_catalog."default" NOT NULL,
    email text COLLATE pg_catalog."default" NOT NULL,
    password text COLLATE pg_catalog."default" NOT NULL,
    country bigint,
    language bigint,
    birth_date date,
    gender bigint,
    token text COLLATE pg_catalog."default",
    address_1 text COLLATE pg_catalog."default",
    address_2 text COLLATE pg_catalog."default",
    address_state text COLLATE pg_catalog."default",
    address_city text COLLATE pg_catalog."default",
    address_country text COLLATE pg_catalog."default",
    address_postcode text COLLATE pg_catalog."default",
    profile_image text COLLATE pg_catalog."default",
    first_year_of_study integer,
    learning_center text COLLATE pg_catalog."default",
    ten_name text COLLATE pg_catalog."default",
    ten_id text COLLATE pg_catalog."default",
    favorite_learning_platform text COLLATE pg_catalog."default",
    native_language bigint,
    additional_languages bigint[],
    language_for_text bigint,
    language_for_notification bigint,
    language_for_video bigint,
    CONSTRAINT users_pkey PRIMARY KEY (id)
)

TABLESPACE pg_default;

ALTER TABLE public.users
    OWNER to postgres;