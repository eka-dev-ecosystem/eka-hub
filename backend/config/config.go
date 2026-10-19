package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

// Realtime bridge transport modes (REALTIME_BRIDGE). The realtime hub uses
// these to fan messages out across instances.
const (
	BridgeRedis  = "redis"
	BridgeRabbit = "rabbitmq"
	BridgeBoth   = "both"
)

// Config aggregates every configuration section for the application.
type Config struct {
	App      AppConfig
	Server   ServerConfig
	Log      LogConfig
	Feature  FeatureConfig
	Database DatabaseConfig
	Redis    RedisConfig
	RabbitMQ RabbitMQConfig
	JWT      JWTConfig
}

// AppConfig holds generic application metadata.
type AppConfig struct {
	Name        string
	Environment string
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Host string
	Port string
}

// LogConfig controls the structured logger. Output is always JSON.
type LogConfig struct {
	Level string // debug | info | warn | error
	File  string // optional path to a JSON log file (e.g. logs/app.log)
}

// FeatureConfig turns optional integrations on/off via environment
// variables (e.g. RABBITMQ_ENABLED=false). Disabled features are never
// initialized, so the application runs even when the backing service is
// unavailable.
type FeatureConfig struct {
	Redis    bool
	RabbitMQ bool

	// Realtime is the master switch for the realtime module. When it is off,
	// RealtimeSSE and RealtimeWS are forced off regardless of their own flags.
	Realtime bool
	// RealtimeSSE / RealtimeWS activate Server-Sent Events and WebSocket
	// independently so one protocol can run without the other.
	RealtimeSSE bool
	RealtimeWS  bool
	// RealtimeBridge selects the distributed fan-out transport for the
	// realtime hub across instances: "redis" (default), "rabbitmq" or "both".
	RealtimeBridge string
}

// Summary returns the enabled/disabled status of every optional feature, used
// for logging and the /health endpoint.
func (f *FeatureConfig) Summary() map[string]bool {
	return map[string]bool{
		"redis":        f.Redis,
		"rabbitmq":     f.RabbitMQ,
		"realtime":     f.Realtime,
		"realtime_sse": f.RealtimeSSE,
		"realtime_ws":  f.RealtimeWS,
	}
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	DB       int
}

type RabbitMQConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Vhost    string
}

type JWTConfig struct {
	AccessSecret         string
	RefreshSecret        string
	AccessTokenDuration  int // in minutes
	RefreshTokenDuration int // in hours
	Issuer               string
}

func Load() *Config {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	_ = viper.ReadInConfig()

	cfg := &Config{
		App: AppConfig{
			Name:        getEnv("APP_NAME", "golang-api"),
			Environment: getEnv("ENVIRONMENT", "development"),
		},
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnv("SERVER_PORT", "3000"),
		},
		Log: LogConfig{
			Level: getEnv("LOG_LEVEL", "info"),
			File:  getEnv("LOG_FILE", ""),
		},
		Feature: FeatureConfig{
			Redis:          getEnvAsBool("REDIS_ENABLED", true),
			RabbitMQ:       getEnvAsBool("RABBITMQ_ENABLED", true),
			Realtime:       getEnvAsBool("REALTIME_ENABLED", true),
			RealtimeSSE:    getEnvAsBool("REALTIME_SSE_ENABLED", true),
			RealtimeWS:     getEnvAsBool("REALTIME_WS_ENABLED", true),
			RealtimeBridge: strings.ToLower(getEnv("REALTIME_BRIDGE", "redis")),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "mydb"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Username: getEnv("REDIS_USERNAME", ""),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		RabbitMQ: RabbitMQConfig{
			Host:     getEnv("RABBITMQ_HOST", "localhost"),
			Port:     getEnv("RABBITMQ_PORT", "5672"),
			User:     getEnv("RABBITMQ_USER", "guest"),
			Password: getEnv("RABBITMQ_PASSWORD", "guest"),
			Vhost:    getEnv("RABBITMQ_VHOST", "/"),
		},
		JWT: JWTConfig{
			AccessSecret:         getEnv("JWT_ACCESS_SECRET", "your-access-secret-key-change-this"),
			RefreshSecret:        getEnv("JWT_REFRESH_SECRET", "your-refresh-secret-key-change-this"),
			AccessTokenDuration:  getEnvAsInt("JWT_ACCESS_DURATION", 15),   // 15 minutes
			RefreshTokenDuration: getEnvAsInt("JWT_REFRESH_DURATION", 168), // 7 days (168 hours)
			Issuer:               getEnv("JWT_ISSUER", "my-api"),
		},
	}

	cfg.normalize()
	return cfg
}

// normalize coerces feature flags into a consistent state: the realtime master
// switch gates SSE/WS, and the bridge mode is clamped to the known transports.
func (c *Config) normalize() {
	if !c.Feature.Realtime {
		c.Feature.RealtimeSSE = false
		c.Feature.RealtimeWS = false
	}
	switch c.Feature.RealtimeBridge {
	case BridgeRedis, BridgeRabbit, BridgeBoth:
	default:
		c.Feature.RealtimeBridge = BridgeRedis
	}
}

func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode)
}

func (r *RedisConfig) Address() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

func (rmq *RabbitMQConfig) URL() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s%s",
		rmq.User, rmq.Password, rmq.Host, rmq.Port, rmq.Vhost)
}

func (c *Config) IsProduction() bool {
	return strings.EqualFold(c.App.Environment, "production")
}

// CORSOrigins returns the comma-separated list of allowed origins or a safe
// default for development.
func (c *Config) CORSOrigins() string {
	if origins := getEnv("CORS_ORIGINS", ""); origins != "" {
		return origins
	}
	return "http://localhost:3000,http://localhost:5173"
}

func getEnv(key, defaultValue string) string {
	viper.BindEnv(key)
	if value := viper.GetString(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	viper.BindEnv(key)
	if viper.IsSet(key) {
		return viper.GetInt(key)
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	viper.BindEnv(key)
	if viper.IsSet(key) {
		val := viper.GetString(key)
		b, err := strconv.ParseBool(val)
		if err != nil {
			return defaultValue
		}
		return b
	}
	return defaultValue
}
