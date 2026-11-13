package hub

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

type mockHubRepo struct {
	sessions   map[string]*HubLinkSession
	linkedApps map[string]*LinkedApp
}

func newMockHubRepo() *mockHubRepo {
	return &mockHubRepo{
		sessions:   make(map[string]*HubLinkSession),
		linkedApps: make(map[string]*LinkedApp),
	}
}

type mockTransactor struct{}

func (m *mockTransactor) RunInTx(ctx context.Context, fn func(txCtx context.Context) error) error {
	return fn(ctx)
}

func (m *mockTransactor) WithTransactionContext(ctx context.Context, fn func(txCtx context.Context) error) error {
	return m.RunInTx(ctx, fn)
}

func (m *mockHubRepo) CreateSession(ctx context.Context, session *HubLinkSession) error {
	m.sessions[session.SessionCode] = session
	return nil
}

func (m *mockHubRepo) GetSession(ctx context.Context, code string) (*HubLinkSession, error) {
	s, ok := m.sessions[code]
	if !ok {
		return nil, ErrSessionNotFound
	}
	return s, nil
}

func (m *mockHubRepo) MarkSessionExchanged(ctx context.Context, code string) error {
	if s, ok := m.sessions[code]; ok {
		s.Status = "EXCHANGED"
	}
	return nil
}

func (m *mockHubRepo) UpsertLinkedApp(ctx context.Context, app *LinkedApp) error {
	m.linkedApps[app.AppID] = app
	return nil
}

func (m *mockHubRepo) GetLinkedApps(ctx context.Context, userID string) ([]LinkedApp, error) {
	var list []LinkedApp
	for _, app := range m.linkedApps {
		list = append(list, *app)
	}
	return list, nil
}

func (m *mockHubRepo) GetLinkedApp(ctx context.Context, userID, appID string) (*LinkedApp, error) {
	app, ok := m.linkedApps[appID]
	if !ok {
		return nil, ErrAppNotLinked
	}
	return app, nil
}

func (m *mockHubRepo) UpdateLinkedAppSummary(ctx context.Context, userID, appID string, metric, label string, aiContext string, syncedAt time.Time) error {
	if app, ok := m.linkedApps[appID]; ok {
		app.HighlightMetric = metric
		app.HighlightLabel = label
		app.AIContextData = aiContext
		app.LastSyncedAt = &syncedAt
	}
	return nil
}

func (m *mockHubRepo) UpdateLinkedAccount(ctx context.Context, userID, appID, connectedUID, connectedEmail, token string) error {
	if app, ok := m.linkedApps[appID]; ok {
		app.ConnectedUID = connectedUID
		app.ConnectedEmail = connectedEmail
		if token != "" {
			app.ReadOnlyToken = token
		}
	}
	return nil
}

func (m *mockHubRepo) DeleteLinkedApp(ctx context.Context, userID, appID string) error {
	delete(m.linkedApps, appID)
	return nil
}

func (m *mockHubRepo) SaveChatMessage(ctx context.Context, userID, role, content string, toolCalls, toolResponses interface{}) error {
	return nil
}

func (m *mockHubRepo) GetRecentChatMessages(ctx context.Context, userID string, limit int) ([]StoredChatMessage, error) {
	return []StoredChatMessage{
		{ID: "1", Role: "user", Content: "Siapa nama saya?"},
		{ID: "2", Role: "assistant", Content: "Nama Anda Eka."},
	}, nil
}

func TestInitiateLink(t *testing.T) {
	repo := newMockHubRepo()
	service := NewHubService(repo, &mockTransactor{})

	resp, err := service.InitiateLink(context.Background(), "user-uuid-1", "ftracker")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.SessionCode == "" {
		t.Errorf("expected session code to be generated")
	}

	if resp.DeepLinkURL == "" {
		t.Errorf("expected deep link url to be generated")
	}

	// Verify session exists in repository
	session, err := repo.GetSession(context.Background(), resp.SessionCode)
	if err != nil || session.AppID != "ftracker" {
		t.Errorf("session was not saved correctly: %v", err)
	}
}

func TestExchangeLink(t *testing.T) {
	repo := newMockHubRepo()
	service := NewHubService(repo, &mockTransactor{})

	initResp, _ := service.InitiateLink(context.Background(), "user-uuid-1", "ftracker")

	req := &ExchangeLinkRequest{
		SessionCode:     initResp.SessionCode,
		AppID:           "ftracker",
		AppName:         "F-Tracker",
		ConnectedUID:    "usr_f123",
		ConnectedEmail:  "test@ftracker.com",
		APIBaseURL:      "https://api.f-tracker.com",
		ReadOnlyToken:   "token_abc",
		CustomURLScheme: "ftracker://",
	}

	if err := service.ExchangeLink(context.Background(), req); err != nil {
		t.Fatalf("exchange failed: %v", err)
	}

	app, err := repo.GetLinkedApp(context.Background(), "user-uuid-1", "ftracker")
	if err != nil {
		t.Fatalf("app was not linked: %v", err)
	}

	if app.ConnectedUID != "usr_f123" {
		t.Errorf("expected usr_f123, got %s", app.ConnectedUID)
	}
}

