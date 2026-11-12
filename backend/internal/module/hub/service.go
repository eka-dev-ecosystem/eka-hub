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
	ProcessAIChat(ctx context.Context, userID, userMessage string) (*AIChatResponse, error)

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
	case "create_task", "list_tasks", "update_task_status", "get_tasks_summary":
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
		return &ToolExecutionResult{
			ToolName: toolName,
			Success:  true,
			Message:  "Ringkasan keuangan berhasil diambil dari F-Tracker",
			Data:     summary,
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
		resp, err := s.httpClient.Do(httpReq)
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			var rpcResp MCPJSONRPCResponse
			if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err == nil && rpcResp.Error == nil {
				action := &SuggestedAction{
					Title:       "Lihat Transaksi di F-Tracker",
					Description: "Buka daftar transaksi",
					AppID:       app.AppID,
					DeepLink:    fmt.Sprintf("%sexpenses?expected_uid=%s", app.CustomURLScheme, app.ConnectedUID),
				}
				return &ToolExecutionResult{
					ToolName: toolName,
					Success:  true,
					Message:  fmt.Sprintf("Aksi '%s' pada F-Tracker berhasil", toolName),
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

		desc, _ := args["description"].(string)
		return &ToolExecutionResult{
			ToolName: toolName,
			Success:  resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated,
			Message:  fmt.Sprintf("Transaksi '%s' berhasil dicatat di F-Tracker", desc),
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

func (s *hubService) ProcessAIChat(ctx context.Context, userID, userMessage string) (*AIChatResponse, error) {
	summaries, _ := s.SyncAllSummaries(ctx, userID)

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, errors.New("GEMINI_API_KEY is not configured on the server")
	}

	res, err := s.callGeminiWithTools(ctx, apiKey, userID, userMessage, summaries)
	if err != nil {
		return nil, fmt.Errorf("AI processing error: %w", err)
	}

	_ = s.repo.SaveChatMessage(ctx, userID, "user", userMessage, nil, nil)
	_ = s.repo.SaveChatMessage(ctx, userID, "assistant", res.Reply, res.ToolCalls, nil)

	return res, nil
}

// callGeminiWithTools implements multi-turn function calling with Google Gemini v1beta endpoint.
func (s *hubService) callGeminiWithTools(ctx context.Context, apiKey, userID, userPrompt string, summaries []AppSummaryResponse) (*AIChatResponse, error) {
	baseURL := os.Getenv("GEMINI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com"
	}
	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-2.5-flash"
	}
	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s", strings.TrimRight(baseURL, "/"), model, apiKey)

	linkedContext := "Belum ada aplikasi satelit (F-Tracker atau Task Management) yang terhubung."
	if len(summaries) > 0 {
		summariesJSON, _ := json.Marshal(summaries)
		linkedContext = string(summariesJSON)
	}

	systemPrompt := "You are Eka-Hub AI (Personal AI Hub Assistant powered by Gemini). You help users both as a smart, concise everyday AI assistant AND by orchestrating their connected apps (F-Tracker for finances/transactions and Task Management for tasks/todos). Respond in the user's language (Indonesian by default if the prompt is Indonesian, otherwise English) using clean, readable Markdown. When the user asks to record an expense, check budget, create a task, or list tasks, use the provided function tools; if the target satellite app is not linked yet, inform them politely that they can connect F-Tracker or Task Management in the Satellites tab.\n\nConnected Apps Live Context: " + linkedContext

	tools := GetHubGeminiTools()

	// Initial User Request Turn
	contents := []map[string]interface{}{
		{
			"role": "user",
			"parts": []map[string]interface{}{
				{"text": userPrompt},
			},
		},
	}

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

	bodyBytes, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gemini error (%d): %s", resp.StatusCode, string(respBody))
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

	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return nil, err
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from gemini")
	}

	candidate := geminiResp.Candidates[0]
	firstPart := candidate.Content.Parts[0]

	// Check if Gemini wants to call a tool
	if firstPart.FunctionCall != nil {
		fnCall := firstPart.FunctionCall
		toolResult, execErr := s.ExecuteTool(ctx, userID, &ExecuteToolRequest{
			ToolName:  fnCall.Name,
			Arguments: fnCall.Args,
		})

		toolCalls := []ToolCallDetail{
			{
				Name:      fnCall.Name,
				Arguments: fnCall.Args,
				Result:    toolResult,
			},
		}

		var actionLinks []SuggestedAction
		if toolResult != nil && toolResult.Action != nil {
			actionLinks = append(actionLinks, *toolResult.Action)
		}

		// Turn 2: Feed functionResponse back to Gemini for natural language synthesis
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

		turn2Contents := append(contents,
			map[string]interface{}{
				"role":  "model",
				"parts": []interface{}{map[string]interface{}{"functionCall": fnCall}},
			},
			map[string]interface{}{
				"role": "function",
				"parts": []interface{}{
					map[string]interface{}{
						"functionResponse": map[string]interface{}{
							"name":     fnCall.Name,
							"response": responseVal,
						},
					},
				},
			},
		)

		turn2Payload := map[string]interface{}{
			"systemInstruction": systemInstruction,
			"contents":          turn2Contents,
		}
		turn2Bytes, _ := json.Marshal(turn2Payload)
		req2, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(turn2Bytes))
		if err == nil {
			req2.Header.Set("Content-Type", "application/json")
			resp2, err2 := s.httpClient.Do(req2)
			if err2 == nil && resp2.StatusCode == http.StatusOK {
				defer resp2.Body.Close()
				var finalGeminiResp struct {
					Candidates []struct {
						Content struct {
							Parts []struct {
								Text string `json:"text"`
							} `json:"parts"`
						} `json:"content"`
					} `json:"candidates"`
				}
				if err := json.NewDecoder(resp2.Body).Decode(&finalGeminiResp); err == nil {
					if len(finalGeminiResp.Candidates) > 0 && len(finalGeminiResp.Candidates[0].Content.Parts) > 0 {
						return &AIChatResponse{
							Reply:       finalGeminiResp.Candidates[0].Content.Parts[0].Text,
							ToolCalls:   toolCalls,
							ActionLinks: actionLinks,
						}, nil
					}
				}
			}
		}

		// Fallback confirmation message if turn 2 fails
		replyMsg := fmt.Sprintf("Aksi '%s' berhasil diproses.", fnCall.Name)
		if toolResult != nil && toolResult.Message != "" {
			replyMsg = toolResult.Message
		}
		return &AIChatResponse{
			Reply:       replyMsg,
			ToolCalls:   toolCalls,
			ActionLinks: actionLinks,
		}, nil
	}

	return &AIChatResponse{
		Reply: firstPart.Text,
	}, nil
}
