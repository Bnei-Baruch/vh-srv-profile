-- Table: public.social_accounts

-- DROP TABLE public.social_accounts;

CREATE TABLE public.social_accounts
(
    id bigint NOT NULL DEFAULT nextval('social_accounts_id_seq'::regclass),
    created timestamp without time zone,
    updated timestamp without time zone,
    active boolean,
    user_id bigint,
    social_id bigint,
    username text COLLATE pg_catalog."default",
    url text COLLATE pg_catalog."default",
    token text COLLATE pg_catalog."default",
    CONSTRAINT social_accounts_pkey PRIMARY KEY (id)
)

TABLESPACE pg_default;

ALTER TABLE public.social_accounts
    OWNER to postgres;