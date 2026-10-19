package auth

import (
	"context"
	"strings"

	"golang/pkg/auth"
	customErrors "golang/pkg/errors"

	"github.com/gofiber/fiber/v2"
)

// bearerScheme is the Authorization header scheme used everywhere in the app.
const bearerScheme = "Bearer"

type userCtxKey struct{}

// ContextWithUser stores JWT claims in the context.
func ContextWithUser(ctx context.Context, claims *auth.JWTClaims) context.Context {
	return context.WithValue(ctx, userCtxKey{}, claims)
}

// UserFromContext extracts JWT claims from context.
func UserFromContext(ctx context.Context) (*auth.JWTClaims, bool) {
	claims, ok := ctx.Value(userCtxKey{}).(*auth.JWTClaims)
	return claims, ok && claims != nil
}

// UserIDFromContext extracts user ID from context.
func UserIDFromContext(ctx context.Context) (string, bool) {
	if claims, ok := UserFromContext(ctx); ok && claims.UserID != "" {
		return claims.UserID, true
	}
	return "", false
}

// AuthMiddleware creates a JWT authentication middleware
func AuthMiddleware(jwtService *auth.JWTService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var tokenString string
		authHeader := c.Get("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != bearerScheme {
				return customErrors.Unauthorized("Invalid authorization header format")
			}
			tokenString = parts[1]
		} else if qToken := c.Query("token"); qToken != "" {
			tokenString = qToken
		} else {
			return customErrors.Unauthorized("Authorization header is required")
		}

		// Validate access token
		claims, err := jwtService.ValidateToken(tokenString, auth.AccessToken)
		if err != nil {
			return err
		}

		// Check if user tokens are revoked
		if jwtService.IsUserRevoked(claims.UserID) {
			return customErrors.Unauthorized("Token has been revoked")
		}

		// Store claims in context
		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("token_id", claims.TokenID)
		c.Locals("claims", claims)
		c.SetUserContext(ContextWithUser(c.UserContext(), claims))

		return c.Next()
	}
}

// OptionalAuthMiddleware is similar to AuthMiddleware but doesn't fail if token is missing
// Useful for endpoints that work differently for authenticated users
func OptionalAuthMiddleware(jwtService *auth.JWTService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var tokenString string
		authHeader := c.Get("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == bearerScheme {
				tokenString = parts[1]
			}
		} else if qToken := c.Query("token"); qToken != "" {
			tokenString = qToken
		}

		if tokenString == "" {
			return c.Next()
		}

		claims, err := jwtService.ValidateToken(tokenString, auth.AccessToken)
		if err != nil {
			return c.Next()
		}

		if !jwtService.IsUserRevoked(claims.UserID) {
			c.Locals("user_id", claims.UserID)
			c.Locals("email", claims.Email)
			c.Locals("token_id", claims.TokenID)
			c.Locals("claims", claims)
			c.SetUserContext(ContextWithUser(c.UserContext(), claims))
		}

		return c.Next()
	}
}

// RefreshTokenMiddleware validates refresh token
func RefreshTokenMiddleware(jwtService *auth.JWTService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return customErrors.Unauthorized("Authorization header is required")
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != bearerScheme {
			return customErrors.Unauthorized("Invalid authorization header format")
		}

		tokenString := parts[1]

		// Validate refresh token
		claims, err := jwtService.ValidateToken(tokenString, auth.RefreshToken)
		if err != nil {
			return err
		}

		if jwtService.IsUserRevoked(claims.UserID) {
			return customErrors.Unauthorized("Token has been revoked")
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("token_id", claims.TokenID)
		c.Locals("claims", claims)
		c.Locals("refresh_token", tokenString)
		c.SetUserContext(ContextWithUser(c.UserContext(), claims))

		return c.Next()
	}
}

// GetUserID extracts user ID from context
func GetUserID(c *fiber.Ctx) (string, error) {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return "", customErrors.Unauthorized("User not authenticated")
	}
	return userID, nil
}

// GetEmail extracts email from context
func GetEmail(c *fiber.Ctx) (string, error) {
	email, ok := c.Locals("email").(string)
	if !ok || email == "" {
		return "", customErrors.Unauthorized("User not authenticated")
	}
	return email, nil
}

// GetClaims extracts JWT claims from context
func GetClaims(c *fiber.Ctx) (*auth.JWTClaims, error) {
	claims, ok := c.Locals("claims").(*auth.JWTClaims)
	if !ok {
		return nil, customErrors.Unauthorized("User not authenticated")
	}
	return claims, nil
}

// MustGetUserID extracts user ID or panics (use with recover middleware)
func MustGetUserID(c *fiber.Ctx) string {
	userID, err := GetUserID(c)
	if err != nil {
		panic(err)
	}
	return userID
}
