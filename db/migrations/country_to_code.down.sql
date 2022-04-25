BEGIN;

ALTER TABLE users DROP CONSTRAINT country_code_fkey;

update users
set country = muc.country 
from migration_users_country muc
where users.user_id = muc.user_id;

DROP TABLE migration_users_country;

COMMIT;