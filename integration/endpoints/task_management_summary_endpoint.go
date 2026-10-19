package endpoints

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// Example implementation for Task Management Backend
// Mount this at: GET /api/v1/hub/summary
// Protected by: Authorization: Bearer <read_only_token>
func RegisterTaskManagementHubEndpoint(app *fiber.App) {
	app.Get("/api/v1/hub/summary", func(c *fiber.Ctx) error {
		// 1. Verify read_only_token
		token := c.Get("Authorization")
		if token == "" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
		}

		// 2. Extract user UID
		userUID := c.Get("X-Hub-User-UID", "usr_task_456")

		// 3. Query Task Management database
		// pendingCount := repo.CountPendingTasks(userUID)
		// dueTodayCount := repo.CountDueToday(userUID)

		// 4. Return standard modular contract
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"app_id":           "taskmanagement",
			"app_name":         "Task Management",
			"connected_uid":    userUID,
			"connected_email":  "eka@taskhub.app",
			"highlight_metric": "5 Tugas Pending",
			"highlight_label":  "Deadline Hari Ini",
			"ai_context_data": fiber.Map{
				"pending_tasks":      5,
				"due_today":          2,
				"overdue":            1,
				"high_priority_task": "Deploy Eka-Hub to Staging",
			},
			"deep_link": "taskhub://tasks?filter=today&expected_uid=" + userUID,
		})
	})
}
