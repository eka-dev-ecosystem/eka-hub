package hub

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"golang/database"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type HubService interface {
	InitiateLink(ctx context.Context, userID, appID string) (*InitiateLinkResponse, error)
	ExchangeLink(ctx context.Context, req *ExchangeLinkRequest) error
	UpdateLinkedAccount(ctx context.Context, userID string, req *UpdateConnectionRequest) error
	GetLinkedApps(ctx context.Context, userID string) ([]LinkedApp, error)
	UnlinkApp(ctx context.Context, userID, appID string) error

	SyncAllSummaries(ctx context.Context, userID string) ([]AppSummaryResponse, error)
	GetDailyBriefing(ctx context.Context, userID string) (*DailyBriefingResponse, error)
	ProcessAIChat(ctx context.Context, userID, userMessage, sessionID string) (*AIChatResponse, error)

	// Chat Session management
	GetChatSessions(ctx context.Context, userID string, limit, offset int) ([]ChatSession, error)
	GetChatSessionDetail(ctx context.Context, userID, sessionID string) (*ChatSessionDetail, error)
	DeleteChatSession(ctx context.Context, userID, sessionID string) error

	ExecuteTool(ctx context.Context, userID string, req *ExecuteToolRequest) (*ToolExecutionResult, error)
	GetAvailableTools(ctx context.Context, userID string) (*HubToolsListResponse, error)
}

type hubService struct {
	database.Transactor
	repo       HubRepository
	httpClient *http.Client
}

