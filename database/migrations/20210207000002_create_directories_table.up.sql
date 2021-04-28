-- public.directories_id_seq definition

-- DROP SEQUENCE public.directories_id_seq;

CREATE SEQUENCE public.directories_id_seq
	INCREMENT BY 1
	MINVALUE 1
	MAXVALUE 9223372036854775807
	CACHE 1
	NO CYCLE;

-- Table: public.directories

-- DROP TABLE public.directories;

CREATE TABLE public.directories
(
    id bigint NOT NULL DEFAULT nextval('directories_id_seq'::regclass),
    created timestamp without time zone NOT NULL,
    updated timestamp without time zone NOT NULL,
    active boolean,
    directory_type text COLLATE pg_catalog."default" NOT NULL,
    name json,
    weight integer,
    CONSTRAINT directories_pkey PRIMARY KEY (id)
)

TABLESPACE pg_default;

ALTER TABLE public.directories
    OWNER to postgres;