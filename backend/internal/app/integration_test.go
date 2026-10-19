//go:build integration

// Package app_test contains real end-to-end integration tests. They boot the
// real application container (no mocks) against a live PostgreSQL instance and
// exercise the HTTP API the same way a client would.
//
// They are isolated behind the "integration" build tag so the regular
// `go test ./...` stays fast and database-free. Run them explicitly with:
//
//	go test -tags=integration ./internal/app/...
//
// Requirements:
//   - A reachable PostgreSQL server (override via TEST_DB_HOST, TEST_DB_PORT,
//     TEST_DB_NAME, TEST_DB_USER, TEST_DB_PASSWORD).
//   - Migrations applied against the test database:
//     go run ./cmd -action up   (with DB_* vars pointing at the test DB)
//
// If the database is unreachable the tests are skipped instead of failing.
package app_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"golang/config"
	"golang/database"
	"golang/database/migrate"
	"golang/internal/app"

	"github.com/gofiber/fiber/v2"
)

// ---------------------------------------------------------------------------
// Setup helpers
// ---------------------------------------------------------------------------

// loadCfg returns the database configuration used by the integration tests,
// honouring TEST_DB_* overrides. It also pushes those values into the
// environment so that app.Boot (which reads env) connects to the same DB.
func loadCfg(t *testing.T) *config.Config {
	t.Helper()
	cfg := config.Load()
	cfg.Database.Host = getEnv("TEST_DB_HOST", "127.0.0.1")
	cfg.Database.Port = getEnv("TEST_DB_PORT", "55432")
	cfg.Database.DBName = getEnv("TEST_DB_NAME", "golang_test")
	cfg.Database.User = getEnv("TEST_DB_USER", "postgres")
	cfg.Database.Password = getEnv("TEST_DB_PASSWORD", "")
	cfg.Database.SSLMode = "disable"

	// Optional integrations stay off unless a server is explicitly available
	// (TEST_REDIS_ENABLED=true). Realtime only needs the local hub.
	cfg.Feature.Redis = getEnvAsBool("TEST_REDIS_ENABLED", false)
	cfg.Feature.RabbitMQ = false
	cfg.Feature.Realtime = true

	t.Setenv("DB_HOST", cfg.Database.Host)
	t.Setenv("DB_PORT", cfg.Database.Port)
	t.Setenv("DB_NAME", cfg.Database.DBName)
	t.Setenv("DB_USER", cfg.Database.User)
	t.Setenv("DB_PASSWORD", cfg.Database.Password)
	t.Setenv("DB_SSLMODE", "disable")
	t.Setenv("REDIS_ENABLED", boolStr(cfg.Feature.Redis))
	t.Setenv("RABBITMQ_ENABLED", "false")
	t.Setenv("REALTIME_ENABLED", "true")

	return cfg
}

// dbReachable reports whether the configured PostgreSQL instance is reachable.
func dbReachable(t *testing.T, cfg *config.Config) bool {
	t.Helper()
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		return false
	}
	db.Close()
	return true
}

// startTestApp applies migrations, boots the real container and returns the
// Fiber application plus a shutdown function. Queries are sent through the
// in-process fiber.Test client.
func startTestApp(t *testing.T, cfg *config.Config) (*app.Container, *fiber.App) {
	t.Helper()

	if err := migrate.RunMigrations(&cfg.Database, "../../migrations"); err != nil {
		t.Fatalf("migrations failed: %v", err)
	}

	container, shutdown := app.Boot()
	t.Cleanup(func() { _ = shutdown() })

	return container, app.NewHTTP(container)
}

// resetUsers empties the users table between tests so each test starts clean.
func resetUsers(t *testing.T, container *app.Container) {
	t.Helper()
	if _, err := container.DB.DB.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate users: %v", err)
	}
}

// ---------------------------------------------------------------------------
// HTTP helpers
// ---------------------------------------------------------------------------

const jsonCT = "application/json"

// call performs an HTTP request against the Fiber test app.
func call(t *testing.T, httpApp *fiber.App, method, path, body, token string) *http.Response {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", jsonCT)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := httpApp.Test(req, 10000)
	if err != nil {
		t.Fatalf("request %s %s failed: %v", method, path, err)
	}
	return resp
}

