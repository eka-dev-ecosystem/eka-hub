-- Remove session_id column from ai_chat_messages
ALTER TABLE ai_chat_messages DROP COLUMN IF EXISTS session_id;

-- Drop index and trigger
DROP INDEX IF EXISTS idx_chat_messages_session;
DROP INDEX IF EXISTS idx_chat_sessions_user_updated;
DROP TRIGGER IF EXISTS update_chat_sessions_updated_at ON ai_chat_sessions;

-- Drop ai_chat_sessions table
DROP TABLE IF EXISTS ai_chat_sessions;
