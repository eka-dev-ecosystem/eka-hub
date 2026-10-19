package endpoints

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// Example implementation for F-Tracker Backend
// Mount this at: GET /api/v1/hub/summary
// Protected by: Authorization: Bearer <read_only_token>
func RegisterFTrackerHubEndpoint(app *fiber.App) {
	app.Get("/api/v1/hub/summary", func(c *fiber.Ctx) error {
		// 1. Verify read_only_token from Hub
		token := c.Get("Authorization")
		if token == "" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
		}

		// 2. Extract user UID from token or header
		userUID := c.Get("X-Hub-User-UID", "usr_ftracker_123")

		// 3. Query F-Tracker database for monthly spent, budget limit, top category
		// Example query results:
		// monthlySpent := repo.GetMonthlyExpense(userUID)
		// budget := repo.GetUserBudget(userUID)

		// 4. Return standard modular contract
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"app_id":           "ftracker",
			"app_name":         "F-Tracker",
			"connected_uid":    userUID,
			"connected_email":  "eka@finance.app",
			"highlight_metric": "Rp 4.250.000",
			"highlight_label":  "Pengeluaran Bulan Ini",
			"ai_context_data": fiber.Map{
				"monthly_spent":      4250000,
				"budget_limit":       10000000,
				"remaining_budget":   5750000,
				"top_category":       "Food & Beverage",
				"overbudget_warning": false,
			},
			"deep_link": "ftracker://expenses?category=food&expected_uid=" + userUID,
		})
	})
}
