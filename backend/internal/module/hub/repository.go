package hub

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"golang/database"

	"github.com/jmoiron/sqlx"
)

var (
	ErrSessionNotFound = errors.New("link session not found")
	ErrSessionExpired  = errors.New("link session expired")
	ErrAppNotLinked    = errors.New("app is not linked to this user")
)

type HubRepository interface {
	CreateSession(ctx context.Context, session *HubLinkSession) error
	GetSession(ctx context.Context, code string) (*HubLinkSession, error)
	MarkSessionExchanged(ctx context.Context, code string) error

	UpsertLinkedApp(ctx context.Context, app *LinkedApp) error
	GetLinkedApps(ctx context.Context, userID string) ([]LinkedApp, error)
	GetLinkedApp(ctx context.Context, userID, appID string) (*LinkedApp, error)
	UpdateLinkedAppSummary(ctx context.Context, userID, appID string, metric, label string, aiContext string, syncedAt time.Time) error
	UpdateLinkedAccount(ctx context.Context, userID, appID, connectedUID, connectedEmail, token string) error
	DeleteLinkedApp(ctx context.Context, userID, appID string) error

	SaveChatMessage(ctx context.Context, userID, role, content string, toolCalls, toolResponses interface{}) error
	GetRecentChatMessages(ctx context.Context, userID string, limit int) ([]StoredChatMessage, error)
}

type hubRepository struct {
	db *sqlx.DB
}

func NewHubRepository(db *sqlx.DB) HubRepository {
	return &hubRepository{db: db}
}

// withTx executes a write operation inside a regular database transaction.
// If an active transaction already exists in context, it is reused.
func (r *hubRepository) withTx(ctx context.Context, fn func(executor database.QueryerExecer) error) error {
	if tx := database.TxFromContext(ctx); tx != nil {
		return fn(tx)
	}
	return database.WithTransactionSqlxCtx(ctx, r.db, func(tx *sqlx.Tx) error {
		return fn(tx)
	})
}

// withReadOnlyTx executes a read operation inside a read-only database transaction.
// If an active transaction already exists in context, it is reused.
func (r *hubRepository) withReadOnlyTx(ctx context.Context, fn func(executor database.QueryerExecer) error) error {
	if tx := database.TxFromContext(ctx); tx != nil {
		return fn(tx)
	}
	return database.WithReadOnlyTransactionSqlxCtx(ctx, r.db, func(tx *sqlx.Tx) error {
		return fn(tx)
	})
}

