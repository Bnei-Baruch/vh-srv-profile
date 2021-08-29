BEGIN;

ALTER TABLE users ADD countriesWithCode text;

UPDATE users
SET countriesWithCode = country_list.name
FROM country_list 
WHERE users.country = country_list.code;

ALTER TABLE users DROP CONSTRAINT country_code_fkey;

ALTER TABLE users DROP COLUMN country;

ALTER TABLE users RENAME countriesWithCode TO country;

COMMIT;