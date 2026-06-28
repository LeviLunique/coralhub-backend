DROP INDEX IF EXISTS users_active_email_idx;

ALTER TABLE users
DROP COLUMN IF EXISTS password_hash;