func NewHubService(repo HubRepository, transactor database.Transactor) HubService {
	return &hubService{
		Transactor: transactor,
		repo:       repo,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// runInTx delegates to the embedded database.Transactor.RunInTx for backward compatibility.
func (s *hubService) runInTx(ctx context.Context, fn func(txCtx context.Context) error) error {
	return s.RunInTx(ctx, fn)
}

func (s *hubService) InitiateLink(ctx context.Context, userID, appID string) (*InitiateLinkResponse, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	sessionCode := "hub_sess_" + hex.EncodeToString(b)

	scheme := "app://"
	switch strings.ToLower(appID) {
	case "ftracker", "f-tracker":
		scheme = "ftracker://"
	case "taskmanagement", "task-management", "taskhub":
		scheme = "taskhub://"
	}

	session := &HubLinkSession{
		SessionCode: sessionCode,
		HubUserID:   userID,
		AppID:       appID,
		Status:      "PENDING",
		CallbackURL: "ekahub://link-callback",
		ExpiresAt:   time.Now().Add(10 * time.Minute),
		CreatedAt:   time.Now(),
	}

	if err := s.runInTx(ctx, func(txCtx context.Context) error {
		return s.repo.CreateSession(txCtx, session)
	}); err != nil {
		return nil, fmt.Errorf("failed to save link session: %w", err)
	}

	deepLink := fmt.Sprintf("%sconnect-hub?session_code=%s&callback_url=%s", scheme, sessionCode, "ekahub://link-callback")

	return &InitiateLinkResponse{
		SessionCode: sessionCode,
		DeepLinkURL: deepLink,
		ExpiresIn:   600,
	}, nil
}

func (s *hubService) ExchangeLink(ctx context.Context, req *ExchangeLinkRequest) error {
	var linkedApp *LinkedApp
	err := s.runInTx(ctx, func(txCtx context.Context) error {
		session, err := s.repo.GetSession(txCtx, req.SessionCode)
		if err != nil {
			return err
		}

		if time.Now().After(session.ExpiresAt) {
			return ErrSessionExpired
		}

		linkedApp = &LinkedApp{
			HubUserID:       session.HubUserID,
			AppID:           req.AppID,
			AppName:         req.AppName,
			ConnectedUID:    req.ConnectedUID,
			ConnectedEmail:  req.ConnectedEmail,
			APIBaseURL:      req.APIBaseURL,
			ReadOnlyToken:   req.ReadOnlyToken,
			CustomURLScheme: req.CustomURLScheme,
			IsActive:        true,
		}

		if err := s.repo.UpsertLinkedApp(txCtx, linkedApp); err != nil {
			return fmt.Errorf("failed to save linked app: %w", err)
		}

		if err := s.repo.MarkSessionExchanged(txCtx, req.SessionCode); err != nil {
			return fmt.Errorf("failed to mark session exchanged: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Fetch initial summary right away
	if linkedApp != nil {
		go func() {
			ctxTimeout, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			_, _ = s.fetchAndPersistAppSummary(ctxTimeout, linkedApp)
		}()
	}

	return nil
}

func (s *hubService) UpdateLinkedAccount(ctx context.Context, userID string, req *UpdateConnectionRequest) error {
	return s.runInTx(ctx, func(txCtx context.Context) error {
		return s.repo.UpdateLinkedAccount(txCtx, userID, req.AppID, req.ConnectedUID, req.ConnectedEmail, req.ReadOnlyToken)
	})
}

func (s *hubService) GetLinkedApps(ctx context.Context, userID string) ([]LinkedApp, error) {
	return s.repo.GetLinkedApps(ctx, userID)
}

func (s *hubService) UnlinkApp(ctx context.Context, userID, appID string) error {
	return s.runInTx(ctx, func(txCtx context.Context) error {
		return s.repo.DeleteLinkedApp(txCtx, userID, appID)
	})
}

func (s *hubService) fetchAndPersistAppSummary(ctx context.Context, app *LinkedApp) (*AppSummaryResponse, error) {
	// Call external satellite app endpoint: GET {api_base_url}/api/v1/hub/summary
	summaryURL := fmt.Sprintf("%s/api/v1/hub/summary", strings.TrimRight(app.APIBaseURL, "/"))
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, summaryURL, nil)
	if err == nil {
		httpReq.Header.Set("Authorization", "Bearer "+app.ReadOnlyToken)
		httpReq.Header.Set("X-Hub-User-UID", app.ConnectedUID)
		httpReq.Header.Set("Accept", "application/json")

		resp, err := s.httpClient.Do(httpReq)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				bodyBytes, readErr := io.ReadAll(resp.Body)
				if readErr == nil {
					var envelope struct {
						Data *AppSummaryResponse `json:"data"`
					}
					var summary AppSummaryResponse
					if err := json.Unmarshal(bodyBytes, &envelope); err == nil && envelope.Data != nil && envelope.Data.AppID != "" {
						summary = *envelope.Data
					} else {
						_ = json.Unmarshal(bodyBytes, &summary)
					}

					if summary.AppID != "" {
						aiDataBytes, _ := json.Marshal(summary.AIContextData)
						_ = s.repo.UpdateLinkedAppSummary(
							ctx,
							app.HubUserID,
							app.AppID,
							summary.HighlightMetric,
							summary.HighlightLabel,
							string(aiDataBytes),
							time.Now(),
						)
						return &summary, nil
					}
				}
			}
		}
	}

	// Fallback to cached summary stored in database (avoiding synthetic fake numbers)
	var aiMap map[string]interface{}
	if app.AIContextData != "" && app.AIContextData != "{}" {
		_ = json.Unmarshal([]byte(app.AIContextData), &aiMap)
	}

	metric := app.HighlightMetric
	label := app.HighlightLabel
	if metric == "" {
		metric = "Terhubung"
	}
	if label == "" {
		label = "Menunggu Sinkronisasi Data"
	}
	deepLink := fmt.Sprintf("%sdashboard?expected_uid=%s", app.CustomURLScheme, app.ConnectedUID)

	return &AppSummaryResponse{
		AppID:           app.AppID,
		AppName:         app.AppName,
		ConnectedUID:    app.ConnectedUID,
		ConnectedEmail:  app.ConnectedEmail,
		HighlightMetric: metric,
		HighlightLabel:  label,
		AIContextData:   aiMap,
		DeepLink:        deepLink,
	}, nil
}

func (s *hubService) SyncAllSummaries(ctx context.Context, userID string) ([]AppSummaryResponse, error) {
	apps, err := s.repo.GetLinkedApps(ctx, userID)
	if err != nil {
		return nil, err
	}

	var results []AppSummaryResponse
	for i := range apps {
		summary, _ := s.fetchAndPersistAppSummary(ctx, &apps[i])
		if summary != nil {
			results = append(results, *summary)
		}
	}
	return results, nil
}

func (s *hubService) GetDailyBriefing(ctx context.Context, userID string) (*DailyBriefingResponse, error) {
	summaries, err := s.SyncAllSummaries(ctx, userID)
	if err != nil {
		return nil, err
	}

	var actions []SuggestedAction
	var fTrackerSummary *AppSummaryResponse
	var taskSummary *AppSummaryResponse

	for i := range summaries {
		s := &summaries[i]
		if strings.Contains(strings.ToLower(s.AppID), "tracker") {
			fTrackerSummary = s
			actions = append(actions, SuggestedAction{
				Title:       "Lihat Detail Pengeluaran Bulan Ini",
				Description: fmt.Sprintf("Total %s (%s)", s.HighlightMetric, s.HighlightLabel),
				AppID:       s.AppID,
				DeepLink:    s.DeepLink,
			})
		} else if strings.Contains(strings.ToLower(s.AppID), "task") {
			taskSummary = s
			actions = append(actions, SuggestedAction{
				Title:       "Buka Daftar Tugas Hari Ini",
				Description: fmt.Sprintf("%s menunggu penyelesaian", s.HighlightMetric),
				AppID:       s.AppID,
				DeepLink:    s.DeepLink,
			})
		}
	}

	// AI synthesis for Daily Briefing text
	briefingText := s.synthesizeDailyBriefing(summaries, fTrackerSummary, taskSummary)

	return &DailyBriefingResponse{
		GeneratedAt:      time.Now(),
		DailyAIBriefing:  briefingText,
		MiniSummaryCards: summaries,
		SuggestedActions: actions,
	}, nil
}

func (s *hubService) synthesizeDailyBriefing(summaries []AppSummaryResponse, finance, tasks *AppSummaryResponse) string {
	if len(summaries) == 0 {
		return "Selamat datang di Personal AI Hub! Belum ada aplikasi yang terhubung. Hubungkan F-Tracker dan Task Management untuk mendapatkan ringkasan harian otomatis."
	}

	var sb strings.Builder
	sb.WriteString("Halo! Berikut briefing harian untuk produktivitas & finansial Anda:\n\n")

	if finance != nil {
		sb.WriteString(fmt.Sprintf("💰 **Keuangan (F-Tracker):** Pengeluaran tercatat %s. ", finance.HighlightMetric))
		if topCat, ok := finance.AIContextData["top_category"].(string); ok {
			sb.WriteString(fmt.Sprintf("Kategori terbesar saat ini adalah %s. Budget masih dalam batas aman.", topCat))
		}
		sb.WriteString("\n")
	}

	if tasks != nil {
		sb.WriteString(fmt.Sprintf("📋 **Aktivitas (Task Management):** Anda memiliki %s. ", tasks.HighlightMetric))
		if highPri, ok := tasks.AIContextData["high_priority_task"].(string); ok {
			sb.WriteString(fmt.Sprintf("Fokus utama hari ini: \"%s\".", highPri))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n✨ Klik kartu atau tautan di bawah untuk membuka aplikasi terkait dengan akun yang tersinkronisasi.")
	return sb.String()
}

func (s *hubService) GetAvailableTools(ctx context.Context, userID string) (*HubToolsListResponse, error) {
	tools := GetHubGeminiTools()
	return &HubToolsListResponse{
		TotalTools: len(tools.FunctionDeclarations),
		Tools:      tools.FunctionDeclarations,
	}, nil
}

func (s *hubService) ExecuteTool(ctx context.Context, userID string, req *ExecuteToolRequest) (*ToolExecutionResult, error) {
	apps, err := s.repo.GetLinkedApps(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve linked apps: %w", err)
	}

	var taskApp *LinkedApp
	var trackerApp *LinkedApp
	for i := range apps {
		appLower := strings.ToLower(apps[i].AppID)
		if strings.Contains(appLower, "task") {
			taskApp = &apps[i]
		} else if strings.Contains(appLower, "tracker") {
			trackerApp = &apps[i]
		}
	}

	switch req.ToolName {
	case "get_tasks_summary":
		if taskApp == nil {
			return &ToolExecutionResult{
				ToolName: req.ToolName,
				Success:  false,
				Message:  "Aplikasi Task Management belum terhubung ke Hub. Silakan hubungkan di menu Connected Apps.",
			}, nil
		}
		summary, _ := s.fetchAndPersistAppSummary(ctx, taskApp)
		listArgs := map[string]interface{}{"limit": 10}
		if filter, ok := req.Arguments["filter"].(string); ok && filter != "" {
			switch strings.ToLower(filter) {
			case "today":
				listArgs["is_focus"] = true
			case "priority":
				listArgs["priority"] = "high"
			case "overdue":
				listArgs["status"] = "in_progress"
			default:
				listArgs["status"] = filter
			}
		} else {
			listArgs["is_focus"] = true
		}
		tasksRes, _ := s.executeTaskAppMCPTool(ctx, taskApp, "list_tasks", listArgs)
		combinedData := map[string]interface{}{
			"summary": summary,
		}
		if tasksRes != nil && tasksRes.Data != nil {
			combinedData["tasks"] = tasksRes.Data
		}
		return &ToolExecutionResult{
			ToolName: req.ToolName,
			Success:  true,
			Message:  "Ringkasan dan daftar tugas berhasil diambil dari Task Management",
			Data:     combinedData,
			Action: &SuggestedAction{
				Title:       "Buka Task Management",
				Description: "Lihat ringkasan tugas hari ini",
				AppID:       taskApp.AppID,
				DeepLink:    fmt.Sprintf("%stasks?expected_uid=%s", taskApp.CustomURLScheme, taskApp.ConnectedUID),
			},
		}, nil

	case "create_task", "list_tasks", "update_task_status":
		return s.executeTaskAppMCPTool(ctx, taskApp, req.ToolName, req.Arguments)

	case "create_transaction", "list_transactions", "get_financial_summary":
		return s.executeFTrackerTool(ctx, trackerApp, req.ToolName, req.Arguments)

	case "generate_cross_app_action":
		appID, _ := req.Arguments["app_id"].(string)
		targetRoute, _ := req.Arguments["target_route"].(string)
		actionTitle, _ := req.Arguments["action_title"].(string)
		queryParams, _ := req.Arguments["query_params"].(string)

		scheme := "ekahub://"
		expectedUID := ""
		if strings.Contains(strings.ToLower(appID), "track") && trackerApp != nil {
			scheme = trackerApp.CustomURLScheme
			expectedUID = trackerApp.ConnectedUID
		} else if strings.Contains(strings.ToLower(appID), "task") && taskApp != nil {
			scheme = taskApp.CustomURLScheme
			expectedUID = taskApp.ConnectedUID
		}

		link := fmt.Sprintf("%s%s?expected_uid=%s", scheme, strings.TrimPrefix(targetRoute, "/"), expectedUID)
		if queryParams != "" {
			link = fmt.Sprintf("%s&%s", link, queryParams)
		}

		action := &SuggestedAction{
			Title:       actionTitle,
			Description: fmt.Sprintf("Buka %s", appID),
			AppID:       appID,
			DeepLink:    link,
		}

		return &ToolExecutionResult{
			ToolName: req.ToolName,
			Success:  true,
			Message:  "Tautan aksi lintas aplikasi berhasil dibuat",
			Data:     map[string]string{"deep_link": link},
			Action:   action,
		}, nil

	default:
		return nil, fmt.Errorf("unknown tool name: %s", req.ToolName)
	}
}

func (s *hubService) executeTaskAppMCPTool(ctx context.Context, app *LinkedApp, toolName string, args map[string]interface{}) (*ToolExecutionResult, error) {
	if app == nil {
		return &ToolExecutionResult{
			ToolName: toolName,
			Success:  false,
			Message:  "Aplikasi Task Management belum terhubung ke Hub. Silakan hubungkan di menu Connected Apps.",
		}, nil
	}

	// Build JSON-RPC payload for Task Management MCP endpoint
	mcpReq := MCPJSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      toolName,
			"arguments": args,
		},
	}

	bodyBytes, err := json.Marshal(mcpReq)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/api/v1/mcp", strings.TrimRight(app.APIBaseURL, "/"))
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+app.ReadOnlyToken)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		// Resilience: Return informative status instead of crashing
		return &ToolExecutionResult{
			ToolName: toolName,
			Success:  false,
			Message:  fmt.Sprintf("Tidak dapat terhubung ke Task Management (%s): %v", endpoint, err),
		}, nil
	}
	defer resp.Body.Close()

	var rpcResp MCPJSONRPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return nil, fmt.Errorf("failed to parse MCP response: %w", err)
	}

	if rpcResp.Error != nil {
		return &ToolExecutionResult{
			ToolName: toolName,
			Success:  false,
			Message:  rpcResp.Error.Message,
		}, nil
	}

	var action *SuggestedAction
	msg := fmt.Sprintf("Aksi '%s' pada Task Management berhasil dieksekusi", toolName)

	if toolName == "create_task" {
		title, _ := args["title"].(string)
		msg = fmt.Sprintf("Tugas '%s' berhasil dibuat di Task Management", title)
		action = &SuggestedAction{
			Title:       "Lihat Tugas di Task Management",
			Description: title,
			AppID:       app.AppID,
			DeepLink:    fmt.Sprintf("%stasks?expected_uid=%s", app.CustomURLScheme, app.ConnectedUID),
		}
	} else if toolName == "list_tasks" {
		msg = "Daftar tugas berhasil diambil dari Task Management"
		action = &SuggestedAction{
			Title:       "Buka Task Management",
			Description: "Lihat semua tugas harian",
			AppID:       app.AppID,
			DeepLink:    fmt.Sprintf("%stasks?expected_uid=%s", app.CustomURLScheme, app.ConnectedUID),
		}
	}

	return &ToolExecutionResult{
		ToolName: toolName,
		Success:  true,
		Message:  msg,
		Data:     rpcResp.Result,
		Action:   action,
	}, nil
}

