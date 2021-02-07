-- Table: public.socials

-- DROP TABLE public.socials;

CREATE TABLE public.socials
(
    id bigint NOT NULL DEFAULT nextval('socials_id_seq'::regclass),
    created timestamp without time zone,
    updated timestamp without time zone,
    active boolean,
    name text COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT socials_pkey PRIMARY KEY (id)
)

TABLESPACE pg_default;

ALTER TABLE public.socials
    OWNER to postgres;