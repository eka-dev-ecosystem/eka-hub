package config

import (
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	cfg := Load()

	if cfg.App.Environment != "development" {
		t.Errorf("expected development env, got %q", cfg.App.Environment)
	}
	if cfg.Log.Level != "info" {
		t.Errorf("expected info level, got %q", cfg.Log.Level)
	}
	if cfg.Database.Host != "localhost" {
		t.Errorf("expected localhost db host, got %q", cfg.Database.Host)
	}
	if !cfg.Feature.Redis || !cfg.Feature.RabbitMQ || !cfg.Feature.Realtime {
		t.Errorf("all features expected enabled by default: %+v", cfg.Feature)
	}
	if !cfg.Feature.RealtimeSSE || !cfg.Feature.RealtimeWS {
		t.Errorf("realtime sse and ws expected enabled by default: %+v", cfg.Feature)
	}
	if cfg.Feature.RealtimeBridge != BridgeRedis {
		t.Errorf("expected realtime bridge redis by default, got %q", cfg.Feature.RealtimeBridge)
	}
}

func TestLoadEnvOverrides(t *testing.T) {
	t.Setenv("APP_NAME", "test-api")
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("LOG_FILE", "logs/app.jsonl")
	t.Setenv("DB_HOST", "db.example.com")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("DB_NAME", "custom_db")
	t.Setenv("REDIS_ENABLED", "false")
	t.Setenv("REALTIME_ENABLED", "false")
	t.Setenv("SERVER_PORT", "8080")

	cfg := Load()

	if cfg.App.Name != "test-api" {
		t.Errorf("expected APP_NAME=test-api, got %q", cfg.App.Name)
	}
	if !cfg.IsProduction() {
		t.Error("expected IsProduction()=true")
	}
	if cfg.Log.Level != "debug" || cfg.Log.File != "logs/app.jsonl" {
		t.Errorf("unexpected log config: %+v", cfg.Log)
	}
	if cfg.Database.Host != "db.example.com" || cfg.Database.Port != "5433" || cfg.Database.DBName != "custom_db" {
		t.Errorf("unexpected database config: %+v", cfg.Database)
	}
	if cfg.Feature.Redis != false || cfg.Feature.Realtime != false {
		t.Errorf("feature flags not honored: %+v", cfg.Feature)
	}
	if cfg.Feature.RealtimeSSE || cfg.Feature.RealtimeWS {
		t.Errorf("expected sse/ws forced off when realtime is disabled: %+v", cfg.Feature)
	}
	if cfg.Server.Port != "8080" {
		t.Errorf("expected server port 8080, got %q", cfg.Server.Port)
	}
}

func TestRealtimeBridgeFromEnv(t *testing.T) {
	t.Setenv("REALTIME_BRIDGE", "rabbitmq")
	if cfg := Load(); cfg.Feature.RealtimeBridge != BridgeRabbit {
		t.Errorf("expected rabbitmq bridge, got %q", cfg.Feature.RealtimeBridge)
	}

	t.Setenv("REALTIME_BRIDGE", "BOTH")
	if cfg := Load(); cfg.Feature.RealtimeBridge != BridgeBoth {
		t.Errorf("expected normalized lowercase both, got %q", cfg.Feature.RealtimeBridge)
	}

	t.Setenv("REALTIME_BRIDGE", "carrier-pigeon")
	if cfg := Load(); cfg.Feature.RealtimeBridge != BridgeRedis {
		t.Errorf("expected fallback to redis for invalid bridge, got %q", cfg.Feature.RealtimeBridge)
	}
}

func TestRealtimeEnabledForcesSSEAndWS(t *testing.T) {
	t.Setenv("REALTIME_SSE_ENABLED", "false")

	cfg := Load()
	if cfg.Feature.Realtime != true || cfg.Feature.RealtimeSSE != false || cfg.Feature.RealtimeWS != true {
		t.Errorf("expected independent sse/ws toggles, got %+v", cfg.Feature)
	}

	t.Setenv("REALTIME_ENABLED", "true")
	t.Setenv("REALTIME_WS_ENABLED", "false")
	cfg = Load()
	if cfg.Feature.RealtimeWS != false || cfg.Feature.RealtimeSSE != false {
		t.Errorf("expected ws disabled while sse stays disabled, got %+v", cfg.Feature)
	}
}

func TestFeatureSummary(t *testing.T) {
	f := FeatureConfig{Redis: true, RabbitMQ: false, Realtime: true, RealtimeSSE: true, RealtimeWS: false}
	got := f.Summary()
	if got["redis"] != true || got["rabbitmq"] != false || got["realtime"] != true {
		t.Errorf("unexpected summary: %#v", got)
	}
	if got["realtime_sse"] != true || got["realtime_ws"] != false {
		t.Errorf("unexpected realtime summary keys: %#v", got)
	}
}

func TestDatabaseDSN(t *testing.T) {
	cfg := DatabaseConfig{Host: "h", Port: "1", User: "u", Password: "p", DBName: "d", SSLMode: "disable"}
	dsn := cfg.DSN()
	want := "host=h port=1 user=u password=p dbname=d sslmode=disable"
	if dsn != want {
		t.Errorf("DSN = %q, want %q", dsn, want)
	}
}

func TestRedisAddress(t *testing.T) {
	cfg := RedisConfig{Host: "redis", Port: "6379"}
	if got := cfg.Address(); got != "redis:6379" {
		t.Errorf("Address() = %q, want redis:6379", got)
	}
}

func TestGetEnvAsBoolInvalid(t *testing.T) {
	t.Setenv("VAR_BOOL_TEST", "not-a-bool")
	if got := getEnvAsBool("VAR_BOOL_TEST", true); got != true {
		t.Errorf("expected fallback true for invalid bool, got %v", got)
	}
}
