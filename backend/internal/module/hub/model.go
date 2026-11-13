package hub

import (
	"time"
)

// AppSummaryResponse defines the standard modular contract for any child app connected to Eka Hub.
// Every satellite app (F-Tracker, Task Management, Health, etc.) implements this schema on GET /summary.
type AppSummaryResponse struct {
	AppID           string                 `json:"app_id"`
	AppName         string                 `json:"app_name"`
	ConnectedUID    string                 `json:"connected_uid"`
	ConnectedEmail  string                 `json:"connected_email"`
	HighlightMetric string                 `json:"highlight_metric"`
	HighlightLabel  string                 `json:"highlight_label"`
	AIContextData   map[string]interface{} `json:"ai_context_data"`
	DeepLink        string                 `json:"deep_link"`
}

// LinkedApp represents an authenticated external application linked to a user in Eka Hub.
type LinkedApp struct {
	ID              string     `json:"id" db:"id"`
	HubUserID       string     `json:"hub_user_id" db:"hub_user_id"`
	AppID           string     `json:"app_id" db:"app_id"`
	AppName         string     `json:"app_name" db:"app_name"`
	ConnectedUID    string     `json:"connected_uid" db:"connected_uid"`
	ConnectedEmail  string     `json:"connected_email" db:"connected_email"`
	APIBaseURL      string     `json:"api_base_url" db:"api_base_url"`
	ReadOnlyToken   string     `json:"-" db:"read_only_token"`
	CustomURLScheme string     `json:"custom_url_scheme" db:"custom_url_scheme"`
	IsActive        bool       `json:"is_active" db:"is_active"`
	HighlightMetric string     `json:"highlight_metric" db:"highlight_metric"`
	HighlightLabel  string     `json:"highlight_label" db:"highlight_label"`
	AIContextData   string     `json:"ai_context_data" db:"ai_context_data"`
	LastSyncedAt    *time.Time `json:"last_synced_at" db:"last_synced_at"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// HubLinkSession handles the OAuth-like deep link handshake state between Hub and Child apps.
type HubLinkSession struct {
	SessionCode string    `json:"session_code" db:"session_code"`
	HubUserID   string    `json:"hub_user_id" db:"hub_user_id"`
	AppID       string    `json:"app_id" db:"app_id"`
	Status      string    `json:"status" db:"status"` // PENDING, EXCHANGED, EXPIRED
	CallbackURL string    `json:"callback_url" db:"callback_url"`
	ExpiresAt   time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// InitiateLinkRequest initiates a connection handshake with a child app.
type InitiateLinkRequest struct {
	AppID string `json:"app_id" validate:"required"`
}

// InitiateLinkResponse returns the session code and the pre-built deep link URL for launching the child app.
type InitiateLinkResponse struct {
	SessionCode string `json:"session_code"`
	DeepLinkURL string `json:"deep_link_url"`
	ExpiresIn   int64  `json:"expires_in_seconds"`
}

// ExchangeLinkRequest is invoked by the child app (or Hub callback) to finalize account linking.
type ExchangeLinkRequest struct {
	SessionCode     string `json:"session_code" validate:"required"`
	AppID           string `json:"app_id" validate:"required"`
	AppName         string `json:"app_name" validate:"required"`
	ConnectedUID    string `json:"connected_uid" validate:"required"`
	ConnectedEmail  string `json:"connected_email" validate:"required,email"`
	APIBaseURL      string `json:"api_base_url" validate:"required,url"`
	ReadOnlyToken   string `json:"read_only_token" validate:"required"`
	CustomURLScheme string `json:"custom_url_scheme" validate:"required"`
}

// UpdateConnectionRequest allows updating the linked account when user chooses "Perbarui Koneksi Hub ke Akun Saat Ini".
type UpdateConnectionRequest struct {
	AppID          string `json:"app_id" validate:"required"`
	ConnectedUID   string `json:"connected_uid" validate:"required"`
	ConnectedEmail string `json:"connected_email" validate:"required,email"`
	ReadOnlyToken  string `json:"read_only_token,omitempty"`
}

// SuggestedAction is an actionable recommendation from AI with a safe expected_uid deep link.
type SuggestedAction struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	AppID       string `json:"app_id"`
	DeepLink    string `json:"deep_link"`
}

// DailyBriefingResponse is the payload for the Hub main screen.
type DailyBriefingResponse struct {
	GeneratedAt      time.Time            `json:"generated_at"`
	DailyAIBriefing  string               `json:"daily_ai_briefing"`
	MiniSummaryCards []AppSummaryResponse `json:"mini_summary_cards"`
	SuggestedActions []SuggestedAction    `json:"suggested_actions"`
}

// AIChatRequest sends a user prompt to the Cross-App AI Agent.
type AIChatRequest struct {
	Message string `json:"message" validate:"required"`
}

// ToolCallDetail records a tool invoked by the AI model.
type ToolCallDetail struct {
	Name      string      `json:"name"`
	Arguments interface{} `json:"arguments"`
	Result    interface{} `json:"result,omitempty"`
}

// AIChatResponse returns the model response, invoked tools, and actionable deep links.
type AIChatResponse struct {
	Reply       string            `json:"reply"`
	ToolCalls   []ToolCallDetail  `json:"tool_calls,omitempty"`
	ActionLinks []SuggestedAction `json:"action_links,omitempty"`
}

// ExecuteToolRequest allows calling an integrated satellite tool directly.
type ExecuteToolRequest struct {
	ToolName  string                 `json:"tool_name" validate:"required"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolExecutionResult represents the output from executing a child app tool.
type ToolExecutionResult struct {
	ToolName string           `json:"tool_name"`
	Success  bool             `json:"success"`
	Message  string           `json:"message"`
	Data     interface{}      `json:"data,omitempty"`
	Action   *SuggestedAction `json:"action,omitempty"`
}

// HubToolsListResponse lists all available tools from linked satellite apps.
type HubToolsListResponse struct {
	TotalTools int                         `json:"total_tools"`
	Tools      []GeminiFunctionDeclaration `json:"tools"`
}

// MCP JSON-RPC 2.0 structures for communicating with child app MCP endpoints.
type MCPJSONRPCRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      interface{}            `json:"id"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

type MCPJSONRPCResponse struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      interface{}      `json:"id"`
	Result  interface{}      `json:"result,omitempty"`
	Error   *MCPJSONRPCError `json:"error,omitempty"`
}

type MCPJSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// StoredChatMessage represents a persisted AI chat message for multi-turn conversational context.
type StoredChatMessage struct {
	ID        string    `json:"id" db:"id"`
	HubUserID string    `json:"hub_user_id" db:"hub_user_id"`
	Role      string    `json:"role" db:"role"`
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
