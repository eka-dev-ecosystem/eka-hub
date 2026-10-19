-- Create linked_apps table
CREATE TABLE IF NOT EXISTS linked_apps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hub_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    app_id VARCHAR(64) NOT NULL,
    app_name VARCHAR(128) NOT NULL,
    connected_uid VARCHAR(128) NOT NULL,
    connected_email VARCHAR(255) NOT NULL,
    api_base_url VARCHAR(255) NOT NULL,
    read_only_token TEXT NOT NULL,
    custom_url_scheme VARCHAR(64) NOT NULL DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT true,
    highlight_metric VARCHAR(128) DEFAULT '',
    highlight_label VARCHAR(128) DEFAULT '',
    ai_context_data JSONB DEFAULT '{}'::jsonb,
    last_synced_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_hub_user_app UNIQUE (hub_user_id, app_id)
);

CREATE INDEX IF NOT EXISTS idx_linked_apps_user ON linked_apps(hub_user_id);
CREATE INDEX IF NOT EXISTS idx_linked_apps_app_id ON linked_apps(app_id);

CREATE TRIGGER update_linked_apps_updated_at
    BEFORE UPDATE ON linked_apps
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create hub_link_sessions for OAuth-like handshake
CREATE TABLE IF NOT EXISTS hub_link_sessions (
    session_code VARCHAR(64) PRIMARY KEY,
    hub_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    app_id VARCHAR(64) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'PENDING',
    callback_url VARCHAR(255),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_hub_sessions_user ON hub_link_sessions(hub_user_id);
CREATE INDEX IF NOT EXISTS idx_hub_sessions_code ON hub_link_sessions(session_code);

-- Create ai_chat_messages for cross-app AI chat
CREATE TABLE IF NOT EXISTS ai_chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hub_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(32) NOT NULL,
    content TEXT NOT NULL,
    tool_calls JSONB DEFAULT NULL,
    tool_responses JSONB DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_chat_messages_user ON ai_chat_messages(hub_user_id, created_at ASC);
