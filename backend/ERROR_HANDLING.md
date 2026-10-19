# Error Handling & Response System

Complete error handling system dengan base response standar dan automatic PostgreSQL error parsing.

## 📦 Package Structure

```
pkg/
├── response/          # Standard API response structures
│   └── response.go
├── errors/            # Custom error types & database error parser
│   └── errors.go
├── constants/         # Error codes & PostgreSQL constraint mappings
│   └── database.go
├── middleware/        # Global error handler middleware
│   └── error_handler.go
└── handlers/          # Example handlers dengan error handling
    └── user_handler.go
```

## 🎯 Base Response Structure

```go
type BaseResponse[T any] struct {
Message   string    `json:"message"`
Success   bool      `json:"success"`
Data      *T        `json:"data"`
Code      string    `json:"code"`
Timestamp string    `json:"timestamp"`
}
```

### Success Response Example:

```json
{
  "message": "User created successfully",
  "success": true,
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "John Doe",
    "email": "john@example.com"
  },
  "code": "CREATED",
  "timestamp": "2026-01-17T10:30:00Z"
}
```

### Error Response Example:

```json
{
  "message": "The email already exists",
  "success": false,
  "data": null,
  "code": "DUPLICATE_ENTRY",
  "timestamp": "2026-01-17T10:30:00Z"
}
```

## 🚨 Error Codes

### Success Codes

- `SUCCESS` - Operation successful
- `CREATED` - Resource created

### Client Error Codes (4xx)

- `BAD_REQUEST` - Invalid request
- `VALIDATION_ERROR` - Validation failed
- `UNAUTHORIZED` - Authentication required
- `FORBIDDEN` - Insufficient permissions
- `NOT_FOUND` - Resource not found
- `CONFLICT` - Resource conflict
- `DUPLICATE_ENTRY` - Duplicate unique field
- `FOREIGN_KEY_VIOLATION` - Foreign key constraint violation
- `UNPROCESSABLE_ENTITY` - Semantic error

### Server Error Codes (5xx)

- `INTERNAL_SERVER_ERROR` - Server error
- `DATABASE_ERROR` - Database error
- `SERVICE_ERROR` - External service error

## 🗄️ PostgreSQL Error Handling

### Automatic Error Parsing

System otomatis mendeteksi dan mem-parse error PostgreSQL:

#### 1. **Unique Constraint Violation (23505)**

**Database Error:**

```sql
ERROR
:  duplicate key value violates unique constraint "users_email_key"
```

**Parsed Response:**

```json
{
  "message": "The email already exists",
  "success": false,
  "data": null,
  "code": "DUPLICATE_ENTRY",
  "timestamp": "2026-01-17T10:30:00Z"
}
```

#### 2. **Foreign Key Violation (23503)**

**Database Error (Insert/Update):**

```sql
ERROR
:  insert or
update on table "orders" violates foreign key constraint "orders_user_id_fkey"
```

**Parsed Response:**

```json
{
  "message": "The referenced user does not exist",
  "success": false,
  "data": null,
  "code": "FOREIGN_KEY_VIOLATION",
  "timestamp": "2026-01-17T10:30:00Z"
}
```

**Database Error (Delete):**

```sql
ERROR
:
update or
delete
on table "users" violates foreign key constraint "orders_user_id_fkey"
    DETAIL: Key (id)=(123) is still referenced
from table "orders"
```

**Parsed Response:**

```json
{
  "message": "Cannot delete this record because it is still referenced by order",
  "success": false,
  "data": null,
  "code": "FOREIGN_KEY_VIOLATION",
  "timestamp": "2026-01-17T10:30:00Z"
}
```

#### 3. **Not Null Violation (23502)**

```json
{
  "message": "Field 'email' is required",
  "success": false,
  "data": null,
  "code": "VALIDATION_ERROR",
  "timestamp": "2026-01-17T10:30:00Z"
}
```

#### 4. **Check Constraint Violation (23514)**

```json
{
  "message": "Data validation failed: age_check",
  "success": false,
  "data": null,
  "code": "VALIDATION_ERROR",
  "timestamp": "2026-01-17T10:30:00Z"
}
```

## 🔧 Configuration

### Add Constraint Mappings

Edit `pkg/constants/database.go`:

```go
var UniqueConstraintFieldMap = map[string]string{
"skills_name_key":       "name",
"users_email_key":       "email",
"users_username_key":    "username",
"products_sku_key":      "sku",
// Add your own mappings here
}

var ForeignKeyConstraintMap = map[string]string{
"users_role_id_fkey":           "role",
"orders_user_id_fkey":          "user",
"order_items_product_id_fkey":  "product",
// Add your own mappings here
}
```

## 💻 Usage Examples

### 1. Setup Fiber App with Error Handler

```go
package main

import (
	"github.com/gofiber/fiber/v2"
	"golang/pkg/middleware"
)

func main() {
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler(), // ← Global error handler
	})

	app.Use(middleware.RecoverMiddleware()) // ← Panic recovery

	// Your routes...

	app.Listen(":3000")
}
```

### 2. Handler dengan Success Response

