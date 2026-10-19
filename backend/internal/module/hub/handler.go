package hub

import (
	"golang/pkg/constants"
	authMiddleware "golang/pkg/middleware/auth"
	"golang/pkg/response"
	"golang/pkg/validation"

	"github.com/gofiber/fiber/v2"
)

type HubHandler struct {
	service HubService
}

func NewHubHandler(service HubService) *HubHandler {
	return &HubHandler{service: service}
}

// InitiateLink handles starting the OAuth-like deep link flow.
func (h *HubHandler) InitiateLink() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := authMiddleware.GetUserID(c)
		if err != nil {
			return response.ErrorResponse(c, fiber.StatusUnauthorized, constants.MsgUnauthorized, constants.CodeUnauthorized)
		}

		var req InitiateLinkRequest
		if err := validation.Default.BindAndValidate(c, &req); err != nil {
			return response.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request payload: "+err.Error(), constants.CodeBadRequest)
		}

		res, err := h.service.InitiateLink(c.UserContext(), userID, req.AppID)
		if err != nil {
			return response.ErrorResponse(c, fiber.StatusInternalServerError, err.Error(), constants.CodeInternalError)
		}

		return response.SuccessResponse(c, res, "Link initiated successfully")
	}
}

// ExchangeLink completes the handshake by receiving credentials from the child app.
func (h *HubHandler) ExchangeLink() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req ExchangeLinkRequest
		if err := validation.Default.BindAndValidate(c, &req); err != nil {
			return response.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request payload: "+err.Error(), constants.CodeBadRequest)
		}

		if err := h.service.ExchangeLink(c.UserContext(), &req); err != nil {
			return response.ErrorResponse(c, fiber.StatusBadRequest, err.Error(), constants.CodeBadRequest)
		}

		return response.SuccessMessageResponse(c, "App linked successfully")
	}
}

// UpdateAccount handles "Perbarui Koneksi Hub ke Akun Saat Ini" (Account Mismatch option 2).
func (h *HubHandler) UpdateAccount() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := authMiddleware.GetUserID(c)
		if err != nil {
			return response.ErrorResponse(c, fiber.StatusUnauthorized, constants.MsgUnauthorized, constants.CodeUnauthorized)
		}

		var req UpdateConnectionRequest
		if err := validation.Default.BindAndValidate(c, &req); err != nil {
			return response.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request payload: "+err.Error(), constants.CodeBadRequest)
		}

		if err := h.service.UpdateLinkedAccount(c.UserContext(), userID, &req); err != nil {
			return response.ErrorResponse(c, fiber.StatusInternalServerError, err.Error(), constants.CodeInternalError)
		}

		return response.SuccessMessageResponse(c, "Connection updated to current account")
	}
}

// GetLinkedApps returns all active linked apps for the authenticated user.
func (h *HubHandler) GetLinkedApps() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := authMiddleware.GetUserID(c)
		if err != nil {
			return response.ErrorResponse(c, fiber.StatusUnauthorized, constants.MsgUnauthorized, constants.CodeUnauthorized)
		}

		apps, err := h.service.GetLinkedApps(c.UserContext(), userID)
		if err != nil {
			return response.ErrorResponse(c, fiber.StatusInternalServerError, err.Error(), constants.CodeInternalError)
		}

		return response.SuccessResponse(c, apps, "Linked apps retrieved successfully")
	}
}

// UnlinkApp removes a connection.
func (h *HubHandler) UnlinkApp() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := authMiddleware.GetUserID(c)
		if err != nil {
			return response.ErrorResponse(c, fiber.StatusUnauthorized, constants.MsgUnauthorized, constants.CodeUnauthorized)
		}

		appID := c.Params("app_id")
		if err := h.service.UnlinkApp(c.UserContext(), userID, appID); err != nil {
			return response.ErrorResponse(c, fiber.StatusInternalServerError, err.Error(), constants.CodeInternalError)
		}

		return response.SuccessMessageResponse(c, "App unlinked successfully")
	}
}

// GetSummary returns Bento Grid mini summary cards for all linked apps.
func (h *HubHandler) GetSummary() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := authMiddleware.GetUserID(c)
		if err != nil {
			return response.ErrorResponse(c, fiber.StatusUnauthorized, constants.MsgUnauthorized, constants.CodeUnauthorized)
		}

		summaries, err := h.service.SyncAllSummaries(c.UserContext(), userID)
		if err != nil {
			return response.ErrorResponse(c, fiber.StatusInternalServerError, err.Error(), constants.CodeInternalError)
		}

		return response.SuccessResponse(c, summaries, "App summaries retrieved successfully")
	}
}

