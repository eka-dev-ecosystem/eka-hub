// Package module contains the feature modules of the application. Every
// module is self-contained: it owns its DTOs, repository, service and
// handler, and registers its own routes.
package module

import (
	"golang/config"
	"golang/database"
	authmodule "golang/internal/module/auth"
	"golang/internal/module/hub"
	usermodule "golang/internal/module/user"
	"golang/pkg/auth"
	authMiddleware "golang/pkg/middleware/auth"
	"golang/pkg/realtime"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// Router bundles module dependencies and registers routes under a base group.
type Router struct {
	DB         *database.PostgresDB
	JWTService *auth.JWTService
	Hub        *realtime.Hub // nil when the realtime feature is disabled
	Config     *config.Config
}

// NewRouter builds the module router.
func NewRouter(db *database.PostgresDB, jwtService *auth.JWTService, hub *realtime.Hub, cfg *config.Config) *Router {
	return &Router{DB: db, JWTService: jwtService, Hub: hub, Config: cfg}
}

// Register mounts every enabled module on the API router.
func (r *Router) Register(api fiber.Router) {
	r.registerAuth(api)
	r.registerUsers(api)
	r.registerHub(api)
	r.registerProtected(api)
	if r.Hub != nil {
		r.registerRealtime(api)
	}
}

func (r *Router) registerAuth(api fiber.Router) {
	handler := authmodule.NewAuthHandler(authmodule.NewAuthService(r.DB, r.JWTService, r.Config))

	auth := api.Group("/auth")
	auth.Post("/register", handler.Register())
	auth.Post("/login", handler.Login())
	auth.Post("/oauth", handler.OAuthLogin())
	auth.Post("/refresh", handler.RefreshToken())

	auth.Use(authMiddleware.AuthMiddleware(r.JWTService))
	auth.Get("/profile", handler.GetProfile())
	auth.Put("/profile", handler.UpdateProfile())
	auth.Post("/change-password", handler.ChangePassword())
	auth.Post("/logout", handler.Logout())
	auth.Post("/logout-all", handler.LogoutAll())
}

func (r *Router) registerUsers(api fiber.Router) {
	repo := usermodule.NewUserRepository(r.DB.GetDB())
	service := usermodule.NewUserService(repo, r.DB)
	handler := usermodule.NewUserHandler(service)

	users := api.Group("/users")
	users.Get("/", handler.GetUsers())
	users.Get("/:id", handler.GetUser())
	users.Post("/", handler.CreateUser())
	users.Put("/:id", handler.UpdateUser())
	users.Delete("/:id", handler.DeleteUser())

	admin := users.Group("/admin")
	admin.Delete("/:id", handler.HardDeleteUser())
}

func (r *Router) registerProtected(api fiber.Router) {
	protected := api.Group("/protected")
	protected.Use(authMiddleware.AuthMiddleware(r.JWTService))
	protected.Get("/data", func(c *fiber.Ctx) error {
		userID, _ := authMiddleware.GetUserID(c)
		email, _ := authMiddleware.GetEmail(c)
		return c.JSON(fiber.Map{
			"message": "This is protected data",
			"user_id": userID,
			"email":   email,
		})
	})

	api.Get("/public-or-private", authMiddleware.OptionalAuthMiddleware(r.JWTService), func(c *fiber.Ctx) error {
		if userID, err := authMiddleware.GetUserID(c); err == nil {
			return c.JSON(fiber.Map{"message": "Private data", "authenticated": true, "user_id": userID})
		}
		return c.JSON(fiber.Map{"message": "Public data", "authenticated": false})
	})
}

func (r *Router) registerRealtime(api fiber.Router) {
	rt := api.Group("/realtime")

	if r.Config.Feature.RealtimeWS {
		r.registerWebSocketRoutes(rt)
	}
	if r.Config.Feature.RealtimeSSE {
		r.registerSSERoutes(rt)
	}
}

func (r *Router) registerWebSocketRoutes(rt fiber.Router) {
	wsHandler := NewWSHandler(r.Hub)
	ws := rt.Group("/ws")
	ws.Get("/connect", upgradeWS(), websocket.New(wsHandler.Handle))
	ws.Get("/stats", authMiddleware.AuthMiddleware(r.JWTService), wsHandler.GetStats)
}

func (r *Router) registerSSERoutes(rt fiber.Router) {
	sseHandler := NewSSEHandler(r.Hub)
	sse := rt.Group("/sse")
	sse.Get("/events", sseHandler.Events)
	sse.Get("/subscribe", sseHandler.Subscribe)
	sse.Post("/broadcast", authMiddleware.AuthMiddleware(r.JWTService), sseHandler.BroadcastToChannel)
	sse.Post("/send", authMiddleware.AuthMiddleware(r.JWTService), sseHandler.SendToUser)
	sse.Get("/stats", authMiddleware.AuthMiddleware(r.JWTService), sseHandler.GetStats)
}

func (r *Router) registerHub(api fiber.Router) {
	repo := hub.NewHubRepository(r.DB.GetDB())
	service := hub.NewHubService(repo, r.DB)
	handler := hub.NewHubHandler(service)

	hubGroup := api.Group("/hub")

	// Handshake exchange from satellite apps
	hubGroup.Post("/link/exchange", handler.ExchangeLink())

	// Reference mock endpoints for F-Tracker & Task Management summary
	hubGroup.Get("/mock/ftracker/summary", hub.MockFTrackerSummary)
	hubGroup.Get("/mock/taskmanagement/summary", hub.MockTaskSummary)

	// Protected Hub APIs (requires user auth)
	protected := hubGroup.Group("/")
	protected.Use(authMiddleware.AuthMiddleware(r.JWTService))

	protected.Post("/link/initiate", handler.InitiateLink())
	protected.Put("/link/update", handler.UpdateAccount())
	protected.Get("/apps", handler.GetLinkedApps())
	protected.Delete("/apps/:app_id", handler.UnlinkApp())

	protected.Get("/summary", handler.GetSummary())
	protected.Get("/briefing", handler.GetDailyBriefing())
	protected.Post("/chat", handler.AIChat())
	protected.Get("/tools", handler.GetAvailableTools())
	protected.Post("/tools/call", handler.ExecuteTool())
}
