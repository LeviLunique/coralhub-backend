ALTER TABLE users
ADD COLUMN password_hash TEXT;

CREATE INDEX users_active_email_idx ON users (lower(email))
WHERE active = TRUE;
