-- Add password column to users table for authentication
ALTER TABLE users ADD COLUMN IF NOT EXISTS password VARCHAR(255);

-- Create index on email for faster lookups
CREATE INDEX IF NOT EXISTS idx_users_email_login ON users(email) WHERE is_active = true;

