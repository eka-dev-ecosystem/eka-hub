# JWT Authentication System

Complete JWT authentication system dengan access token dan refresh token flow.

## 🎯 Features

- ✅ **Access Token & Refresh Token** - Dual token system
- ✅ **Token Revocation** - Redis-based token blacklisting
- ✅ **Password Hashing** - Bcrypt encryption
- ✅ **JWT Middleware** - Protect routes easily
- ✅ **User Management** - Register, Login, Logout
- ✅ **Profile Management** - Get/Update profile, Change password
- ✅ **Logout All Devices** - Revoke all user tokens
- ✅ **Optional Auth** - Routes that work for both auth/non-auth users

## 📦 Components

### 1. JWT Service (`pkg/auth/jwt.go`)

Core JWT operations:

- Generate token pair (access + refresh)
- Validate tokens
- Refresh access token
- Revoke tokens
- User-level revocation

### 2. JWT Middleware (`pkg/middleware/auth/jwt_middleware.go`)

- `AuthMiddleware` - Require authentication
- `OptionalAuthMiddleware` - Optional authentication
- `RefreshTokenMiddleware` - Validate refresh token
- Helper functions: `GetUserID`, `GetEmail`, `GetClaims`

### 3. Auth Handler (`pkg/handlers/auth_handler.go`)

HTTP endpoints:

- POST `/auth/register` - User registration
- POST `/auth/login` - User login
- POST `/auth/refresh` - Refresh access token
- POST `/auth/logout` - Logout (revoke token)
- POST `/auth/logout-all` - Logout all devices
- GET `/auth/profile` - Get user profile
- PUT `/auth/profile` - Update profile
- POST `/auth/change-password` - Change password

## 🔧 Configuration

### Environment Variables

Add to your `.env` file:

```env
# JWT Configuration
JWT_ACCESS_SECRET=your-super-secret-access-key-change-this-in-production
JWT_REFRESH_SECRET=your-super-secret-refresh-key-change-this-in-production
JWT_ACCESS_DURATION=15    # in minutes
JWT_REFRESH_DURATION=168  # in hours (7 days)
JWT_ISSUER=my-api
```

### Security Best Practices

1. **Strong Secrets**: Use at least 32 characters for JWT secrets
2. **Different Secrets**: Use different secrets for access and refresh tokens
3. **Short Access Token**: Keep access token duration short (15-30 minutes)
4. **Long Refresh Token**: Refresh token can be longer (7-30 days)

## 🚀 Quick Start

### 1. Run Migrations

```bash
migrate.bat up
```

### 2. Start Server

```bash
go run cmd/server/main.go
```

## 📝 API Usage Examples

### Register New User

```bash
POST /api/v1/auth/register
Content-Type: application/json

{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "SecurePassword123"
}
```

**Response:**

```json
{
  "message": "User registered successfully",
  "success": true,
  "data": {
    "user": {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "name": "John Doe",
      "email": "john@example.com",
      "created_at": "2026-01-17T10:30:00Z"
    },
    "tokens": {
      "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "expires_at": "2026-01-17T10:45:00Z",
      "token_type": "Bearer"
    }
  },
  "code": "CREATED",
  "timestamp": "2026-01-17T10:30:00Z"
}
```

### Login

```bash
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "john@example.com",
  "password": "SecurePassword123"
}
```

**Response:**

```json
{
  "message": "Login successful",
  "success": true,
  "data": {
    "user": {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "name": "John Doe",
      "email": "john@example.com",
      "created_at": "2026-01-17T10:30:00Z"
    },
    "tokens": {
      "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "expires_at": "2026-01-17T10:45:00Z",
      "token_type": "Bearer"
    }
  },
  "code": "SUCCESS",
  "timestamp": "2026-01-17T10:30:00Z"
}
```

### Refresh Access Token

```bash
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response:**

```json
{
  "message": "Token refreshed successfully",
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2026-01-17T11:00:00Z",
    "token_type": "Bearer"
  },
  "code": "SUCCESS",
  "timestamp": "2026-01-17T10:45:00Z"
}
```

### Get Profile (Protected Route)

```bash
GET /api/v1/auth/profile
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response:**

```json
{
  "message": "Profile retrieved successfully",
  "success": true,
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "John Doe",
    "email": "john@example.com",
    "created_at": "2026-01-17T10:30:00Z"
  },
  "code": "SUCCESS",
  "timestamp": "2026-01-17T10:45:00Z"
}
```

### Update Profile

```bash
PUT /api/v1/auth/profile
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json

{
  "name": "John Smith"
}
```

### Change Password

```bash
POST /api/v1/auth/change-password
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json

{
  "current_password": "OldPassword123",
  "new_password": "NewPassword456"
}
```

### Logout

```bash
POST /api/v1/auth/logout
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Logout All Devices

```bash
POST /api/v1/auth/logout-all
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

## 💻 Code Examples

### Protecting Routes

```go
package main

import (
	"github.com/gofiber/fiber/v2"
	"golang/pkg/auth"
	authMiddleware "golang/pkg/middleware/auth"
)

func setupRoutes(app *fiber.App, jwtService *auth.JWTService) {
	api := app.Group("/api/v1")

	// Public route
	api.Get("/public", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "This is public"})
	})

	// Protected routes
	protected := api.Group("/protected")
	protected.Use(authMiddleware.AuthMiddleware(jwtService))
	{
		protected.Get("/data", func(c *fiber.Ctx) error {
			userID, _ := authMiddleware.GetUserID(c)
			return c.JSON(fiber.Map{
				"message": "Protected data",
				"user_id": userID,
			})
		})
	}
}
```

