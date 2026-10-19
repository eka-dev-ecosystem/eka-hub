package authmodule

import (
	"golang/pkg/errors"
	"golang/pkg/logging"
	"golang/pkg/response"
	"golang/pkg/validation"

	"github.com/gofiber/fiber/v2"
)

// AuthHandler exposes the auth endpoints.
type AuthHandler struct {
	service *AuthService
}

// NewAuthHandler creates the auth HTTP handler.
func NewAuthHandler(service *AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// Register handles POST /api/v1/auth/register.
func (h *AuthHandler) Register() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req RegisterRequest
		if err := validation.Default.BindAndValidate(c, &req); err != nil {
			return err
		}

		res, err := h.service.Register(c.UserContext(), &req)
		if err != nil {
			return err
		}

		logging.LoggerFromFiber(c).Info("user registered", "email", req.Email)
		return response.CreatedResponse(c, res, "User registered successfully")
	}
}

// Login handles POST /api/v1/auth/login.
func (h *AuthHandler) Login() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req LoginRequest
		if err := validation.Default.BindAndValidate(c, &req); err != nil {
			return err
		}

		res, err := h.service.Login(c.UserContext(), &req)
		if err != nil {
			return err
		}

		logging.LoggerFromFiber(c).Info("user logged in", "user_id", res.User.ID)
		return response.SuccessResponse(c, res, "Login successful")
	}
}

// RefreshToken handles POST /api/v1/auth/refresh.
func (h *AuthHandler) RefreshToken() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req RefreshTokenRequest
		if err := c.BodyParser(&req); err != nil || req.RefreshToken == "" {
			if token, ok := c.Locals("refresh_token").(string); ok && token != "" {
				req.RefreshToken = token
			} else {
				return errors.Unauthorized("Refresh token is required")
			}
		}

		tokens, err := h.service.RefreshToken(c.UserContext(), req.RefreshToken)
		if err != nil {
			return err
		}

		return response.SuccessResponse(c, tokens, "Token refreshed successfully")
	}
}

// Logout handles POST /api/v1/auth/logout.
func (h *AuthHandler) Logout() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return errors.Unauthorized("User not authenticated")
		}

		token := extractBearer(c)
		if token != "" {
			if err := h.service.jwtService.RevokeToken(token, "access"); err != nil {
				logging.LoggerFromFiber(c).Warn("failed to revoke token on logout", logging.Err(err))
			}
		}

		return response.SuccessResponse(c, fiber.Map{"user_id": userID}, "Logout successful")
	}
}

// LogoutAll handles POST /api/v1/auth/logout-all.
func (h *AuthHandler) LogoutAll() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return errors.Unauthorized("User not authenticated")
		}

		if err := h.service.jwtService.RevokeAllUserTokens(userID); err != nil {
			return err
		}

		return response.SuccessResponse(c, fiber.Map{"user_id": userID}, "Logout successful")
	}
}

// GetProfile handles GET /api/v1/auth/profile.
func (h *AuthHandler) GetProfile() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return errors.Unauthorized("User not authenticated")
		}

		user, err := h.service.GetCurrentUser(c.UserContext(), userID)
		if err != nil {
			return err
		}

		return response.SuccessResponse(c, user, "Profile retrieved successfully")
	}
}

// UpdateProfile handles PUT /api/v1/auth/profile.
func (h *AuthHandler) UpdateProfile() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return errors.Unauthorized("User not authenticated")
		}

		var req UpdateProfileRequest
		if err := validation.Default.BindAndValidate(c, &req); err != nil {
			return err
		}

		user, err := h.service.UpdateProfile(c.UserContext(), userID, &req)
		if err != nil {
			return err
		}

		return response.SuccessResponse(c, user, "Profile updated successfully")
	}
}

// ChangePassword handles POST /api/v1/auth/change-password.
func (h *AuthHandler) ChangePassword() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return errors.Unauthorized("User not authenticated")
		}

		var req ChangePasswordRequest
		if err := validation.Default.BindAndValidate(c, &req); err != nil {
			return err
		}

		if err := h.service.ChangePassword(c.UserContext(), userID, &req); err != nil {
			return err
		}

		return response.SuccessResponse(c, fiber.Map{}, "Password changed successfully. Please login again.")
	}
}

// extractBearer strips the "Bearer " prefix from the Authorization header.
func extractBearer(c *fiber.Ctx) string {
	header := c.Get("Authorization")
	if len(header) > 7 && header[:7] == "Bearer " {
		return header[7:]
	}
	return ""
}
