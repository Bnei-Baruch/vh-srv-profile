BEGIN;

ALTER TABLE users ADD countriesWithCode text;

UPDATE users
SET countriesWithCode = country_list.code
FROM country_list 
WHERE users.country = country_list.name;

ALTER TABLE users ADD CONSTRAINT country_code_fkey FOREIGN KEY(countriesWithCode) REFERENCES country_list(code);

ALTER TABLE users DROP COLUMN country;

ALTER TABLE users RENAME countriesWithCode TO country;

COMMIT;