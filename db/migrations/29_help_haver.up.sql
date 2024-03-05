BEGIN;

ALTER TABLE request
    DROP COLUMN event_slug;

ALTER TABLE "grant"
    DROP COLUMN amount,
    DROP COLUMN currency,
    DROP COLUMN loaned,
    DROP COLUMN granted,
    DROP COLUMN repayed,
    DROP COLUMN deleted_at,
    ADD COLUMN properties JSONB DEFAULT null,
    ADD UNIQUE(request_id);

DROP TABLE grant_membership;

COMMIT;
