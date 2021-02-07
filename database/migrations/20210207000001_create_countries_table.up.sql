-- Table: public.countries

-- DROP TABLE public.countries;

CREATE TABLE public.countries
(
    id bigint NOT NULL DEFAULT nextval('countries_id_seq'::regclass),
    created timestamp without time zone NOT NULL,
    updated timestamp without time zone NOT NULL,
    active boolean,
    name json NOT NULL,
    CONSTRAINT countries_pkey PRIMARY KEY (id)
)

TABLESPACE pg_default;

ALTER TABLE public.countries
    OWNER to postgres;