func (s *hubService) executeFTrackerTool(ctx context.Context, app *LinkedApp, toolName string, args map[string]interface{}) (*ToolExecutionResult, error) {
	if app == nil {
		return &ToolExecutionResult{
			ToolName: toolName,
			Success:  false,
			Message:  "Aplikasi F-Tracker belum terhubung ke Hub. Silakan hubungkan di menu Connected Apps.",
		}, nil
	}

	if toolName == "get_financial_summary" {
		summary, _ := s.fetchAndPersistAppSummary(ctx, app)
		transRes, _ := s.executeFTrackerTool(ctx, app, "list_transactions", map[string]interface{}{"limit": 5})
		dataMap := map[string]interface{}{
			"summary": summary,
		}
		if transRes != nil && transRes.Data != nil {
			dataMap["recent_transactions"] = transRes.Data
		}
		return &ToolExecutionResult{
			ToolName: toolName,
			Success:  true,
			Message:  "Ringkasan keuangan berhasil diambil dari F-Tracker",
			Data:     dataMap,
			Action: &SuggestedAction{
				Title:       "Buka Laporan Keuangan",
				Description: summary.HighlightMetric,
				AppID:       app.AppID,
				DeepLink:    summary.DeepLink,
			},
		}, nil
	}

	// Try calling F-Tracker MCP endpoint first, fallback to REST /api/v1/transactions
	mcpEndpoint := fmt.Sprintf("%s/api/v1/hub/mcp", strings.TrimRight(app.APIBaseURL, "/"))
	mcpReq := MCPJSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      toolName,
			"arguments": args,
		},
	}
	bodyBytes, _ := json.Marshal(mcpReq)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, mcpEndpoint, bytes.NewReader(bodyBytes))
	if err == nil {
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+app.ReadOnlyToken)
		httpReq.Header.Set("X-Hub-User-UID", app.ConnectedUID)
		httpReq.Header.Set("X-Timezone", "Asia/Jakarta")
		resp, err := s.httpClient.Do(httpReq)
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			var rpcResp MCPJSONRPCResponse
			if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err == nil {
				if rpcResp.Error != nil {
					return &ToolExecutionResult{
						ToolName: toolName,
						Success:  false,
						Message:  fmt.Sprintf("Gagal mengeksekusi pada F-Tracker: %s", rpcResp.Error.Message),
					}, nil
				}

				if resMap, ok := rpcResp.Result.(map[string]interface{}); ok {
					if isErr, _ := resMap["isError"].(bool); isErr {
						errMsg := "Gagal mengeksekusi aksi pada F-Tracker"
						if contentList, ok := resMap["content"].([]interface{}); ok && len(contentList) > 0 {
							if firstContent, ok := contentList[0].(map[string]interface{}); ok {
								if txt, ok := firstContent["text"].(string); ok && txt != "" {
									errMsg = txt
								}
							}
						}
						return &ToolExecutionResult{
							ToolName: toolName,
							Success:  false,
							Message:  errMsg,
							Data:     resMap,
						}, nil
					}
				}

				action := &SuggestedAction{
					Title:       "Lihat Transaksi di F-Tracker",
					Description: "Buka daftar transaksi",
					AppID:       app.AppID,
					DeepLink:    fmt.Sprintf("%sexpenses?expected_uid=%s", app.CustomURLScheme, app.ConnectedUID),
				}

				successMsg := fmt.Sprintf("Aksi '%s' pada F-Tracker berhasil", toolName)
				if toolName == "create_transaction" {
					desc, _ := args["description"].(string)
					var amt float64
					switch v := args["amount"].(type) {
					case float64:
						amt = v
					case int:
						amt = float64(v)
					case json.Number:
						amt, _ = v.Float64()
					}
					if amt > 0 {
						successMsg = fmt.Sprintf("Transaksi '%s' sebesar Rp %.0f berhasil dicatat di F-Tracker", desc, amt)
					} else {
						successMsg = fmt.Sprintf("Transaksi '%s' berhasil dicatat di F-Tracker", desc)
					}
					action = &SuggestedAction{
						Title:       "Lihat Pengeluaran di F-Tracker",
						Description: desc,
						AppID:       app.AppID,
						DeepLink:    fmt.Sprintf("%sexpenses?expected_uid=%s", app.CustomURLScheme, app.ConnectedUID),
					}
				}

				return &ToolExecutionResult{
					ToolName: toolName,
					Success:  true,
					Message:  successMsg,
					Data:     rpcResp.Result,
					Action:   action,
				}, nil
			}
		}
	}

	// Fallback to direct REST endpoint: POST /api/v1/transactions
	if toolName == "create_transaction" {
		restEndpoint := fmt.Sprintf("%s/api/v1/transactions", strings.TrimRight(app.APIBaseURL, "/"))
		restBytes, _ := json.Marshal(args)
		restReq, err := http.NewRequestWithContext(ctx, http.MethodPost, restEndpoint, bytes.NewReader(restBytes))
		if err != nil {
			return nil, err
		}
		restReq.Header.Set("Content-Type", "application/json")
		restReq.Header.Set("Authorization", "Bearer "+app.ReadOnlyToken)
		restReq.Header.Set("X-Hub-User-UID", app.ConnectedUID)
		restReq.Header.Set("X-Timezone", "Asia/Jakarta")
		resp, err := s.httpClient.Do(restReq)
		if err != nil {
			return &ToolExecutionResult{
				ToolName: toolName,
				Success:  false,
				Message:  fmt.Sprintf("Tidak dapat terhubung ke F-Tracker (%s): %v", restEndpoint, err),
			}, nil
		}
		defer resp.Body.Close()

		var resBody map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&resBody)

		isSuccess := resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated
		desc, _ := args["description"].(string)

		var msg string
		if isSuccess {
			var amt float64
			switch v := args["amount"].(type) {
			case float64:
				amt = v
			case int:
				amt = float64(v)
			}
			if amt > 0 {
				msg = fmt.Sprintf("Transaksi '%s' sebesar Rp %.0f berhasil dicatat di F-Tracker", desc, amt)
			} else {
				msg = fmt.Sprintf("Transaksi '%s' berhasil dicatat di F-Tracker", desc)
			}
		} else {
			errMsg, _ := resBody["message"].(string)
			if errMsg == "" {
				errMsg = fmt.Sprintf("HTTP %d", resp.StatusCode)
			}
			msg = fmt.Sprintf("Gagal mencatat transaksi di F-Tracker: %s", errMsg)
		}

		return &ToolExecutionResult{
			ToolName: toolName,
			Success:  isSuccess,
			Message:  msg,
			Data:     resBody,
			Action: &SuggestedAction{
				Title:       "Lihat Pengeluaran di F-Tracker",
				Description: desc,
				AppID:       app.AppID,
				DeepLink:    fmt.Sprintf("%sexpenses?expected_uid=%s", app.CustomURLScheme, app.ConnectedUID),
			},
		}, nil
	}

	return &ToolExecutionResult{
		ToolName: toolName,
		Success:  true,
		Message:  fmt.Sprintf("Aksi %s diproses", toolName),
		Action: &SuggestedAction{
			Title:       "Buka F-Tracker",
			Description: "Lihat keuangan Anda",
			AppID:       app.AppID,
			DeepLink:    fmt.Sprintf("%sdashboard?expected_uid=%s", app.CustomURLScheme, app.ConnectedUID),
		},
	}, nil
}