func get(t *testing.T, httpApp *fiber.App, path string) *http.Response {
	t.Helper()
	return call(t, httpApp, http.MethodGet, path, "", "")
}

func post(t *testing.T, httpApp *fiber.App, path, body string) *http.Response {
	t.Helper()
	return call(t, httpApp, http.MethodPost, path, body, "")
}

func callAuth(t *testing.T, httpApp *fiber.App, method, path, body, token string) *http.Response {
	t.Helper()
	if token == "" {
		t.Fatalf("callAuth requires a token for %s %s", method, path)
	}
	return call(t, httpApp, method, path, body, token)
}

func requireStatus(t *testing.T, resp *http.Response, want int) {
	t.Helper()
	if resp.StatusCode != want {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status %d, got %d. body: %s", want, resp.StatusCode, string(body))
	}
}

// decodeJSON reads and unmarshals a JSON body into a generic map.
func decodeJSON(t *testing.T, reader io.Reader) map[string]interface{} {
	t.Helper()
	raw, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("failed to decode JSON %q: %v", string(raw), err)
	}
	return out
}

// dataOf returns the "data" member of a response as a map.
func dataOf(t *testing.T, payload map[string]interface{}) map[string]interface{} {
	t.Helper()
	data, _ := payload["data"].(map[string]interface{})
	return data
}

func requireCode(t *testing.T, payload map[string]interface{}, want string) {
	t.Helper()
	if got, _ := payload["code"].(string); got != want {
		t.Fatalf("expected response code %q, got %q (payload: %#v)", want, got, payload)
	}
}

