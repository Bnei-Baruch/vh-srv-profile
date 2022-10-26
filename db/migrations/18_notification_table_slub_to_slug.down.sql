BEGIN;

DO
$$
    BEGIN
        ALTER TABLE notification
            RENAME COLUMN "slug" TO "slub";
    EXCEPTION
        WHEN undefined_column THEN RAISE NOTICE 'column notification.slug does not exist';
    END;
$$;

COMMIT;