func TestStandardSummaryContractJSON(t *testing.T) {
	summary := AppSummaryResponse{
		AppID:           "ftracker",
		AppName:         "F-Tracker",
		ConnectedUID:    "usr_123",
		ConnectedEmail:  "user@domain.com",
		HighlightMetric: "Rp 1.500.000",
		HighlightLabel:  "Total Pengeluaran",
		AIContextData: map[string]interface{}{
			"spent": 1500000,
		},
		DeepLink: "ftracker://dashboard?expected_uid=usr_123",
	}

	bytes, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("failed to marshal summary: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(bytes, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	requiredFields := []string{
		"app_id", "app_name", "connected_uid", "connected_email",
		"highlight_metric", "highlight_label", "ai_context_data", "deep_link",
	}

	for _, field := range requiredFields {
		if _, exists := parsed[field]; !exists {
			t.Errorf("missing required field in standard contract: %s", field)
		}
	}
}

func TestGeminiToolsSchema(t *testing.T) {
	tools := GetHubGeminiTools()
	if len(tools.FunctionDeclarations) < 3 {
		t.Errorf("expected at least 3 tools, got %d", len(tools.FunctionDeclarations))
	}

	var foundFinance, foundTasks, foundAction bool
	for _, f := range tools.FunctionDeclarations {
		if f.Name == "get_financial_summary" {
			foundFinance = true
		}
		if f.Name == "get_tasks_summary" {
			foundTasks = true
		}
		if f.Name == "generate_cross_app_action" {
			foundAction = true
		}
	}

	if !foundFinance || !foundTasks || !foundAction {
		t.Errorf("missing one of the expected Gemini tool declarations")
	}
}

func TestGetAvailableTools(t *testing.T) {
	repo := newMockHubRepo()
	service := NewHubService(repo, &mockTransactor{})

	res, err := service.GetAvailableTools(context.Background(), "user-uuid-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if res.TotalTools < 5 {
		t.Errorf("expected at least 5 tools, got %d", res.TotalTools)
	}

	toolMap := make(map[string]bool)
	for _, tool := range res.Tools {
		toolMap[tool.Name] = true
	}

	requiredTools := []string{
		"create_task", "list_tasks", "update_task_status",
		"create_transaction", "list_transactions", "get_financial_summary",
	}

	for _, req := range requiredTools {
		if !toolMap[req] {
			t.Errorf("expected tool %s to be declared", req)
		}
	}
}

