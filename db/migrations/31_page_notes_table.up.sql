BEGIN;

CREATE TABLE IF NOT EXISTS page_notes (
    id SERIAL PRIMARY KEY,
    user_id uuid NOT NULL,
    page_id integer NOT NULL,
    note text COLLATE pg_catalog."default" NOT NULL,
    created_at date NOT NULL,
    modified_at date,
	CONSTRAINT fk_page_notes_user_id FOREIGN KEY (user_id)
        REFERENCES public.users (user_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

COMMIT;