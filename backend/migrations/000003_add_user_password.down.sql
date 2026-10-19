-- Remove password column
ALTER TABLE users DROP COLUMN IF EXISTS password;

-- Drop index
DROP INDEX IF EXISTS idx_users_email_login;

