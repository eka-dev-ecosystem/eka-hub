# 🎉 Error Handling System - Complete!

## ✅ Yang Telah Dibuat

### 📁 File Structure

```
pkg/
├── response/
│   └── response.go                    ✅ Base response & helper functions
├── errors/
│   └── errors.go                      ✅ Custom errors & PostgreSQL parser
├── constants/
│   └── database.go                    ✅ Error codes & constraint mappings
├── middleware/
│   └── error_handler.go               ✅ Global error handler middleware
└── handlers/
    └── user_handler.go                ✅ Example handlers with error handling

cmd/
└── server/
    └── main.go                        ✅ Server setup with error handler

Documentation:
└── ERROR_HANDLING.md                  ✅ Complete documentation
```

## 🚀 Features

### 1. **Base Response Structure (TypeScript-like)**

```go
type BaseResponse[T any] struct {
    Message   string    `json:"message"`
    Success   bool      `json:"success"`
    Data      *T        `json:"data"`
    Code      string    `json:"code"`
    Timestamp string    `json:"timestamp"`
}
```

**Sesuai dengan TypeScript:**

```typescript
export type BaseResponse<T = any> = {
    message: string
    success: boolean
    data: T | null
    code: string
    timestamp: string
}
```

### 2. **PostgreSQL Error Handling**

#### ✅ Unique Constraint Violations

**Constraint Mapping:**

```go
var UniqueConstraintFieldMap = map[string]string{
    "skills_name_key": "name",
    "users_email_key": "email",
}
```

**Error Response:**

```json
{
  "message": "The email already exists",
  "success": false,
  "data": null,
  "code": "DUPLICATE_ENTRY",
  "timestamp": "2026-01-17T10:30:00Z"
}
```

#### ✅ Foreign Key Violations

**Constraint Mapping:**

```go
var ForeignKeyConstraintMap = map[string]string{
    "orders_user_id_fkey": "user",
    "products_category_id_fkey": "category",
}
```

**Delete Error Response:**

```json
{
  "message": "Cannot delete this record because it is still referenced by order",
  "success": false,
  "data": null,
  "code": "FOREIGN_KEY_VIOLATION",
  "timestamp": "2026-01-17T10:30:00Z"
}
```

**Insert/Update Error Response:**

```json
{
  "message": "The referenced user does not exist",
  "success": false,
  "data": null,
  "code": "FOREIGN_KEY_VIOLATION",
  "timestamp": "2026-01-17T10:30:00Z"
}
```

### 3. **Automatic Error Detection**

System otomatis mendeteksi dan mem-parse:

- ✅ PostgreSQL unique violations (23505)
- ✅ Foreign key violations (23503)
- ✅ Not null violations (23502)
- ✅ Check constraint violations (23514)
- ✅ Validation errors
- ✅ Not found errors (sql.ErrNoRows)
- ✅ Fiber errors
- ✅ Custom app errors

### 4. **Validation Support**

```go
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required,min=3"`
    Email string `json:"email" validate:"required,email"`
}
```

**Validation Error Response:**

```json
{
  "message": "Validation failed",
  "success": false,
  "data": [
    {
      "field": "Name",
      "tag": "min",
      "value": "3",
      "message": "Name must be at least 3 characters"
    }
  ],
  "code": "VALIDATION_ERROR",
  "timestamp": "2026-01-17T10:30:00Z"
}
```

### 5. **Response Helpers**

```go
// Success responses
response.SuccessResponse(c, data, "Success message")
response.CreatedResponse(c, data, "Created message")
response.PaginatedSuccessResponse(c, paginatedData, "Message")

// Error responses
response.ErrorResponse(c, statusCode, message, code)
response.ValidationErrorResponse(c, validationErrors)

// Custom errors
customErrors.BadRequest("Message")
customErrors.Unauthorized("Message")
customErrors.Forbidden("Message")
customErrors.NotFound("Message")
customErrors.Conflict("Message")
customErrors.DatabaseError("Message", err)
customErrors.ParseDatabaseError(err) // Auto-parse PostgreSQL errors
```

## 💻 Usage Example

### 1. Setup Server

```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "golang/pkg/middleware"
)

func main() {
    app := fiber.New(fiber.Config{
        ErrorHandler: middleware.ErrorHandler(), // Global error handler
    })
    
    app.Use(middleware.RecoverMiddleware()) // Panic recovery
    
    // Your routes...
    
    app.Listen(":3000")
}
```

### 2. Handler Example

```go
func CreateUser(db *database.PostgresDB) fiber.Handler {
    return func(c *fiber.Ctx) error {
        var req CreateUserRequest
        
        if err := c.BodyParser(&req); err != nil {
            return customErrors.BadRequest("Invalid request")
        }

        if err := validate.Struct(req); err != nil {
            return err // Auto-formatted by middleware
        }

        _, err := db.DB.Exec("INSERT INTO users...", req.Name, req.Email)
        if err != nil {
            // Automatically detects duplicate email, foreign key errors, etc.
            return customErrors.ParseDatabaseError(err)
        }

        return response.CreatedResponse(c, user, "User created")
    }
}
```

## 🔧 Configuration

### Add Your Constraints

Edit `pkg/constants/database.go`:

```go
var UniqueConstraintFieldMap = map[string]string{
    "skills_name_key":       "name",
    "users_email_key":       "email",
    "users_username_key":    "username",
    "products_sku_key":      "sku",
    // Add more...
}

