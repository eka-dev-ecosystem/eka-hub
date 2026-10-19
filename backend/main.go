// Command main is a small bootstrap demo that initializes every configured
// integration (PostgreSQL, Redis, RabbitMQ) and prints their status.
//
// Optional integrations are controlled through feature flags. Example:
//
//	RABBITMQ_ENABLED=false go run main.go
//
// The real HTTP API lives in cmd/server.
package main

import (
	"context"
	"log/slog"
	"time"

	"golang/cache"
	"golang/config"
	"golang/database"
	"golang/pkg/logging"
	queuepkg "golang/pkg/queue"
)

func main() {
	cfg := config.Load()
	logger := logging.Setup(cfg.Log.Level, cfg.Log.File)
	defer logging.Destroy()

	logger.Info("initializing services", slog.String("env", cfg.App.Environment))

	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		logger.Warn("postgres unavailable, continuing", logging.Err(err))
	} else {
		defer db.Close()
		if err := db.HealthCheck(); err != nil {
			logger.Warn("postgres health check failed", logging.Err(err))
		} else {
			logger.Info("postgres connected and healthy")
		}
	}

	if !cfg.Feature.Redis {
		logger.Info("redis feature disabled, skipping")
	} else {
		redisCache, err := cache.NewRedisCache(&cfg.Redis)
		if err != nil {
			logger.Warn("redis unavailable, continuing", logging.Err(err))
		} else {
			defer redisCache.Close()
			ctx := context.Background()
			if err := redisCache.HealthCheck(ctx); err != nil {
				logger.Warn("redis health check failed", logging.Err(err))
			} else {
				logger.Info("redis connected and healthy")

				if err := redisCache.Set(ctx, "demo:key", "Hello from Redis!", time.Minute); err != nil {
					logger.Warn("redis demo set failed", logging.Err(err))
				}
				val, err := redisCache.Get(ctx, "demo:key")
				if err != nil {
					logger.Warn("redis demo get failed", logging.Err(err))
				} else {
					logger.Info("redis demo", slog.String("value", val))
				}
			}
		}
	}

	if !cfg.Feature.RabbitMQ {
		logger.Info("rabbitmq feature disabled, skipping")
	} else {
		broker, err := queuepkg.NewRabbitMQ(&cfg.RabbitMQ, logger)
		if err != nil {
			logger.Warn("rabbitmq unavailable, continuing", logging.Err(err))
		} else {
			defer broker.Close()
			if err := broker.HealthCheck(); err != nil {
				logger.Warn("rabbitmq health check failed", logging.Err(err))
			} else {
				logger.Info("rabbitmq connected and healthy")

				queueName := "demo_queue"
				if _, err := broker.DeclareQueue(queueName, true, false, false); err != nil {
					logger.Warn("rabbitmq declare queue failed", logging.Err(err))
				}
				if err := broker.PublishToQueue(context.Background(), queueName, []byte(`{"message":"Hello from RabbitMQ!"}`)); err != nil {
					logger.Warn("rabbitmq publish failed", logging.Err(err))
				} else {
					logger.Info("rabbitmq demo published", slog.String("queue", queueName))
				}
			}
		}
	}

	logger.Info("all enabled services initialized")
	logger.Info("see examples in the examples directory and docs in README.md")
}