### Using Optional Auth

```go
// Route that works differently for authenticated users
api.Get("/posts", authMiddleware.OptionalAuthMiddleware(jwtService), func (c *fiber.Ctx) error {
userID, err := authMiddleware.GetUserID(c)

if err != nil {
// User not authenticated - show public posts
return c.JSON(fiber.Map{
"posts": getPublicPosts(),
})
}

// User authenticated - show personalized posts
return c.JSON(fiber.Map{
"posts": getPersonalizedPosts(userID),
})
})
```

### Extracting User Info

```go
func MyProtectedHandler(c *fiber.Ctx) error {
// Get user ID
userID, err := authMiddleware.GetUserID(c)
if err != nil {
return err
}

// Get email
email, err := authMiddleware.GetEmail(c)
if err != nil {
return err
}

// Get full claims
claims, err := authMiddleware.GetClaims(c)
if err != nil {
return err
}

return c.JSON(fiber.Map{
"user_id":  userID,
"email":    email,
"token_id": claims.TokenID,
})
}
```

### Custom Authentication Handler

```go
func MyCustomAuthHandler(db *database.PostgresDB, jwtService *auth.JWTService) fiber.Handler {
return func (c *fiber.Ctx) error {
// Your login logic
email := "user@example.com"
userID := "user-uuid"

// Generate tokens
tokens, err := jwtService.GenerateTokenPair(userID, email)
if err != nil {
return customErrors.InternalError("Failed to generate tokens", err)
}

return c.JSON(fiber.Map{
"user": fiber.Map{"id": userID, "email": email},
"tokens": tokens,
})
}
}
```

## 🔐 Token Flow

### Authentication Flow

```
1. User registers/logs in
   ↓
2. Server generates access token (15 min) + refresh token (7 days)
   ↓
3. Server stores token IDs in Redis
   ↓
4. Client receives both tokens
   ↓
5. Client uses access token for API requests
```

### Token Refresh Flow

```
1. Access token expires
   ↓
2. Client sends refresh token to /auth/refresh
   ↓
3. Server validates refresh token
   ↓
4. Server generates new token pair
   ↓
5. Server revokes old tokens (optional)
   ↓
6. Client receives new tokens
```

### Token Revocation Flow

```
1. User logs out or changes password
   ↓
2. Server removes token ID from Redis
   ↓
3. Token becomes invalid immediately
   ↓
4. User must login again
```

## 🛡️ Security Features

### 1. Token Revocation

- Tokens stored in Redis with expiration
- Instant revocation by deleting from Redis
- No need to wait for token expiration

### 2. Password Security

- Bcrypt hashing with default cost (10)
- Automatic salt generation
- Secure password comparison

### 3. User-Level Revocation

- Revoke all tokens for a user
- Useful for "logout all devices"
- Useful for security incidents

### 4. Token Validation

- Signature verification
- Expiration check
- Token type validation
- Revocation status check

## 🔍 Error Responses

### Invalid Token

```json
{
  "message": "Invalid token",
  "success": false,
  "data": null,
  "code": "UNAUTHORIZED",
  "timestamp": "2026-01-17T10:45:00Z"
}
```

### Missing Authorization Header

```json
{
  "message": "Authorization header is required",
  "success": false,
  "data": null,
  "code": "UNAUTHORIZED",
  "timestamp": "2026-01-17T10:45:00Z"
}
```

### Expired Token

```json
{
  "message": "Invalid token",
  "success": false,
  "data": null,
  "code": "UNAUTHORIZED",
  "timestamp": "2026-01-17T10:45:00Z"
}
```

### Revoked Token

```json
{
  "message": "Token has been revoked",
  "success": false,
  "data": null,
  "code": "UNAUTHORIZED",
  "timestamp": "2026-01-17T10:45:00Z"
}
```

## 📊 Token Structure

### Access Token Claims

```json
{
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "email": "john@example.com",
  "token_type": "access",
  "token_id": "550e8400-e29b-41d4-a716-446655440000",
  "exp": 1705489200,
  "iat": 1705488300,
  "nbf": 1705488300,
  "iss": "my-api",
  "sub": "123e4567-e89b-12d3-a456-426614174000",
  "jti": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Refresh Token Claims

Same structure but with `token_type: "refresh"` and longer expiration.

## 🧪 Testing

### Test with cURL

```bash
# Register
curl -X POST http://localhost:3000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com","password":"SecurePass123"}'

# Login
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"SecurePass123"}'

# Get Profile (replace TOKEN with your access token)
curl -X GET http://localhost:3000/api/v1/auth/profile \
  -H "Authorization: Bearer TOKEN"

# Refresh Token
curl -X POST http://localhost:3000/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"REFRESH_TOKEN"}'

# Logout
curl -X POST http://localhost:3000/api/v1/auth/logout \
  -H "Authorization: Bearer TOKEN"
```

## ✅ Checklist

- [x] JWT service implementation
- [x] Access token generation
- [x] Refresh token generation
- [x] Token validation
- [x] Token revocation (Redis-based)
- [x] User-level revocation
- [x] Auth middleware
- [x] Optional auth middleware
- [x] Refresh token middleware
- [x] Auth handlers (register, login, logout, etc.)
- [x] Password hashing (bcrypt)
- [x] Profile management
- [x] Change password
- [x] Configuration
- [x] Migrations
- [x] Documentation

## 🎉 Ready to Use!

JWT authentication system dengan access token dan refresh token flow sudah **100% complete**!

**Happy coding!** 🚀

