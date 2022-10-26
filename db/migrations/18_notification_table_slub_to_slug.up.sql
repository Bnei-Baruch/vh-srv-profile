BEGIN;

DO
$$
    BEGIN
        ALTER TABLE notification
            RENAME COLUMN "slub" TO "slug";
    EXCEPTION
        WHEN undefined_column THEN RAISE NOTICE 'column notification.slub does not exist';
    END;
$$;

COMMIT;