var ForeignKeyConstraintMap = map[string]string{
    "users_role_id_fkey":    "role",
    "orders_user_id_fkey":   "user",
    // Add more...
}
```

## 📊 Error Codes

| Code                    | Status  | Description                 |
|-------------------------|---------|-----------------------------|
| `SUCCESS`               | 200     | Operation successful        |
| `CREATED`               | 201     | Resource created            |
| `BAD_REQUEST`           | 400     | Invalid request             |
| `VALIDATION_ERROR`      | 422     | Validation failed           |
| `UNAUTHORIZED`          | 401     | Authentication required     |
| `FORBIDDEN`             | 403     | Permission denied           |
| `NOT_FOUND`             | 404     | Resource not found          |
| `CONFLICT`              | 409     | Resource conflict           |
| `DUPLICATE_ENTRY`       | 409     | Unique constraint violation |
| `FOREIGN_KEY_VIOLATION` | 409/422 | Foreign key error           |
| `INTERNAL_SERVER_ERROR` | 500     | Server error                |
| `DATABASE_ERROR`        | 500     | Database error              |

## 🧪 Test Examples

### Test Duplicate Email

```bash
curl -X POST http://localhost:3000/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"John","email":"existing@example.com"}'
```

**Response:**

```json
{
  "message": "The email already exists",
  "success": false,
  "data": null,
  "code": "DUPLICATE_ENTRY",
  "timestamp": "2026-01-17T10:30:00Z"
}
```

### Test Validation

```bash
curl -X POST http://localhost:3000/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Jo","email":"invalid"}'
```

**Response:**

```json
{
  "message": "Validation failed",
  "success": false,
  "data": [
    {
      "field": "Name",
      "tag": "min",
      "value": "3",
      "message": "Name must be at least 3 characters"
    },
    {
      "field": "Email",
      "tag": "email",
      "value": "",
      "message": "Email must be a valid email address"
    }
  ],
  "code": "VALIDATION_ERROR",
  "timestamp": "2026-01-17T10:30:00Z"
}
```

### Test Foreign Key Violation

```bash
curl -X DELETE http://localhost:3000/api/v1/users/123
```

**Response (if referenced):**

```json
{
  "message": "Cannot delete this record because it is still referenced by order",
  "success": false,
  "data": null,
  "code": "FOREIGN_KEY_VIOLATION",
  "timestamp": "2026-01-17T10:30:00Z"
}
```

## 📝 Migration with Constraints

```sql
-- migrations/000003_create_users.up.sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,  -- creates users_email_key
    role_id UUID NOT NULL,
    
    CONSTRAINT users_role_id_fkey 
        FOREIGN KEY (role_id) 
        REFERENCES roles(id)
);
```

## 🎯 Error Flow Diagram

```
Request
  ↓
Handler
  ↓
Error Occurred?
  ↓
Middleware ErrorHandler
  ↓
├─ AppError? → Format & Return
├─ ValidationError? → Format & Return  
├─ sql.ErrNoRows? → Return 404
├─ PostgreSQL Error?
│   ├─ 23505 (unique) → Parse constraint → Return friendly message
│   ├─ 23503 (FK) → Parse constraint → Return friendly message
│   └─ Other → Return database error
└─ Other → Return 500 Internal Error
```

## ✅ Comparison dengan TypeScript

### TypeScript (server/constants/pgConstraints.ts)

```typescript
export const UNIQUE_CONSTRAINT_FIELD_MAP: Record<string, string> = {
    skills_name_key: 'name',
    users_email_key: 'email',
}
```

### Go (pkg/constants/database.go)

```go
var UniqueConstraintFieldMap = map[string]string{
    "skills_name_key": "name",
    "users_email_key": "email",
}
```

**✅ Sama persis! Hanya syntax yang berbeda.**

## 🚀 Quick Start

1. **Start server:**

```bash
go run cmd/server/main.go
```

2. **Update constraints** di `pkg/constants/database.go`

3. **Create handlers** menggunakan pattern di `pkg/handlers/user_handler.go`

4. **Test error responses** untuk setiap scenario

## 📚 Documentation

- `ERROR_HANDLING.md` - Complete error handling guide
- `pkg/response/response.go` - Response helpers
- `pkg/errors/errors.go` - Error types & parser
- `pkg/constants/database.go` - Error codes & mappings
- `pkg/handlers/user_handler.go` - Example implementation

## ✨ Benefits

1. ✅ **Consistent API responses** - Same format untuk semua endpoints
2. ✅ **User-friendly errors** - Pesan error yang mudah dipahami
3. ✅ **Automatic parsing** - PostgreSQL errors auto-parsed
4. ✅ **Type-safe** - Menggunakan Go generics
5. ✅ **Easy to maintain** - Central error handling
6. ✅ **Production ready** - Logging, panic recovery, validation
7. ✅ **TypeScript compatible** - Same structure dengan frontend

## 🎉 Ready to Use!

Error handling system sudah **100% complete** dan siap digunakan!

**Semua database errors (duplicate, foreign key, dll) sudah otomatis ter-handle!**

Happy coding! 🚀

