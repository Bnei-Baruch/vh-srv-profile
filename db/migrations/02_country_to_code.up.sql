BEGIN;

create table if not exists migration_users_country
as
select u.user_id, u.country, (
    select code
    from (
        select * 
        from country_list
        union 
        select 'USA', '1', 'US'
        union
        select 'Iran, Islamic Republic of', '98', 'IR'
        union
        select 'Moldova, Republic of', '373', 'MD') cc
    where cc.code = u.country or cc."name" = u.country
    limit 1) country_code
from users u;

update users
set country = muc.country_code 
from migration_users_country muc
where users.user_id = muc.user_id;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'country_code_fkey') THEN
        ALTER TABLE users ADD CONSTRAINT country_code_fkey FOREIGN KEY(country) REFERENCES country_list(code);
    END IF;
END$$;

COMMIT;