func TestExecuteTool_CreateTask(t *testing.T) {
	// Mock Task Management MCP Server
	mockTaskServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/mcp" {
			http.NotFound(w, r)
			return
		}

		var req MCPJSONRPCRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.Method == "tools/call" && req.Params["name"] == "create_task" {
			resp := MCPJSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: map[string]interface{}{
					"id":         101,
					"title":      "Selesaikan PR DeepLinkGuard",
					"status":     "in_progress",
					"is_focus":   true,
					"created_at": time.Now().Format(time.RFC3339),
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		http.Error(w, "unexpected tool call", http.StatusBadRequest)
	}))
	defer mockTaskServer.Close()

	repo := newMockHubRepo()
	_ = repo.UpsertLinkedApp(context.Background(), &LinkedApp{
		HubUserID:       "user-uuid-1",
		AppID:           "taskmanagement",
		AppName:         "Task Management",
		ConnectedUID:    "usr_task_123",
		ConnectedEmail:  "user@taskhub.app",
		APIBaseURL:      mockTaskServer.URL,
		ReadOnlyToken:   "mcp_pat_sample_token",
		CustomURLScheme: "taskhub://",
		IsActive:        true,
	})

	service := NewHubService(repo, &mockTransactor{})

	execRes, err := service.ExecuteTool(context.Background(), "user-uuid-1", &ExecuteToolRequest{
		ToolName: "create_task",
		Arguments: map[string]interface{}{
			"title":    "Selesaikan PR DeepLinkGuard",
			"priority": "high",
			"is_focus": true,
		},
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !execRes.Success {
		t.Fatalf("expected tool execution to succeed, got: %s", execRes.Message)
	}

	if execRes.Action == nil {
		t.Fatalf("expected suggested action link to be created")
	}

	if !strings.Contains(execRes.Action.DeepLink, "expected_uid=usr_task_123") {
		t.Errorf("expected deep link to contain expected_uid, got: %s", execRes.Action.DeepLink)
	}
}

func TestExecuteTool_CreateTransaction(t *testing.T) {
	// Mock F-Tracker MCP Server
	mockTrackerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/hub/mcp" {
			resp := MCPJSONRPCResponse{
				JSONRPC: "2.0",
				ID:      1,
				Result: map[string]interface{}{
					"message": "Transaction created successfully",
					"id":      "tx_abc_123",
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		http.NotFound(w, r)
	}))
	defer mockTrackerServer.Close()

	repo := newMockHubRepo()
	_ = repo.UpsertLinkedApp(context.Background(), &LinkedApp{
		HubUserID:       "user-uuid-1",
		AppID:           "ftracker",
		AppName:         "F-Tracker",
		ConnectedUID:    "usr_ftracker_456",
		ConnectedEmail:  "user@finance.app",
		APIBaseURL:      mockTrackerServer.URL,
		ReadOnlyToken:   "sample_ftracker_token",
		CustomURLScheme: "ftracker://",
		IsActive:        true,
	})

	service := NewHubService(repo, &mockTransactor{})

	execRes, err := service.ExecuteTool(context.Background(), "user-uuid-1", &ExecuteToolRequest{
		ToolName: "create_transaction",
		Arguments: map[string]interface{}{
			"amount":      50000,
			"type":        "expense",
			"description": "Makan siang padang",
			"category":    "Food",
		},
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !execRes.Success {
		t.Fatalf("expected tool execution to succeed, got: %s", execRes.Message)
	}

	if execRes.Action == nil {
		t.Fatalf("expected suggested action link")
	}

	if !strings.Contains(execRes.Action.DeepLink, "expected_uid=usr_ftracker_456") {
		t.Errorf("expected deep link to contain expected_uid, got: %s", execRes.Action.DeepLink)
	}
}

func TestProcessAIChat_PureAIFunctionCalling(t *testing.T) {
	repo := newMockHubRepo()
	service := NewHubService(repo, &mockTransactor{})

	// Case 1: When GEMINI_API_KEY is not set, must return error (no local heuristic fallback)
	_ = os.Unsetenv("GEMINI_API_KEY")
	_, err := service.ProcessAIChat(context.Background(), "user-uuid-1", "buat task review PR")
	if err == nil {
		t.Fatalf("expected error when GEMINI_API_KEY is not configured, got nil")
	}

	// Case 2: Multi-turn Gemini Function Calling loop
	// Mock Task Management MCP Server
	mockTaskServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := MCPJSONRPCResponse{
			JSONRPC: "2.0",
			ID:      1,
			Result: map[string]interface{}{
				"id":    101,
				"title": "Review PR Eka Hub",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockTaskServer.Close()

	_ = repo.UpsertLinkedApp(context.Background(), &LinkedApp{
		HubUserID:       "user-uuid-1",
		AppID:           "taskmanagement",
		AppName:         "Task Management",
		ConnectedUID:    "usr_task_123",
		ConnectedEmail:  "user@taskhub.app",
		APIBaseURL:      mockTaskServer.URL,
		ReadOnlyToken:   "token",
		CustomURLScheme: "taskhub://",
		IsActive:        true,
	})

	// Mock Google Gemini API Server
	turnCount := 0
	mockGeminiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		turnCount++
		w.Header().Set("Content-Type", "application/json")

		if turnCount == 1 {
			// Turn 1: Gemini decides to invoke functionCall "create_task"
			resp := map[string]interface{}{
				"candidates": []map[string]interface{}{
					{
						"content": map[string]interface{}{
							"role": "model",
							"parts": []map[string]interface{}{
								{
									"functionCall": map[string]interface{}{
										"name": "create_task",
										"args": map[string]interface{}{
											"title":    "Review PR Eka Hub",
											"priority": "high",
										},
									},
								},
							},
						},
					},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		// Turn 2: Gemini synthesizes response based on tool execution result
		resp := map[string]interface{}{
			"candidates": []map[string]interface{}{
				{
					"content": map[string]interface{}{
						"role": "model",
						"parts": []map[string]interface{}{
							{
								"text": "Tugas 'Review PR Eka Hub' berhasil dibuat di Task Management!",
							},
						},
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockGeminiServer.Close()

	t.Setenv("GEMINI_API_KEY", "test-key-123")
	t.Setenv("GEMINI_BASE_URL", mockGeminiServer.URL)

	chatResp, err := service.ProcessAIChat(context.Background(), "user-uuid-1", "Buatkan task review PR")
	if err != nil {
		t.Fatalf("expected no error from Gemini function calling, got %v", err)
	}

	if len(chatResp.ToolCalls) == 0 {
		t.Fatalf("expected tool call to be recorded")
	}

	if chatResp.ToolCalls[0].Name != "create_task" {
		t.Errorf("expected tool call create_task, got %s", chatResp.ToolCalls[0].Name)
	}

	if !strings.Contains(chatResp.Reply, "Review PR Eka Hub") {
		t.Errorf("expected synthesized reply from Gemini, got %s", chatResp.Reply)
	}

	if len(chatResp.ActionLinks) == 0 {
		t.Errorf("expected action deep link to be generated")
	}
}

func TestProcessAIChat_MultiToolOrchestration(t *testing.T) {
	repo := newMockHubRepo()
	service := NewHubService(repo, &mockTransactor{})

	// Setup linked F-Tracker & Task Management apps
	_ = repo.UpsertLinkedApp(context.Background(), &LinkedApp{
		AppID:           "ftracker",
		AppName:         "F-Tracker",
		HubUserID:       "user-uuid-1",
		ConnectedUID:    "usr_ftracker_123",
		CustomURLScheme: "ftracker://",
		HighlightMetric: "Rp 3.500.000",
		HighlightLabel:  "Pengeluaran Bulan Ini",
		IsActive:        true,
	})
	_ = repo.UpsertLinkedApp(context.Background(), &LinkedApp{
		AppID:           "taskmanagement",
		AppName:         "Task Management",
		HubUserID:       "user-uuid-1",
		ConnectedUID:    "usr_task_456",
		CustomURLScheme: "taskhub://",
		HighlightMetric: "3 Tugas",
		HighlightLabel:  "Prioritas Hari Ini",
		IsActive:        true,
	})

	turnCounter := 0
	mockGeminiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		turnCounter++
		w.Header().Set("Content-Type", "application/json")

		if turnCounter == 1 {
			// Turn 1: Return PARALLEL function calls for both F-Tracker and Task Management
			resp := map[string]interface{}{
				"candidates": []map[string]interface{}{
					{
						"content": map[string]interface{}{
							"role": "model",
							"parts": []map[string]interface{}{
								{
									"functionCall": map[string]interface{}{
										"name": "get_financial_summary",
										"args": map[string]interface{}{},
									},
								},
								{
									"functionCall": map[string]interface{}{
										"name": "get_tasks_summary",
										"args": map[string]interface{}{"filter": "today"},
									},
								},
							},
						},
					},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		// Turn 2: Synthesize combined reply
		resp := map[string]interface{}{
			"candidates": []map[string]interface{}{
				{
					"content": map[string]interface{}{
						"role": "model",
						"parts": []map[string]interface{}{
							{
								"text": "Ringkasan Finansial & Tugas Hari Ini: Pengeluaran tercatat Rp 3.500.000 dan Anda memiliki 3 tugas prioritas hari ini.",
							},
						},
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockGeminiServer.Close()

	t.Setenv("GEMINI_API_KEY", "test-key-123")
	t.Setenv("GEMINI_BASE_URL", mockGeminiServer.URL)

	chatResp, err := service.ProcessAIChat(context.Background(), "user-uuid-1", "Berikan ringkasan keuangan dan tugas hari ini")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(chatResp.ToolCalls) != 2 {
		t.Fatalf("expected 2 tool calls for multi-tool orchestration, got %d", len(chatResp.ToolCalls))
	}

	names := []string{chatResp.ToolCalls[0].Name, chatResp.ToolCalls[1].Name}
	if !((names[0] == "get_financial_summary" && names[1] == "get_tasks_summary") ||
		(names[0] == "get_tasks_summary" && names[1] == "get_financial_summary")) {
		t.Errorf("unexpected tool call names: %v", names)
	}

	if !strings.Contains(chatResp.Reply, "Pengeluaran") || !strings.Contains(chatResp.Reply, "tugas") {
		t.Errorf("expected combined synthesized reply, got: %s", chatResp.Reply)
	}
}

func TestBuildGeminiConversationContents(t *testing.T) {
	history := []StoredChatMessage{
		{Role: "user", Content: "Halo, nama saya Budi"},
		{Role: "assistant", Content: "Halo Budi! Ada yang bisa saya bantu?"},
	}
	currentPrompt := "Berapa pengeluaran saya hari ini?"

	contents := buildGeminiConversationContents(history, currentPrompt)
	if len(contents) != 3 {
		t.Fatalf("expected 3 conversation turns, got %d", len(contents))
	}

	if contents[0]["role"] != "user" {
		t.Errorf("expected turn 0 to be user, got %v", contents[0]["role"])
	}
	if contents[1]["role"] != "model" {
		t.Errorf("expected turn 1 to be model, got %v", contents[1]["role"])
	}
	if contents[2]["role"] != "user" {
		t.Errorf("expected turn 2 to be user, got %v", contents[2]["role"])
	}

	parts, ok := contents[2]["parts"].([]map[string]interface{})
	if !ok || parts[0]["text"] != currentPrompt {
		t.Errorf("expected turn 2 to contain currentPrompt, got %v", parts)
	}
}
