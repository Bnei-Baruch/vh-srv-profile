BEGIN;

ALTER TABLE users
DROP CONSTRAINT users_first_language_fkey,
DROP CONSTRAINT users_email_language_fkey,
DROP CONSTRAINT users_listening_language_fkey,
DROP CONSTRAINT users_reading_language_fkey,
DROP CONSTRAINT users_other_language_1_fkey,
DROP CONSTRAINT users_other_language_2_fkey,
DROP CONSTRAINT users_other_language_3_fkey,
DROP CONSTRAINT users_other_language_4_fkey;

UPDATE users SET first_language = (SELECT code FROM language_list WHERE name=users.first_language);
UPDATE users SET email_language = (SELECT code FROM language_list WHERE name=users.email_language);
UPDATE users SET listening_language = (SELECT code FROM language_list WHERE name=users.listening_language);
UPDATE users SET reading_language = (SELECT code FROM language_list WHERE name=users.reading_language);
UPDATE users SET other_language_1 = (SELECT code FROM language_list WHERE name=users.other_language_1);
UPDATE users SET other_language_2 = (SELECT code FROM language_list WHERE name=users.other_language_2);
UPDATE users SET other_language_3 = (SELECT code FROM language_list WHERE name=users.other_language_3);
UPDATE users SET other_language_4 = (SELECT code FROM language_list WHERE name=users.other_language_4);

ALTER TABLE language_list DROP CONSTRAINT language_list_pkey;
ALTER TABLE language_list ADD PRIMARY KEY (code);

ALTER TABLE users
ADD FOREIGN KEY(first_language) REFERENCES language_list(code),
ADD FOREIGN KEY(email_language) REFERENCES language_list(code),
ADD FOREIGN KEY(listening_language) REFERENCES language_list(code),
ADD FOREIGN KEY(reading_language) REFERENCES language_list(code),
ADD FOREIGN KEY(other_language_1) REFERENCES language_list(code),
ADD FOREIGN KEY(other_language_2) REFERENCES language_list(code),
ADD FOREIGN KEY(other_language_3) REFERENCES language_list(code),
ADD FOREIGN KEY(other_language_4) REFERENCES language_list(code);

COMMIT;
