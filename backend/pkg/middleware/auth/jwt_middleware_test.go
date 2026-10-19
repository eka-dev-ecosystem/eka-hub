package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwtservice "golang/pkg/auth"
	customErrors "golang/pkg/errors"
	"golang/pkg/response"

	"github.com/gofiber/fiber/v2"
)

func newTestService() *jwtservice.JWTService {
	return jwtservice.NewJWTService(&jwtservice.JWTConfig{
		AccessSecret:         "test-access-secret-very-long",
		RefreshSecret:        "test-refresh-secret-very-long",
		AccessTokenDuration:  15 * time.Minute,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:               "test-issuer",
	}, nil)
}

func testApp() (*fiber.App, *jwtservice.JWTService) {
	svc := newTestService()

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if appErr, ok := err.(*customErrors.AppError); ok {
				return response.ErrorResponse(c, appErr.StatusCode, appErr.Message, appErr.Code)
			}
			return response.ErrorResponse(c, fiber.StatusInternalServerError, "internal", "INTERNAL_SERVER_ERROR")
		},
	})

	app.Use(AuthMiddleware(svc))
	app.Get("/whoami", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"user_id": c.Locals("user_id"),
			"email":   c.Locals("email"),
		})
	})
	return app, svc
}

func TestAuthMiddlewareWithoutToken(t *testing.T) {
	app, _ := testApp()
	req := httptest.NewRequest(http.MethodGet, "/whoami", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
}

func TestAuthMiddlewareWithValidToken(t *testing.T) {
	app, svc := testApp()
	pair, err := svc.GenerateTokenPair("user-42", "a@b.com")
	if err != nil {
		t.Fatalf("token generation failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/whoami", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	got := make([]byte, 0, 64)
	buf := make([]byte, 128)
	n, _ := resp.Body.Read(buf)
	got = append(got, buf[:n]...)
	if string(got) != `{"email":"a@b.com","user_id":"user-42"}` {
		t.Errorf("unexpected body: %s", string(got))
	}
}

func TestAuthMiddlewareRejectsWrongScheme(t *testing.T) {
	app, _ := testApp()
	req := httptest.NewRequest(http.MethodGet, "/whoami", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
}

func TestGetUserIDUnset(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if appErr, ok := err.(*customErrors.AppError); ok {
				return response.ErrorResponse(c, appErr.StatusCode, appErr.Message, appErr.Code)
			}
			return response.ErrorResponse(c, fiber.StatusInternalServerError, "internal", "INTERNAL_SERVER_ERROR")
		},
	})
	app.Get("/x", func(c *fiber.Ctx) error {
		if _, err := GetUserID(c); err != nil {
			return err
		}
		return c.SendStatus(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	resp, _ := app.Test(req)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
}