func (s *hubService) ProcessAIChat(ctx context.Context, userID, userMessage, sessionID string) (*AIChatResponse, error) {
	summaries, _ := s.SyncAllSummaries(ctx, userID)

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, errors.New("GEMINI_API_KEY is not configured on the server")
	}

	// Session management: create new session if none provided
	var activeSessionID *string
	if sessionID == "" {
		// Generate a smart title from the first ~60 chars of the prompt
		title := userMessage
		if len(title) > 60 {
			title = title[:60] + "..."
		}
		session, err := s.repo.CreateChatSession(ctx, userID, title)
		if err != nil {
			return nil, fmt.Errorf("failed to create chat session: %w", err)
		}
		activeSessionID = &session.ID
	} else {
		// Validate session belongs to user
		_, err := s.repo.GetChatSessionByID(ctx, userID, sessionID)
		if err != nil {
			return nil, fmt.Errorf("invalid session: %w", err)
		}
		activeSessionID = &sessionID
	}

	// Load session-scoped history (up to 30 messages for deeper context)
	var history []StoredChatMessage
	if activeSessionID != nil {
		history, _ = s.repo.GetSessionMessages(ctx, *activeSessionID, 30)
	}

	res, err := s.callGeminiWithTools(ctx, apiKey, userID, userMessage, summaries, history)
	if err != nil {
		return nil, fmt.Errorf("AI processing error: %w", err)
	}

	// Persist messages with session association
	_ = s.repo.SaveChatMessage(ctx, userID, "user", userMessage, activeSessionID, nil, nil)
	_ = s.repo.SaveChatMessage(ctx, userID, "assistant", res.Reply, activeSessionID, res.ToolCalls, nil)

	// Touch session to update its updated_at timestamp
	if activeSessionID != nil {
		_ = s.repo.TouchChatSession(ctx, *activeSessionID)
	}

	res.SessionID = *activeSessionID
	return res, nil
}

