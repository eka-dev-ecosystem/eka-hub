package usermodule

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	customErrors "golang/pkg/errors"
	"golang/pkg/response"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func testUserApp(repo UserRepository) *fiber.App {
	handler := NewUserHandler(newTestUserService(repo))
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if appErr, ok := err.(*customErrors.AppError); ok {
				return response.ErrorResponse(c, appErr.StatusCode, appErr.Message, appErr.Code)
			}
			if _, ok := err.(validator.ValidationErrors); ok {
				return response.ValidationErrorResponse(c, "validation failed")
			}
			return response.ErrorResponse(c, fiber.StatusInternalServerError, "internal", "INTERNAL_SERVER_ERROR")
		},
	})
	users := app.Group("/api/v1/users")
	users.Get("/", handler.GetUsers())
	users.Get("/:id", handler.GetUser())
	users.Post("/", handler.CreateUser())
	users.Put("/:id", handler.UpdateUser())
	users.Delete("/:id", handler.DeleteUser())
	return app
}

func doRequest(t *testing.T, app *fiber.App, method, path, body string) *http.Response {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request %s %s failed: %v", method, path, err)
	}
	return resp
}

func decodeResp(t *testing.T, resp *http.Response) map[string]interface{} {
	t.Helper()
	defer resp.Body.Close()
	var out map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return out
}

func TestHandlerGetUsers(t *testing.T) {
	app := testUserApp(newFakeRepo(seedUser("u1", "a@x.com")))
	resp := doRequest(t, app, "GET", "/api/v1/users/", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	payload := decodeResp(t, resp)
	if code, _ := payload["code"].(string); code != "SUCCESS" {
		t.Errorf("expected SUCCESS code, got %#v", payload)
	}
}

func TestHandlerCreateUserValidationError(t *testing.T) {
	app := testUserApp(newFakeRepo())
	resp := doRequest(t, app, "POST", "/api/v1/users/", `{"name":"","email":"bad"}`)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
	payload := decodeResp(t, resp)
	if code, _ := payload["code"].(string); code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR code, got %#v", payload)
	}
}

func TestHandlerCreateUserSuccess(t *testing.T) {
	app := testUserApp(newFakeRepo())
	resp := doRequest(t, app, "POST", "/api/v1/users/", `{"name":"Bob","email":"bob@x.com","phone":"0812345678"}`)
	if resp.StatusCode != http.StatusCreated {
		payload := decodeResp(t, resp)
		t.Fatalf("status = %d, want 201. payload: %#v", resp.StatusCode, payload)
	}
}

func TestHandlerGetUserNotFound(t *testing.T) {
	app := testUserApp(newFakeRepo())
	resp := doRequest(t, app, "GET", "/api/v1/users/"+uuid.New().String(), "")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	payload := decodeResp(t, resp)
	if code, _ := payload["code"].(string); code != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND code, got %#v", payload)
	}
}

func TestHandlerDeleteUser(t *testing.T) {
	id := uuid.New().String()
	app := testUserApp(newFakeRepo(seedUser(id, "a@x.com")))
	resp := doRequest(t, app, "DELETE", "/api/v1/users/"+id, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	payload := decodeResp(t, resp)
	if success, _ := payload["success"].(bool); !success {
		t.Errorf("expected success=true, got %#v", payload)
	}
}
