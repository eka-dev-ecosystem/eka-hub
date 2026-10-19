package examples

import (
	"log"

	"golang/database"
	usermodule "golang/internal/module/user"

	"github.com/gofiber/fiber/v2"
)

// SetupUserRoutesWithRepositoryPattern demonstrates how to setup user routes using repository pattern
// This is an example of how to wire up the layers: Repository -> Service -> Handler -> Router
func SetupUserRoutesWithRepositoryPattern(app *fiber.App, db *database.PostgresDB) {
	log.Println("=== Setting up User Routes with Repository Pattern ===")

	// Layer 1: Repository Layer (Data Access)
	// Responsible for direct database operations
	userRepo := usermodule.NewUserRepository(db.GetDB())

	// Layer 2: Service Layer (Business Logic)
	// Responsible for business rules, validation, and orchestration
	userService := usermodule.NewUserService(userRepo, db)

	// Layer 3: Handler Layer (HTTP/API Layer)
	// Responsible for HTTP request/response handling
	userHandler := usermodule.NewUserHandler(userService)

	// Layer 4: Router Layer (Routes Configuration)
	api := app.Group("/api/v1")
	users := api.Group("/users")
	{
		// Public routes
		users.Get("/", userHandler.GetUsers())         // GET /api/v1/users
		users.Get("/:id", userHandler.GetUser())       // GET /api/v1/users/:id
		users.Post("/", userHandler.CreateUser())      // POST /api/v1/users
		users.Put("/:id", userHandler.UpdateUser())    // PUT /api/v1/users/:id
		users.Delete("/:id", userHandler.DeleteUser()) // DELETE /api/v1/users/:id (soft delete)

		// Admin routes (example - you can add auth middleware here)
		admin := users.Group("/admin")
		{
			admin.Delete("/:id", userHandler.HardDeleteUser()) // DELETE /api/v1/users/admin/:id (permanent delete)
		}
	}

	log.Println("✓ User routes configured successfully")
}

// Example: How the layers work together
//
// 1. Client sends HTTP Request
//    POST /api/v1/users
//    Body: {"name": "John Doe", "email": "john@example.com", "phone": "1234567890"}
//
// 2. Handler Layer (handlers/user_handler.go)
//    - Parses request body
//    - Validates input using go-playground/validator
//    - Calls Service Layer
//
// 3. Service Layer (services/user_service.go)
//    - Implements business logic (check if email exists, generate UUID, etc.)
//    - Calls Repository Layer
//
// 4. Repository Layer (repository/user_repository.go)
//    - Executes SQL queries
//    - Returns data to Service Layer
//
// 5. Service Layer
//    - Processes repository response
//    - Returns formatted data to Handler
//
// 6. Handler Layer
//    - Formats HTTP response
//    - Returns JSON response to client
//
// Benefits of Repository Pattern:
// - Separation of Concerns: Each layer has a specific responsibility
// - Testability: Easy to mock repositories for unit testing
// - Maintainability: Changes in database don't affect business logic
// - Reusability: Services can be reused across different handlers
// - Flexibility: Easy to swap database implementations (e.g., PostgreSQL to MongoDB)
