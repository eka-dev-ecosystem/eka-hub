package hub

// GeminiToolDeclaration defines the official Google Gemini API function calling schema.
type GeminiFunctionDeclaration struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type GeminiTool struct {
	FunctionDeclarations []GeminiFunctionDeclaration `json:"function_declarations"`
}

// GetHubGeminiTools returns the function definitions provided to the Gemini model
// for Cross-App orchestration (F-Tracker & Task Management).
func GetHubGeminiTools() GeminiTool {
	return GeminiTool{
		FunctionDeclarations: []GeminiFunctionDeclaration{
			{
				Name:        "create_task",
				Description: "Create a new actionable task directly in Task Management app pipeline.",
				Parameters: map[string]interface{}{
					"type": "OBJECT",
					"properties": map[string]interface{}{
						"title": map[string]interface{}{
							"type":        "STRING",
							"description": "Task title or description of what needs to be done.",
						},
						"description": map[string]interface{}{
							"type":        "STRING",
							"description": "Optional markdown details or notes about the task.",
						},
						"priority": map[string]interface{}{
							"type":        "STRING",
							"description": "Task priority level: 'low', 'medium', 'high', 'urgent'. Default is 'medium'.",
						},
						"due_date": map[string]interface{}{
							"type":        "STRING",
							"description": "Optional deadline in ISO-8601 or YYYY-MM-DD format (e.g. '2026-10-20').",
						},
						"is_focus": map[string]interface{}{
							"type":        "BOOLEAN",
							"description": "Whether to pin this task to Today's Focus.",
						},
						"tags": map[string]interface{}{
							"type": "ARRAY",
							"items": map[string]interface{}{
								"type": "STRING",
							},
							"description": "Optional categorization tags, e.g. ['work', 'urgent'].",
						},
					},
					"required": []string{"title"},
				},
			},
			{
				Name:        "list_tasks",
				Description: "Retrieve tasks filtered by status, priority, or focus from Task Management.",
				Parameters: map[string]interface{}{
					"type": "OBJECT",
					"properties": map[string]interface{}{
						"status": map[string]interface{}{
							"type":        "STRING",
							"description": "Filter by status: 'backlog', 'todo', 'in_progress', 'review', 'done'.",
						},
						"priority": map[string]interface{}{
							"type":        "STRING",
							"description": "Filter by priority: 'low', 'medium', 'high', 'urgent'.",
						},
						"is_focus": map[string]interface{}{
							"type":        "BOOLEAN",
							"description": "Filter by pinned Today's Focus tasks (true for today's priority tasks).",
						},
						"search": map[string]interface{}{
							"type":        "STRING",
							"description": "Keyword search across task titles and descriptions.",
						},
						"limit": map[string]interface{}{
							"type":        "INTEGER",
							"description": "Maximum number of tasks to return (default: 10).",
						},
					},
				},
			},
			{
				Name:        "update_task_status",
				Description: "Update the status of an existing task in Task Management.",
				Parameters: map[string]interface{}{
					"type": "OBJECT",
					"properties": map[string]interface{}{
						"task_id": map[string]interface{}{
							"type":        "INTEGER",
							"description": "The ID of the task to update.",
						},
						"status": map[string]interface{}{
							"type":        "STRING",
							"description": "New status: 'backlog', 'todo', 'in_progress', 'review', 'done'.",
						},
					},
					"required": []string{"task_id", "status"},
				},
			},
			{
				Name:        "create_transaction",
				Description: "Record a new expense or income transaction in F-Tracker finance app whenever user mentions buying, spending, receiving money, or recording a transaction.",
				Parameters: map[string]interface{}{
					"type": "OBJECT",
					"properties": map[string]interface{}{
						"amount": map[string]interface{}{
							"type":        "NUMBER",
							"description": "Transaction amount in numerical value (e.g. 50000, 35000, 100000).",
						},
						"type": map[string]interface{}{
							"type":        "STRING",
							"description": "Transaction type: 'expense' (pengeluaran) or 'income' (pemasukan).",
						},
						"description": map[string]interface{}{
							"type":        "STRING",
							"description": "Description of the expense or income (e.g., 'Kopi susu', 'Makan siang padang', 'Beli bensin').",
						},
						"category": map[string]interface{}{
							"type":        "STRING",
							"description": "Category name (e.g. 'Food', 'Transportation', 'Salary', 'Shopping', 'General').",
						},
						"date": map[string]interface{}{
							"type":        "STRING",
							"description": "Optional transaction date (e.g. 'YYYY-MM-DD', 'today', 'yesterday'). Defaults to today.",
						},
					},
					"required": []string{"amount", "type", "description"},
				},
			},
			{
				Name:        "list_transactions",
				Description: "Retrieve recent financial transactions from F-Tracker.",
				Parameters: map[string]interface{}{
					"type": "OBJECT",
					"properties": map[string]interface{}{
						"type": map[string]interface{}{
							"type":        "STRING",
							"description": "Filter by type: 'income', 'expense', or 'all'.",
						},
						"category": map[string]interface{}{
							"type":        "STRING",
							"description": "Filter by category.",
						},
						"date": map[string]interface{}{
							"type":        "STRING",
							"description": "Filter by date: 'today', 'yesterday', or specific date 'YYYY-MM-DD'.",
						},
						"limit": map[string]interface{}{
							"type":        "INTEGER",
							"description": "Max number of transactions to return (default: 10).",
						},
					},
				},
			},
			{
				Name:        "get_financial_summary",
				Description: "Retrieve current financial metrics from F-Tracker including total expenses, remaining monthly budget, and top spending categories.",
				Parameters: map[string]interface{}{
					"type": "OBJECT",
					"properties": map[string]interface{}{
						"month": map[string]interface{}{
							"type":        "STRING",
							"description": "Optional month in YYYY-MM format. Defaults to current month.",
						},
						"category": map[string]interface{}{
							"type":        "STRING",
							"description": "Optional category filter like Food, Transport, Utilities.",
						},
					},
				},
			},
			{
				Name:        "get_tasks_summary",
				Description: "Retrieve daily task overview from Task Management app including overdue tasks, tasks due today, and high-priority items.",
				Parameters: map[string]interface{}{
					"type": "OBJECT",
					"properties": map[string]interface{}{
						"filter": map[string]interface{}{
							"type":        "STRING",
							"description": "Task filter: 'all', 'today', 'overdue', 'priority'.",
						},
						"limit": map[string]interface{}{
							"type":        "INTEGER",
							"description": "Maximum number of tasks to return (default: 5).",
						},
					},
				},
			},
			{
				Name:        "generate_cross_app_action",
				Description: "Generate an authenticated, safe deep link with expected_uid to navigate the user directly into a child app to view or modify details.",
				Parameters: map[string]interface{}{
					"type": "OBJECT",
					"properties": map[string]interface{}{
						"app_id": map[string]interface{}{
							"type":        "STRING",
							"description": "Target app identifier: 'ftracker' or 'taskmanagement'.",
						},
						"target_route": map[string]interface{}{
							"type":        "STRING",
							"description": "Target screen route inside the app, e.g., 'expenses', 'add-expense', 'tasks', 'task-detail'.",
						},
						"action_title": map[string]interface{}{
							"type":        "STRING",
							"description": "Short user-facing action title, e.g., 'Buka Laporan Pengeluaran', 'Lihat Tugas Tertunda'.",
						},
						"query_params": map[string]interface{}{
							"type":        "STRING",
							"description": "Optional key=value query string like 'category=food' or 'task_id=99'.",
						},
					},
					"required": []string{"app_id", "target_route", "action_title"},
				},
			},
		},
	}
}