// GetDailyBriefing returns the Daily AI Briefing synthesized from all connected apps.
func (h *HubHandler) GetDailyBriefing() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := authMiddleware.GetUserID(c)
		if err != nil {
			return response.ErrorResponse(c, fiber.StatusUnauthorized, constants.MsgUnauthorized, constants.CodeUnauthorized)
		}

		briefing, err := h.service.GetDailyBriefing(c.UserContext(), userID)
		if err != nil {
			return response.ErrorResponse(c, fiber.StatusInternalServerError, err.Error(), constants.CodeInternalError)
		}

		return response.SuccessResponse(c, briefing, "Daily briefing generated successfully")
	}
}

// AIChat handles Cross-App AI Chat prompts with Function Calling / Tool Use.
func (h *HubHandler) AIChat() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := authMiddleware.GetUserID(c)
		if err != nil {
			return response.ErrorResponse(c, fiber.StatusUnauthorized, constants.MsgUnauthorized, constants.CodeUnauthorized)
		}

		var req AIChatRequest
		if err := validation.Default.BindAndValidate(c, &req); err != nil {
			return response.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request: "+err.Error(), constants.CodeBadRequest)
		}

		resp, err := h.service.ProcessAIChat(c.UserContext(), userID, req.Message)
		if err != nil {
			return response.ErrorResponse(c, fiber.StatusInternalServerError, err.Error(), constants.CodeInternalError)
		}

		return response.SuccessResponse(c, resp, "Chat response processed successfully")
	}
}

// ExecuteTool directly calls an integrated satellite tool.
func (h *HubHandler) ExecuteTool() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := authMiddleware.GetUserID(c)
		if err != nil {
			return response.ErrorResponse(c, fiber.StatusUnauthorized, constants.MsgUnauthorized, constants.CodeUnauthorized)
		}

		var req ExecuteToolRequest
		if err := validation.Default.BindAndValidate(c, &req); err != nil {
			return response.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request payload: "+err.Error(), constants.CodeBadRequest)
		}

		res, err := h.service.ExecuteTool(c.UserContext(), userID, &req)
		if err != nil {
			return response.ErrorResponse(c, fiber.StatusInternalServerError, err.Error(), constants.CodeInternalError)
		}

		return response.SuccessResponse(c, res, "Tool executed successfully")
	}
}

// GetAvailableTools returns all tools available in Eka-Hub for AI / MCP orchestration.
func (h *HubHandler) GetAvailableTools() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, err := authMiddleware.GetUserID(c)
		if err != nil {
			return response.ErrorResponse(c, fiber.StatusUnauthorized, constants.MsgUnauthorized, constants.CodeUnauthorized)
		}

		res, err := h.service.GetAvailableTools(c.UserContext(), userID)
		if err != nil {
			return response.ErrorResponse(c, fiber.StatusInternalServerError, err.Error(), constants.CodeInternalError)
		}

		return response.SuccessResponse(c, res, "Available tools retrieved successfully")
	}
}

// MockFTrackerSummary provides a reference implementation for F-Tracker GET /api/v1/hub/summary.
func MockFTrackerSummary(c *fiber.Ctx) error {
	userUID := c.Get("X-Hub-User-UID", "usr_ftracker_sample_123")
	data := AppSummaryResponse{
		AppID:           "ftracker",
		AppName:         "F-Tracker",
		ConnectedUID:    userUID,
		ConnectedEmail:  "user@finance.app",
		HighlightMetric: "Rp 4.250.000",
		HighlightLabel:  "Pengeluaran Bulan Ini",
		AIContextData: map[string]interface{}{
			"monthly_spent":    4250000,
			"budget_limit":     10000000,
			"remaining_budget": 5750000,
			"top_category":     "Food & Beverage",
		},
		DeepLink: "ftracker://expenses?category=food&expected_uid=" + userUID,
	}
	return response.SuccessResponse(c, data, "F-Tracker summary retrieved successfully")
}

// MockTaskSummary provides a reference implementation for Task Management GET /api/v1/hub/summary.
func MockTaskSummary(c *fiber.Ctx) error {
	userUID := c.Get("X-Hub-User-UID", "usr_task_sample_456")
	data := AppSummaryResponse{
		AppID:           "taskmanagement",
		AppName:         "Task Management",
		ConnectedUID:    userUID,
		ConnectedEmail:  "user@taskhub.app",
		HighlightMetric: "5 Tugas Menunggu",
		HighlightLabel:  "Deadline Hari Ini",
		AIContextData: map[string]interface{}{
			"pending_tasks":      5,
			"due_today":          2,
			"overdue":            1,
			"high_priority_task": "Deploy Eka-Hub to Staging",
		},
		DeepLink: "taskhub://tasks?filter=today&expected_uid=" + userUID,
	}
	return response.SuccessResponse(c, data, "Task Management summary retrieved successfully")
}
