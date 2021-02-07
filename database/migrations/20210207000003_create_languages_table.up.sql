-- Table: public.languages

-- DROP TABLE public.languages;

CREATE TABLE public.languages
(
    id bigint NOT NULL DEFAULT nextval('languages_id_seq'::regclass),
    created timestamp without time zone NOT NULL,
    updated timestamp without time zone NOT NULL,
    active boolean,
    code text COLLATE pg_catalog."default" NOT NULL,
    name json NOT NULL,
    CONSTRAINT languages_pkey PRIMARY KEY (id)
)

TABLESPACE pg_default;

ALTER TABLE public.languages
    OWNER to postgres;