```go
func GetUsers(db *database.PostgresDB) fiber.Handler {
return func (c *fiber.Ctx) error {
var users []User

err := db.DB.Select(&users, "SELECT * FROM users")
if err != nil {
return customErrors.DatabaseError("Failed to fetch users", err)
}

return response.SuccessResponse(c, users, "Users retrieved successfully")
}
}
```

### 3. Handler dengan Validation

```go
type CreateUserRequest struct {
Name  string `json:"name" validate:"required,min=3"`
Email string `json:"email" validate:"required,email"`
}

func CreateUser(db *database.PostgresDB) fiber.Handler {
return func (c *fiber.Ctx) error {
var req CreateUserRequest

if err := c.BodyParser(&req); err != nil {
return customErrors.BadRequest("Invalid request body")
}

// Validation errors automatically handled
if err := validate.Struct(req); err != nil {
return err // Middleware will format this nicely
}

// Insert to database
_, err := db.DB.Exec("INSERT INTO users (name, email) VALUES ($1, $2)",
req.Name, req.Email)

if err != nil {
// Automatically parses duplicate email, foreign key errors, etc.
return customErrors.ParseDatabaseError(err)
}

return response.CreatedResponse(c, user, "User created successfully")
}
}
```

### 4. Custom Error Responses

```go
// 400 Bad Request
return customErrors.BadRequest("Invalid input data")

// 401 Unauthorized
return customErrors.Unauthorized("Please login to continue")

// 403 Forbidden
return customErrors.Forbidden("You don't have permission to access this resource")

// 404 Not Found
return customErrors.NotFound("User not found")

// 409 Conflict
return customErrors.Conflict("Resource already exists")

// 500 Internal Error
return customErrors.InternalError("Something went wrong", err)

// Database Error
return customErrors.DatabaseError("Failed to save data", err)

// Validation Error
return customErrors.ValidationError("Invalid email format")
```

### 5. Using Response Helpers

```go
// Success with data
response.SuccessResponse(c, data, "Operation successful")

// Created (201)
response.CreatedResponse(c, data, "Resource created")

// No content (204)
response.NoContentResponse(c)

// Validation error
response.ValidationErrorResponse(c, validationErrors)

// Error response
response.ErrorResponse(c, statusCode, message, code)

// Paginated response
response.PaginatedSuccessResponse(c, paginatedData, "Data retrieved")
```

## 🧪 Testing Error Responses

### Test Duplicate Email:

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

### Test Validation Error:

```bash
curl -X POST http://localhost:3000/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Jo","email":"invalid-email"}'
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

### Test Foreign Key Violation:

```bash
curl -X DELETE http://localhost:3000/api/v1/users/123
```

**Response (if user has orders):**

```json
{
  "message": "Cannot delete this record because it is still referenced by order",
  "success": false,
  "data": null,
  "code": "FOREIGN_KEY_VIOLATION",
  "timestamp": "2026-01-17T10:30:00Z"
}
```

## 📝 Migration Example with Constraints

```sql
-- migrations/000003_create_products.up.sql
CREATE TABLE products
(
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255)        NOT NULL,
    sku         VARCHAR(100) UNIQUE NOT NULL, -- Will trigger products_sku_key
    category_id UUID                NOT NULL,
    price       DECIMAL(10, 2) CHECK (price >= 0),
    created_at  TIMESTAMP        DEFAULT CURRENT_TIMESTAMP,

    -- Foreign key with named constraint
    CONSTRAINT products_category_id_fkey
        FOREIGN KEY (category_id)
            REFERENCES categories (id)
            ON DELETE RESTRICT
);

CREATE INDEX idx_products_sku ON products (sku);
```

Then add to `constants/database.go`:

```go
var UniqueConstraintFieldMap = map[string]string{
"products_sku_key": "sku",
}

var ForeignKeyConstraintMap = map[string]string{
"products_category_id_fkey": "category",
}
```

## 🎯 Best Practices

1. **Always use ParseDatabaseError** untuk database operations
2. **Return errors, don't handle in handler** - let middleware handle them
3. **Add constraint mappings** di constants untuk user-friendly messages
4. **Use validation tags** untuk input validation
5. **Log errors** untuk debugging (middleware sudah log otomatis)
6. **Use proper HTTP status codes** dengan error types
7. **Consistent error messages** menggunakan constants

## 🔍 Error Flow

```
Handler Error
    ↓
Middleware ErrorHandler
    ↓
Check Error Type:
  - AppError? → Return formatted response
  - Fiber Error? → Return formatted response
  - Validation Error? → Format and return
  - sql.ErrNoRows? → Return 404
  - Database Error? → Parse PostgreSQL error → Return formatted
  - Other? → Return 500 Internal Error
```

## ✅ Complete Setup Checklist

- [x] Base response structure created
- [x] Error types & handlers created
- [x] PostgreSQL error parser created
- [x] Constraint mappings configured
- [x] Global error handler middleware created
- [x] Validation error formatter created
- [x] Example handlers created
- [x] Server setup with error handler
- [x] Documentation complete

## 🚀 Next Steps

1. Update `pkg/constants/database.go` dengan constraint mappings Anda
2. Buat handlers menggunakan pattern yang sudah disediakan
3. Test error responses untuk setiap endpoint
4. Add more custom error types jika diperlukan

Happy coding! 🎉

