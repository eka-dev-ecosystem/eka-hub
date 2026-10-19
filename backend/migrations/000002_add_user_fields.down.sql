-- Drop index
DROP INDEX IF EXISTS idx_users_is_active;

-- Remove columns
ALTER TABLE users DROP COLUMN IF EXISTS last_login;
ALTER TABLE users DROP COLUMN IF EXISTS is_active;
ALTER TABLE users DROP COLUMN IF EXISTS phone;

