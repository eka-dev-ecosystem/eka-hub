package cache

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"golang/config"
	"golang/pkg/logging"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	Client *redis.Client
}

// NewRedisCache creates a new Redis cache connection
func NewRedisCache(cfg *config.RedisConfig) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Address(),
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	logging.Info("connected to Redis",
		slog.String("host", cfg.Host),
		slog.String("port", cfg.Port),
	)

	return &RedisCache{Client: client}, nil
}

// Set sets a key-value pair with expiration
func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.Client.Set(ctx, key, value, expiration).Err()
}

// Get retrieves a value by key
func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}

// Delete deletes a key
func (r *RedisCache) Delete(ctx context.Context, keys ...string) error {
	return r.Client.Del(ctx, keys...).Err()
}

// Exists checks if a key exists
func (r *RedisCache) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.Client.Exists(ctx, keys...).Result()
}

// SetNX sets a key-value pair only if it doesn't exist
func (r *RedisCache) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return r.Client.SetNX(ctx, key, value, expiration).Result()
}

// Increment increments the integer value of a key
func (r *RedisCache) Increment(ctx context.Context, key string) (int64, error) {
	return r.Client.Incr(ctx, key).Result()
}

// Expire sets a timeout on key
func (r *RedisCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.Client.Expire(ctx, key, expiration).Err()
}

// Close closes the Redis connection
func (r *RedisCache) Close() error {
	return r.Client.Close()
}

// HealthCheck checks if Redis connection is alive
func (r *RedisCache) HealthCheck(ctx context.Context) error {
	return r.Client.Ping(ctx).Err()
}