// ---- Chat Session Service Methods ----

func (s *hubService) GetChatSessions(ctx context.Context, userID string, limit, offset int) ([]ChatSession, error) {
	return s.repo.GetChatSessions(ctx, userID, limit, offset)
}

func (s *hubService) GetChatSessionDetail(ctx context.Context, userID, sessionID string) (*ChatSessionDetail, error) {
	session, err := s.repo.GetChatSessionByID(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}
	messages, err := s.repo.GetSessionMessages(ctx, sessionID, 100)
	if err != nil {
		return nil, err
	}
	return &ChatSessionDetail{
		ChatSession: *session,
		Messages:    messages,
	}, nil
}

func (s *hubService) DeleteChatSession(ctx context.Context, userID, sessionID string) error {
	return s.repo.DeleteChatSession(ctx, userID, sessionID)
}

func (s *hubService) executeGeminiRequestWithRetry(ctx context.Context, apiKey, preferredModel string, payload map[string]interface{}) ([]byte, error) {
	baseURL := os.Getenv("GEMINI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com"
	}

	modelsToTry := []string{preferredModel}
	if preferredModel != "gemini-2.5-flash" {
		modelsToTry = append(modelsToTry, "gemini-2.5-flash")
	}
	if preferredModel != "gemini-2.0-flash" {
		modelsToTry = append(modelsToTry, "gemini-2.0-flash")
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	var lastErr error
	for _, m := range modelsToTry {
		url := fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s", strings.TrimRight(baseURL, "/"), m, apiKey)

		for attempt := 0; attempt < 3; attempt++ {
			if attempt > 0 {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(time.Duration(400*(1<<attempt)) * time.Millisecond):
				}
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
			if err != nil {
				lastErr = err
				continue
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := s.httpClient.Do(req)
			if err != nil {
				lastErr = err
				continue
			}

			respBody, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				return respBody, nil
			}

			lastErr = fmt.Errorf("gemini HTTP %d: %s", resp.StatusCode, string(respBody))
			// Retry on rate limit (429) or server errors (5xx)
			if resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode < 500 {
				break
			}
		}
	}

	return nil, lastErr
}

func buildGeminiConversationContents(history []StoredChatMessage, currentPrompt string) []map[string]interface{} {
	var contents []map[string]interface{}

	for _, msg := range history {
		text := strings.TrimSpace(msg.Content)
		if text == "" {
			continue
		}

		role := "user"
		if msg.Role == "assistant" || msg.Role == "model" {
			role = "model"
		}

		// Ensure the first turn in conversation is "user"
		if len(contents) == 0 && role != "user" {
			continue
		}

		// If consecutive turns have the same role, combine them
		if len(contents) > 0 && contents[len(contents)-1]["role"] == role {
			prevParts, ok := contents[len(contents)-1]["parts"].([]map[string]interface{})
			if ok && len(prevParts) > 0 {
				prevText, _ := prevParts[0]["text"].(string)
				contents[len(contents)-1]["parts"] = []map[string]interface{}{
					{"text": prevText + "\n" + text},
				}
				continue
			}
		}

		contents = append(contents, map[string]interface{}{
			"role": role,
			"parts": []map[string]interface{}{
				{"text": text},
			},
		})
	}

	cleanPrompt := strings.TrimSpace(currentPrompt)
	if len(contents) > 0 && contents[len(contents)-1]["role"] == "user" {
		prevParts, ok := contents[len(contents)-1]["parts"].([]map[string]interface{})
		if ok && len(prevParts) > 0 {
			prevText, _ := prevParts[0]["text"].(string)
			contents[len(contents)-1]["parts"] = []map[string]interface{}{
				{"text": prevText + "\n" + cleanPrompt},
			}
			return contents
		}
	}

	contents = append(contents, map[string]interface{}{
		"role": "user",
		"parts": []map[string]interface{}{
			{"text": cleanPrompt},
		},
	})

	return contents
}

// callGeminiWithTools implements multi-turn parallel function calling with Google Gemini v1beta endpoint.
func (s *hubService) callGeminiWithTools(ctx context.Context, apiKey, userID, userPrompt string, summaries []AppSummaryResponse, history []StoredChatMessage) (*AIChatResponse, error) {
	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-2.5-flash"
	}

	linkedContext := "Belum ada aplikasi satelit (F-Tracker atau Task Management) yang terhubung."
	if len(summaries) > 0 {
		summariesJSON, _ := json.Marshal(summaries)
		linkedContext = string(summariesJSON)
	}

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.FixedZone("WIB", 7*3600)
	}
	nowLocal := time.Now().In(loc)
	currentTimeContext := fmt.Sprintf(
		"Waktu & Tanggal Pengguna Saat Ini: %s, %s (Timezone: Asia/Jakarta, WIB, UTC+7).",
		nowLocal.Format("Monday, 02 January 2006 15:04:05"),
		nowLocal.Format("2006-01-02"),
	)

	systemPrompt := fmt.Sprintf(`You are Eka-Hub AI, a personal AI life aggregator assistant. You help users both as a smart, concise everyday assistant AND by orchestrating their connected apps (F-Tracker for finances/transactions and Task Management for tasks/todos). Respond in the user's language (Indonesian by default if the prompt is Indonesian, otherwise English) using clean, readable Markdown.

%s

### MANDATORY CONFIRMATION PROTOCOL (DRAFT & CONFIRM) FOR MUTATING ACTIONS:
IMPORTANT: NEVER immediately or blindly execute data-creating/mutating actions ('create_transaction' in F-Tracker or 'create_task' in Task Management) on initial user requests without explicit confirmation!
Users MUST always be shown a clear draft confirmation first before any new transaction or task is created in external apps.

1. INITIAL TRANSACTION REQUEST (e.g. 'catat pengeluaran kopi 35000', 'beli makan 25000', 'catat bensin 50000', 'tambah transaksi'):
- DO NOT invoke 'create_transaction' yet!
- Check existing categories from Connected Apps Live Context below. Match against existing user categories with semantic intelligence (e.g. if the user has 'Food and Drink' or 'Food & Drink', use that; DO NOT invent 'Makanan & Minuman'; if user has 'Transport', use 'Transport').
- Formulate a clear, structured draft proposal and ask the user to confirm:
  ### 📝 Konfirmasi Pencatatan Transaksi
  Saya telah menyiapkan rincian transaksi untuk dicatat ke **F-Tracker**:
  - **Aplikasi**: F-Tracker
  - **Jenis**: Pengeluaran (*Expense*) atau Pemasukan (*Income*)
  - **Nominal**: Rp [nominal terformat, contoh: 25.000]
  - **Keterangan**: [deskripsi transaksi]
  - **Kategori Rekomendasi**: [nama kategori yang cocok dari F-Tracker]
  - **Tanggal**: [tanggal hari ini %s / yang diminta]

  Apakah rincian dan kategorinya sudah sesuai? Balas **"Ya, catat"** untuk menyimpan, atau beri tahu jika ingin mengganti kategori atau nominalnya.

2. INITIAL TASK REQUEST (e.g. 'buat task belajar go', 'tambah todo beli susu', 'tolong bikin tugas meeting'):
- DO NOT invoke 'create_task' yet!
- Formulate a clear, structured draft proposal and ask the user to confirm:
  ### 📋 Konfirmasi Pembuatan Tugas
  Saya telah menyiapkan rincian tugas untuk dicatat ke **Task Management**:
  - **Aplikasi**: Task Management
  - **Judul Tugas**: [judul tugas]
  - **Prioritas**: [prioritas: low / medium / high / urgent]
  - **Tenggat Waktu**: [deadline jika disebutkan atau '-']

  Apakah tugas ini sudah sesuai? Balas **"Ya, buatkan"** untuk menyimpan, atau beri tahu jika ada detail yang ingin diubah.

3. EXECUTION CONDITIONS (WHEN TO ACTUALLY INVOKE MUTATING TOOLS):
- You MUST ONLY invoke 'create_transaction' or 'create_task' when:
  a) The user explicitly CONFIRMS a previously proposed draft in the conversation history (e.g. user says: 'ya', 'catat', 'oke', 'buatkan', 'simpan', 'lanjutkan', 'eksekusi', 'yes', 'confirm', 'gas', 'sudah sesuai', 'setuju', 'silakan'), OR
  b) The user explicitly specifies direct creation in their initial prompt (e.g. 'langsung catat sekarang tanpa konfirmasi', 'force create').
- When executing upon confirmation, use the exact agreed parameters and notify the user with the result and action link.

4. CATEGORY & PROJECT INTELLIGENCE:
- NEVER invent or create unnecessary new categories if a semantically equivalent category exists in F-Tracker.
- Understand bilingual equivalence:
  * "Food and Drink" / "Food & Drink" == "Makanan & Minuman" == "Makanan dan Minuman" == "Food" == "Cafe / Restaurant / Kopi"
  * "Transport" / "Transportation" == "Transportasi" == "Bensin" == "Ojol" == "Fuel"
  * "Shopping" == "Belanja" == "Groceries" == "Supermarket"
  * "Bills & Utilities" / "Utilities" == "Tagihan & Utilitas" == "Listrik" == "Pulsa" == "Internet"
  * "Salary" == "Gaji & Pendapatan" == "Gaji" == "Pendapatan"
  * "Investment" == "Investasi & Tabungan" == "Investasi" == "Tabungan"
  * "Bonus & Other" == "Bonus & Lain-lain"
- If the user specifies or already has an English category name like "Food and Drink", use "Food and Drink"! Do NOT convert it to "Makanan & Minuman".

5. READ-ONLY TOOLS (IMMEDIATE EXECUTION):
- Read-only tools ('get_financial_summary', 'list_transactions', 'get_tasks_summary', 'list_tasks') DO NOT mutate user data.
- You MUST invoke read-only tools IMMEDIATELY when the user asks for summaries, spending, budgets, balances, or task lists.

6. SATELLITE APP STATUS:
- If a requested satellite app is not linked yet, inform them politely that they can connect F-Tracker or Task Management in the Satellites tab.

Connected Apps Live Context: %s`, currentTimeContext, nowLocal.Format("2006-01-02"), linkedContext)

	tools := GetHubGeminiTools()

	// Multi-Turn Conversation History + Current User Request
	contents := buildGeminiConversationContents(history, userPrompt)

	systemInstruction := map[string]interface{}{
		"parts": []map[string]interface{}{
			{"text": systemPrompt},
		},
	}

	payload := map[string]interface{}{
		"systemInstruction": systemInstruction,
		"contents":          contents,
		"tools":             []interface{}{tools},
	}

	respBody, err := s.executeGeminiRequestWithRetry(ctx, apiKey, model, payload)
	if err != nil {
		return nil, fmt.Errorf("AI processing error: %w", err)
	}

	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Role  string `json:"role"`
				Parts []struct {
					Text         string `json:"text"`
					FunctionCall *struct {
						Name string                 `json:"name"`
						Args map[string]interface{} `json:"args"`
					} `json:"functionCall"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return nil, err
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from gemini")
	}

	candidate := geminiResp.Candidates[0]

	// Collect ALL function calls returned by the model (supporting parallel / multi-tool calls)
	type fnCallInfo struct {
		Name string
		Args map[string]interface{}
	}
	var fnCalls []fnCallInfo
	for _, part := range candidate.Content.Parts {
		if part.FunctionCall != nil {
			fnCalls = append(fnCalls, fnCallInfo{
				Name: part.FunctionCall.Name,
				Args: part.FunctionCall.Args,
			})
		}
	}

	// Check if Gemini wants to call tools
	if len(fnCalls) > 0 {
		var toolCalls []ToolCallDetail
		var actionLinks []SuggestedAction
		var modelParts []interface{}
		var fnResponses []interface{}

		for _, fnCall := range fnCalls {
			toolResult, execErr := s.ExecuteTool(ctx, userID, &ExecuteToolRequest{
				ToolName:  fnCall.Name,
				Arguments: fnCall.Args,
			})

			toolCalls = append(toolCalls, ToolCallDetail{
				Name:      fnCall.Name,
				Arguments: fnCall.Args,
				Result:    toolResult,
			})

			if toolResult != nil && toolResult.Action != nil {
				actionLinks = append(actionLinks, *toolResult.Action)
			}

			responseVal := map[string]interface{}{"status": "success"}
			if execErr != nil {
				responseVal = map[string]interface{}{"status": "error", "error": execErr.Error()}
			} else if toolResult != nil {
				responseVal = map[string]interface{}{
					"status":  "success",
					"message": toolResult.Message,
					"data":    toolResult.Data,
				}
			}

			modelParts = append(modelParts, map[string]interface{}{
				"functionCall": map[string]interface{}{
					"name": fnCall.Name,
					"args": fnCall.Args,
				},
			})

			fnResponses = append(fnResponses, map[string]interface{}{
				"functionResponse": map[string]interface{}{
					"name":     fnCall.Name,
					"response": responseVal,
				},
			})
		}

		// Turn 2: Feed all function responses back to Gemini for natural language synthesis
		turn2Contents := append(contents,
			map[string]interface{}{
				"role":  "model",
				"parts": modelParts,
			},
			map[string]interface{}{
				"role":  "function",
				"parts": fnResponses,
			},
		)

		turn2Payload := map[string]interface{}{
			"systemInstruction": systemInstruction,
			"contents":          turn2Contents,
		}

		turn2RespBody, turn2Err := s.executeGeminiRequestWithRetry(ctx, apiKey, model, turn2Payload)
		if turn2Err == nil {
			var finalGeminiResp struct {
				Candidates []struct {
					Content struct {
						Parts []struct {
							Text string `json:"text"`
						} `json:"parts"`
					} `json:"content"`
				} `json:"candidates"`
			}
			if err := json.Unmarshal(turn2RespBody, &finalGeminiResp); err == nil {
				if len(finalGeminiResp.Candidates) > 0 {
					var replyBuilder strings.Builder
					for _, part := range finalGeminiResp.Candidates[0].Content.Parts {
						if part.Text != "" {
							replyBuilder.WriteString(part.Text)
						}
					}
					finalReply := strings.TrimSpace(replyBuilder.String())
					if finalReply != "" {
						return &AIChatResponse{
							Reply:       finalReply,
							ToolCalls:   toolCalls,
							ActionLinks: actionLinks,
						}, nil
					}
				}
			}
		}

		// Graceful synthesis fallback if Turn 2 fails or is empty
		var sb strings.Builder
		sb.WriteString("Berikut ringkasan data dari aplikasi terhubung Anda:\n\n")
		for _, tc := range toolCalls {
			if res, ok := tc.Result.(*ToolExecutionResult); ok && res != nil {
				sb.WriteString(fmt.Sprintf("• **%s**: %s\n", tc.Name, res.Message))
			}
		}
		return &AIChatResponse{
			Reply:       sb.String(),
			ToolCalls:   toolCalls,
			ActionLinks: actionLinks,
		}, nil
	}

	var replyBuilder strings.Builder
	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			replyBuilder.WriteString(part.Text)
		}
	}

	return &AIChatResponse{
		Reply: strings.TrimSpace(replyBuilder.String()),
	}, nil
}
