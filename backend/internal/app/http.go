package app

import (
	"context"
	"log/slog"
	"os"
	"time"

	"golang/internal/module"
	"golang/pkg/constants"
	"golang/pkg/logging"
	"golang/pkg/middleware"
	"golang/pkg/response"

	"github.com/gofiber/contrib/otelfiber/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// Health probe status values (kept as constants to stay consistent across the
// /health payload and to satisfy goconst).
const (
	statusConnected = "connected"
	statusError     = "error"
	statusDegraded  = "degraded"
	statusDisabled  = "disabled"
)

// NewHTTP builds the Fiber application and registers all routes via feature
// modules.
func NewHTTP(c *Container) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler:          middleware.ErrorHandler(),
		AppName:               c.Config.App.Name,
		DisableStartupMessage: true, // keep the output stream pure JSON
	})

	if os.Getenv("OTEL_ENABLED") == "true" {
		app.Use(otelfiber.Middleware())
	}

	// Global middleware: correlation ID, structured access log, CORS, panic
	// recovery. Order matters: request ID first so everything downstream can
	// correlate logs to a single request.
	app.Use(logging.RequestIDMiddleware())
	app.Use(logging.AccessLogMiddleware())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     c.Config.CORSOrigins(),
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Request-ID",
		AllowMethods:     "GET, POST, PUT, DELETE, PATCH, OPTIONS",
		AllowCredentials: true,
	}))
	app.Use(recover.New(recover.Config{
		EnableStackTrace: !c.Config.IsProduction(),
	}))

	c.registerHealthRoutes(app)

	// Feature modules own their routes.
	api := app.Group("/api/v1")
	r := module.NewRouter(c.DB, c.JWTService, c.Hub, c.Config)
	r.Register(api)

	// 404 handler.
	app.Use(func(ct *fiber.Ctx) error {
		return response.ErrorResponse(ct, fiber.StatusNotFound, "Route not found", constants.CodeNotFound)
	})

	return app
}

func (c *Container) registerHealthRoutes(app *fiber.App) {
	app.Get("/", func(ct *fiber.Ctx) error {
		return ct.JSON(map[string]interface{}{
			"success": true,
			"message": "Welcome to Golang API Template",
			"data": map[string]interface{}{
				"status": "running",
				"time":   time.Now().Format(time.RFC3339),
			},
			"code": "SUCCESS",
		})
	})

	app.Get("/health", func(ct *fiber.Ctx) error {
		status := map[string]interface{}{
			"status": "healthy",
			"app":    c.Config.App.Name,
			"features": map[string]interface{}{
				"redis":           c.Config.Feature.Redis,
				"rabbitmq":        c.Config.Feature.RabbitMQ,
				"realtime":        c.Config.Feature.Realtime,
				"realtime_sse":    c.Config.Feature.RealtimeSSE,
				"realtime_ws":     c.Config.Feature.RealtimeWS,
				"realtime_bridge": c.Config.Feature.RealtimeBridge,
			},
		}

		dbStatus := statusConnected
		if err := c.DB.HealthCheck(); err != nil {
			dbStatus = statusError
			status["status"] = statusDegraded
		}
		status["database"] = dbStatus

		status["redis"] = statusDisabled
		if c.Cache != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if err := c.Cache.HealthCheck(ctx); err != nil {
				status["redis"] = statusError
				status["status"] = statusDegraded
			} else {
				status["redis"] = statusConnected
			}
		}

		status["rabbitmq"] = statusDisabled
		if c.Broker != nil {
			if err := c.Broker.HealthCheck(); err != nil {
				status["rabbitmq"] = statusError
				status["status"] = statusDegraded
			} else {
				status["rabbitmq"] = statusConnected
			}
		}

		status["realtime"] = statusDisabled
		if c.Hub != nil {
			status["realtime"] = statusConnected
		}

		return ct.JSON(status)
	})
}

// Serve starts the HTTP server on the configured host:port and blocks until
// the listener returns.
func (c *Container) Serve(app *fiber.App) error {
	addr := c.Config.Server.Host + ":" + c.Config.Server.Port
	c.Logger.Info("server listening",
		slog.String("addr", addr),
		slog.String("env", c.Config.App.Environment),
		slog.String("log_level", c.Config.Log.Level),
		slog.String("log_file", c.Config.Log.File),
	)
	return app.Listen(addr)
}
