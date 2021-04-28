-- public.socials_id_seq definition

-- DROP SEQUENCE public.socials_id_seq;

CREATE SEQUENCE public.socials_id_seq
	INCREMENT BY 1
	MINVALUE 1
	MAXVALUE 9223372036854775807
	CACHE 1
	NO CYCLE;

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