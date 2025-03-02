BEGIN;

CREATE TABLE IF NOT EXISTS page_notes (
    id SERIAL PRIMARY KEY,
    author_keycloak_id text NOT NULL,
    page_keycloak_id text NOT NULL,
    page_id integer NOT NULL,
    note text COLLATE pg_catalog."default" NOT NULL,
    created_at date NOT NULL default now(),
	CONSTRAINT fk_page_notes_author_keycloak_id FOREIGN KEY (author_keycloak_id)
        REFERENCES public.users (keycloak_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

COMMIT;