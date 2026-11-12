-- Rollback OAuth fields from users table
DROP INDEX IF EXISTS idx_users_apple_id;
DROP INDEX IF EXISTS idx_users_google_id;

ALTER TABLE users DROP COLUMN IF EXISTS avatar_url;
ALTER TABLE users DROP COLUMN IF EXISTS apple_id;
ALTER TABLE users DROP COLUMN IF EXISTS google_id;
ALTER TABLE users DROP COLUMN IF EXISTS provider_id;
ALTER TABLE users DROP COLUMN IF EXISTS auth_provider;