func (r *hubRepository) CreateSession(ctx context.Context, s *HubLinkSession) error {
	query := `
		INSERT INTO hub_link_sessions (session_code, hub_user_id, app_id, status, callback_url, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	return r.withTx(ctx, func(executor database.QueryerExecer) error {
		_, err := executor.ExecContext(ctx, query, s.SessionCode, s.HubUserID, s.AppID, s.Status, s.CallbackURL, s.ExpiresAt, s.CreatedAt)
		return err
	})
}

func (r *hubRepository) GetSession(ctx context.Context, code string) (*HubLinkSession, error) {
	var s HubLinkSession
	query := `
		SELECT session_code, hub_user_id, app_id, status, callback_url, expires_at, created_at
		FROM hub_link_sessions
		WHERE session_code = $1
	`
	err := r.withReadOnlyTx(ctx, func(executor database.QueryerExecer) error {
		if err := executor.GetContext(ctx, &s, query, code); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrSessionNotFound
			}
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *hubRepository) MarkSessionExchanged(ctx context.Context, code string) error {
	query := `UPDATE hub_link_sessions SET status = 'EXCHANGED' WHERE session_code = $1`
	return r.withTx(ctx, func(executor database.QueryerExecer) error {
		_, err := executor.ExecContext(ctx, query, code)
		return err
	})
}

func (r *hubRepository) UpsertLinkedApp(ctx context.Context, app *LinkedApp) error {
	query := `
		INSERT INTO linked_apps (
			hub_user_id, app_id, app_name, connected_uid, connected_email,
			api_base_url, read_only_token, custom_url_scheme, is_active,
			highlight_metric, highlight_label, ai_context_data, last_synced_at,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12, $13,
			NOW(), NOW()
		)
		ON CONFLICT (hub_user_id, app_id) DO UPDATE SET
			app_name = EXCLUDED.app_name,
			connected_uid = EXCLUDED.connected_uid,
			connected_email = EXCLUDED.connected_email,
			api_base_url = EXCLUDED.api_base_url,
			read_only_token = EXCLUDED.read_only_token,
			custom_url_scheme = EXCLUDED.custom_url_scheme,
			is_active = true,
			updated_at = NOW()
	`
	aiCtx := app.AIContextData
	if aiCtx == "" {
		aiCtx = "{}"
	}
	return r.withTx(ctx, func(executor database.QueryerExecer) error {
		_, err := executor.ExecContext(ctx, query,
			app.HubUserID, app.AppID, app.AppName, app.ConnectedUID, app.ConnectedEmail,
			app.APIBaseURL, app.ReadOnlyToken, app.CustomURLScheme, app.IsActive,
			app.HighlightMetric, app.HighlightLabel, aiCtx, app.LastSyncedAt,
		)
		return err
	})
}

func (r *hubRepository) GetLinkedApps(ctx context.Context, userID string) ([]LinkedApp, error) {
	var apps []LinkedApp
	query := `
		SELECT id, hub_user_id, app_id, app_name, connected_uid, connected_email,
		       api_base_url, read_only_token, custom_url_scheme, is_active,
		       highlight_metric, highlight_label, ai_context_data, last_synced_at,
		       created_at, updated_at
		FROM linked_apps
		WHERE hub_user_id = $1 AND is_active = true
		ORDER BY created_at ASC
	`
	err := r.withReadOnlyTx(ctx, func(executor database.QueryerExecer) error {
		return executor.SelectContext(ctx, &apps, query, userID)
	})
	if err != nil {
		return nil, err
	}
	return apps, nil
}

func (r *hubRepository) GetLinkedApp(ctx context.Context, userID, appID string) (*LinkedApp, error) {
	var app LinkedApp
	query := `
		SELECT id, hub_user_id, app_id, app_name, connected_uid, connected_email,
		       api_base_url, read_only_token, custom_url_scheme, is_active,
		       highlight_metric, highlight_label, ai_context_data, last_synced_at,
		       created_at, updated_at
		FROM linked_apps
		WHERE hub_user_id = $1 AND app_id = $2
	`
	err := r.withReadOnlyTx(ctx, func(executor database.QueryerExecer) error {
		if err := executor.GetContext(ctx, &app, query, userID, appID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrAppNotLinked
			}
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *hubRepository) UpdateLinkedAppSummary(ctx context.Context, userID, appID string, metric, label string, aiContext string, syncedAt time.Time) error {
	query := `
		UPDATE linked_apps
		SET highlight_metric = $1, highlight_label = $2, ai_context_data = $3, last_synced_at = $4, updated_at = NOW()
		WHERE hub_user_id = $5 AND app_id = $6
	`
	return r.withTx(ctx, func(executor database.QueryerExecer) error {
		_, err := executor.ExecContext(ctx, query, metric, label, aiContext, syncedAt, userID, appID)
		return err
	})
}

func (r *hubRepository) UpdateLinkedAccount(ctx context.Context, userID, appID, connectedUID, connectedEmail, token string) error {
	if token != "" {
		query := `
			UPDATE linked_apps
			SET connected_uid = $1, connected_email = $2, read_only_token = $3, updated_at = NOW()
			WHERE hub_user_id = $4 AND app_id = $5
		`
		return r.withTx(ctx, func(executor database.QueryerExecer) error {
			_, err := executor.ExecContext(ctx, query, connectedUID, connectedEmail, token, userID, appID)
			return err
		})
	}

	query := `
		UPDATE linked_apps
		SET connected_uid = $1, connected_email = $2, updated_at = NOW()
		WHERE hub_user_id = $3 AND app_id = $4
	`
	return r.withTx(ctx, func(executor database.QueryerExecer) error {
		_, err := executor.ExecContext(ctx, query, connectedUID, connectedEmail, userID, appID)
		return err
	})
}

func (r *hubRepository) DeleteLinkedApp(ctx context.Context, userID, appID string) error {
	query := `DELETE FROM linked_apps WHERE hub_user_id = $1 AND app_id = $2`
	return r.withTx(ctx, func(executor database.QueryerExecer) error {
		_, err := executor.ExecContext(ctx, query, userID, appID)
		return err
	})
}

func (r *hubRepository) SaveChatMessage(ctx context.Context, userID, role, content string, toolCalls, toolResponses interface{}) error {
	query := `
		INSERT INTO ai_chat_messages (hub_user_id, role, content, tool_calls, tool_responses, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`
	tcJSON, _ := json.Marshal(toolCalls)
	trJSON, _ := json.Marshal(toolResponses)
	return r.withTx(ctx, func(executor database.QueryerExecer) error {
		_, err := executor.ExecContext(ctx, query, userID, role, content, tcJSON, trJSON)
		return err
	})
}

func (r *hubRepository) GetRecentChatMessages(ctx context.Context, userID string, limit int) ([]StoredChatMessage, error) {
	if limit <= 0 {
		limit = 10
	}
	query := `
		SELECT id, hub_user_id, role, content, created_at
		FROM ai_chat_messages
		WHERE hub_user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	var rows []StoredChatMessage
	err := r.withTx(ctx, func(executor database.QueryerExecer) error {
		return executor.SelectContext(ctx, &rows, query, userID, limit)
	})
	if err != nil {
		return nil, err
	}

	// Reverse to chronological order (oldest first)
	for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
		rows[i], rows[j] = rows[j], rows[i]
	}
	return rows, nil
}