func requireErrorCode(t *testing.T, resp *http.Response, want string) {
	t.Helper()
	defer resp.Body.Close()
	payload := decodeJSON(t, resp.Body)
	if success, _ := payload["success"].(bool); success {
		t.Fatalf("expected non-success response for %q, got success payload: %#v", want, payload)
	}
	if got, _ := payload["code"].(string); got != want {
		t.Fatalf("expected error code %q, got %q (payload: %#v)", want, got, payload)
	}
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestHealthEndpoint(t *testing.T) {
	cfg := loadCfg(t)
	if !dbReachable(t, cfg) {
		t.Skip("PostgreSQL not reachable; skipping integration test")
	}
	_, httpApp := startTestApp(t, cfg)

	resp := get(t, httpApp, "/health")
	defer resp.Body.Close()
	requireStatus(t, resp, http.StatusOK)

	payload := decodeJSON(t, resp.Body)
	if got, _ := payload["status"].(string); got != "healthy" {
		t.Fatalf("expected healthy status, got %#v", payload)
	}
	features, ok := payload["features"].(map[string]interface{})
	if !ok {
		t.Fatalf("health payload missing features map: %#v", payload)
	}
	if got := features["realtime"]; got != true {
		t.Errorf("expected realtime=true, got %v", got)
	}
	if got := features["redis"]; got != false {
		t.Errorf("expected redis=false (disabled in tests), got %v", got)
	}
}

func TestAuthLifecycle(t *testing.T) {
	cfg := loadCfg(t)
	if !dbReachable(t, cfg) {
		t.Skip("PostgreSQL not reachable; skipping integration test")
	}
	container, httpApp := startTestApp(t, cfg)
	resetUsers(t, container)

	email := fmt.Sprintf("user-%s@example.com", randSuffix())
	password := "SuperSecret1"

	// 1. Register.
	regResp := post(t, httpApp, "/api/v1/auth/register",
		fmt.Sprintf(`{"name":"Test User","email":%q,"password":%q}`, email, password))
	requireStatus(t, regResp, http.StatusCreated)
	reg := decodeJSON(t, regResp.Body)
	requireCode(t, reg, "CREATED")
	regData := dataOf(t, reg)
	tokens, _ := regData["tokens"].(map[string]interface{})
	accessToken, _ := tokens["access_token"].(string)
	refreshToken, _ := tokens["refresh_token"].(string)
	if accessToken == "" || refreshToken == "" {
		t.Fatalf("register response missing tokens: %#v", regData)
	}
	regResp.Body.Close()

	// 2. Login with wrong password -> 401 UNAUTHORIZED.
	badLogin := post(t, httpApp, "/api/v1/auth/login",
		fmt.Sprintf(`{"email":%q,"password":"wrong"}`, email))
	requireStatus(t, badLogin, http.StatusUnauthorized)
	requireErrorCode(t, badLogin, "UNAUTHORIZED")

	// 3. Login.
	loginResp := post(t, httpApp, "/api/v1/auth/login",
		fmt.Sprintf(`{"email":%q,"password":%q}`, email, password))
	requireStatus(t, loginResp, http.StatusOK)
	requireCode(t, decodeJSON(t, loginResp.Body), "SUCCESS")
	loginResp.Body.Close()

	// 4. Protected profile with the access token.
	profileResp := callAuth(t, httpApp, http.MethodGet, "/api/v1/auth/profile", "", accessToken)
	requireStatus(t, profileResp, http.StatusOK)
	profile := dataOf(t, decodeJSON(t, profileResp.Body))
	if got, _ := profile["email"].(string); got != email {
		t.Errorf("expected profile email %q, got %#v", email, profile)
	}
	profileResp.Body.Close()

	// 5. Refresh the access token.
	refreshResp := post(t, httpApp, "/api/v1/auth/refresh",
		fmt.Sprintf(`{"refresh_token":%q}`, refreshToken))
	requireStatus(t, refreshResp, http.StatusOK)
	refreshData := dataOf(t, decodeJSON(t, refreshResp.Body))
	newAccess, _ := refreshData["access_token"].(string)
	if newAccess == "" {
		t.Fatalf("refresh did not return a new access token: %#v", refreshData)
	}
	refreshResp.Body.Close()

	// 6. Update profile.
	updResp := callAuth(t, httpApp, http.MethodPut, "/api/v1/auth/profile", `{"name":"Updated Name"}`, newAccess)
	requireStatus(t, updResp, http.StatusOK)
	upd := dataOf(t, decodeJSON(t, updResp.Body))
	if got, _ := upd["name"].(string); got != "Updated Name" {
		t.Errorf("expected updated name, got %#v", upd)
	}
	updResp.Body.Close()

	// 7. Change password (revocation is logged, not fatal, without Redis).
	cpResp := callAuth(t, httpApp, http.MethodPost, "/api/v1/auth/change-password",
		fmt.Sprintf(`{"current_password":%q,"new_password":"NewPassword2!"}`, password), newAccess)
	requireStatus(t, cpResp, http.StatusOK)
	cpResp.Body.Close()

	// 8. Old password rejected after change.
	oldLogin := post(t, httpApp, "/api/v1/auth/login",
		fmt.Sprintf(`{"email":%q,"password":%q}`, email, password))
	requireStatus(t, oldLogin, http.StatusUnauthorized)
	requireErrorCode(t, oldLogin, "UNAUTHORIZED")

	// 9. Logout returns success (revocation unavailable without Redis is only
	// logged, not fatal).
	logoutResp := callAuth(t, httpApp, http.MethodPost, "/api/v1/auth/logout", `{}`, accessToken)
	requireStatus(t, logoutResp, http.StatusOK)
	logoutResp.Body.Close()
}

func TestUserCRUD(t *testing.T) {
	cfg := loadCfg(t)
	if !dbReachable(t, cfg) {
		t.Skip("PostgreSQL not reachable; skipping integration test")
	}
	container, httpApp := startTestApp(t, cfg)
	resetUsers(t, container)

	jsonBody := `{"name": "Alice", "email": "alice@example.com", "phone": "0812345678"}`

	// Create.
	createResp := call(t, httpApp, "POST", "/api/v1/users", jsonBody, "")
	requireStatus(t, createResp, http.StatusCreated)
	created := decodeJSON(t, createResp.Body)
	requireCode(t, created, "CREATED")
	id, _ := dataOf(t, created)["id"].(string)
	if id == "" {
		t.Fatalf("expected created user id, %#v", created)
	}
	createResp.Body.Close()

	// Duplicate email -> 409 (service pre-checks, so code is CONFLICT).
	dupResp := call(t, httpApp, "POST", "/api/v1/users", jsonBody, "")
	requireStatus(t, dupResp, http.StatusConflict)
	requireErrorCode(t, dupResp, "CONFLICT")

	// Validation failure -> 400 VALIDATION_ERROR.
	badResp := call(t, httpApp, "POST", "/api/v1/users", `{"name":"","email":"nope"}`, "")
	requireStatus(t, badResp, http.StatusBadRequest)
	requireErrorCode(t, badResp, "VALIDATION_ERROR")

	// Get by ID.
	getResp := get(t, httpApp, "/api/v1/users/"+id)
	requireStatus(t, getResp, http.StatusOK)
	got := dataOf(t, decodeJSON(t, getResp.Body))
	if got["email"] != "alice@example.com" {
		t.Errorf("expected alice@example.com, got %#v", got)
	}
	getResp.Body.Close()

	// Update.
	updResp := call(t, httpApp, "PUT", "/api/v1/users/"+id, `{"name":"Alice Updated"}`, "")
	requireStatus(t, updResp, http.StatusOK)
	upd := dataOf(t, decodeJSON(t, updResp.Body))
	if got, _ := upd["name"].(string); got != "Alice Updated" {
		t.Errorf("expected updated name, got %#v", upd)
	}
	updResp.Body.Close()

	// Soft delete.
	delResp := call(t, httpApp, "DELETE", "/api/v1/users/"+id, "", "")
	requireStatus(t, delResp, http.StatusOK)
	delResp.Body.Close()

	// Gone after deletion -> 404 NOT_FOUND.
	goneResp := get(t, httpApp, "/api/v1/users/"+id)
	requireStatus(t, goneResp, http.StatusNotFound)
	requireErrorCode(t, goneResp, "NOT_FOUND")

	// Invalid UUID -> 400 BAD_REQUEST.
	invalidResp := get(t, httpApp, "/api/v1/users/not-a-uuid")
	requireStatus(t, invalidResp, http.StatusBadRequest)
	requireErrorCode(t, invalidResp, "BAD_REQUEST")
}

func TestProtectedRouteRequiresAuth(t *testing.T) {
	cfg := loadCfg(t)
	if !dbReachable(t, cfg) {
		t.Skip("PostgreSQL not reachable; skipping integration test")
	}
	_, httpApp := startTestApp(t, cfg)

	for _, path := range []string{"/api/v1/auth/profile", "/api/v1/realtime/sse/stats", "/api/v1/protected/data"} {
		resp := get(t, httpApp, path)
		requireStatus(t, resp, http.StatusUnauthorized)
		requireErrorCode(t, resp, "UNAUTHORIZED")
	}
}

func TestRouteNotFound(t *testing.T) {
	cfg := loadCfg(t)
	if !dbReachable(t, cfg) {
		t.Skip("PostgreSQL not reachable; skipping integration test")
	}
	_, httpApp := startTestApp(t, cfg)
	resp := get(t, httpApp, "/api/v1/does-not-exist")
	requireStatus(t, resp, http.StatusNotFound)
	requireErrorCode(t, resp, "NOT_FOUND")
}

func TestRequestIDHeader(t *testing.T) {
	cfg := loadCfg(t)
	if !dbReachable(t, cfg) {
		t.Skip("PostgreSQL not reachable; skipping integration test")
	}
	_, httpApp := startTestApp(t, cfg)
	resp := get(t, httpApp, "/health")
	defer resp.Body.Close()
	requireStatus(t, resp, http.StatusOK)
	if rid := resp.Header.Get("X-Request-ID"); rid == "" {
		t.Error("expected X-Request-ID header on responses")
	}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("X-Request-ID", "my-custom-id")
	resp2, err := httpApp.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	if got := resp2.Header.Get("X-Request-ID"); got != "my-custom-id" {
		t.Errorf("expected echoed request id, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func randSuffix() string {
	const alpha = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	var seed uint64 = 0x0badc0de
	for i := len(b) - 1; i >= 0; i-- {
		seed = seed*6364136223846793005 + 1442695040888963407
		b[i] = alpha[(seed>>33)%uint64(len(alpha))]
	}
	return string(b)
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func getEnvAsBool(key string, def bool) bool {
	switch os.Getenv(key) {
	case "1", "true", "TRUE", "yes":
		return true
	case "0", "false", "FALSE", "no":
		return false
	}
	return def
}
