# Repository Pattern Implementation

This project implements the **Repository Pattern** for clean architecture and separation of concerns.

## Architecture Layers

```
┌─────────────────────────────────────────────────────────────┐
│                    HTTP Request (Client)                     │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Handler Layer (HTTP/API)                   │
│  - Parse HTTP requests                                       │
│  - Validate input                                            │
│  - Format HTTP responses                                     │
│  Location: pkg/handlers/                                     │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Service Layer (Business)                   │
│  - Business logic                                            │
│  - Validation & authorization                                │
│  - Orchestration between repositories                        │
│  Location: pkg/services/                                     │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                Repository Layer (Data Access)                │
│  - Database operations                                       │
│  - SQL queries                                               │
│  - Data mapping                                              │
│  Location: pkg/repository/                                   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      Database (PostgreSQL)                   │
└─────────────────────────────────────────────────────────────┘
```

## Project Structure

```
golang/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point & route configuration
├── pkg/
│   ├── models/                  # Data models & DTOs
│   │   └── user.go
│   ├── repository/              # Data access layer (Database operations)
│   │   └── user_repository.go
│   ├── services/                # Business logic layer
│   │   └── user_service.go
│   ├── handlers/                # HTTP handler layer (Controllers)
│   │   ├── user_handler.go
│   │   └── auth_handler.go
│   ├── middleware/              # HTTP middlewares
│   ├── errors/                  # Custom error handling
│   └── response/                # Standard response formats
└── database/
    └── postgres.go              # Database connection
```

## Layer Responsibilities

### 1. Models Layer (`pkg/models/`)

**Purpose**: Define data structures

```go
type User struct {
ID        string    `json:"id" db:"id"`
Name      string    `json:"name" db:"name"`
Email     string    `json:"email" db:"email"`
Phone     string    `json:"phone" db:"phone"`
Password  string    `json:"-" db:"password"`
IsActive  bool      `json:"is_active" db:"is_active"`
CreatedAt time.Time `json:"created_at" db:"created_at"`
UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
```

### 2. Repository Layer (`pkg/repository/`)

**Purpose**: Database operations only

```go
type UserRepository interface {
GetAll() ([]models.User, error)
GetByID(id string) (*models.User, error)
GetByEmail(email string) (*models.User, error)
Create(user *models.User) error
Update(id string, updates map[string]interface{}) (*models.User, error)
Delete(id string) error
}
```

**Responsibilities**:

- Execute SQL queries
- Map database results to models
- Handle database-specific errors
- NO business logic

### 3. Service Layer (`pkg/services/`)

**Purpose**: Business logic & orchestration

```go
type UserService interface {
GetAllUsers() ([]models.UserResponse, error)
GetUserByID(id string) (*models.UserResponse, error)
CreateUser(req *models.CreateUserRequest) (*models.UserResponse, error)
UpdateUser(id string, req *models.UpdateUserRequest) (*models.UserResponse, error)
DeleteUser(id string) error
}
```

**Responsibilities**:

- Implement business rules
- Validate data
- Check permissions
- Transform data (e.g., hide sensitive fields)
- Orchestrate multiple repository calls
- Handle application-level errors

### 4. Handler Layer (`pkg/handlers/`)

**Purpose**: HTTP request/response handling

```go
type UserHandler struct {
service services.UserService
}

func (h *UserHandler) CreateUser() fiber.Handler {
return func (c *fiber.Ctx) error {
// 1. Parse request
var req models.CreateUserRequest
if err := c.BodyParser(&req); err != nil {
return customErrors.BadRequest("Invalid request body")
}

// 2. Validate
if err := validate.Struct(req); err != nil {
return err
}

// 3. Call service
user, err := h.service.CreateUser(&req)
if err != nil {
return err
}

// 4. Return response
return response.CreatedResponse(c, user, "User created successfully")
}
}
```

**Responsibilities**:

- Parse HTTP requests
- Validate input format
- Call service methods
- Format HTTP responses
- Handle HTTP status codes

## Request Flow Example

### Creating a User

**1. Client Request**

```bash
POST /api/v1/users HTTP/1.1
Content-Type: application/json

{
  "name": "John Doe",
  "email": "john@example.com",
  "phone": "1234567890"
}
```

**2. Handler Layer** (`pkg/handlers/user_handler.go`)

```go
func (h *UserHandler) CreateUser() fiber.Handler {
// Parse JSON body
var req models.CreateUserRequest
c.BodyParser(&req)

// Validate input
validate.Struct(req)

// Call service
user, err := h.service.CreateUser(&req)

// Return JSON response
return response.CreatedResponse(c, user, "User created successfully")
}
```

**3. Service Layer** (`pkg/services/user_service.go`)

```go
func (s *userService) CreateUser(req *models.CreateUserRequest) (*models.UserResponse, error) {
// Business logic: Check if email exists
exists, err := s.repo.ExistsByEmail(req.Email)
if exists {
return nil, customErrors.Conflict("Email already exists")
}

// Business logic: Generate UUID, set defaults
user := &models.User{
ID:        uuid.New().String(),
Name:      req.Name,
Email:     req.Email,
Phone:     req.Phone,
IsActive:  true,
CreatedAt: time.Now(),
UpdatedAt: time.Now(),
}

// Call repository
err = s.repo.Create(user)

// Transform to response (hide sensitive data)
return &user.ToResponse(), nil
}
```

**4. Repository Layer** (`pkg/repository/user_repository.go`)

```go
func (r *userRepository) Create(user *models.User) error {
// Execute SQL query
query := `
        INSERT INTO users (id, name, email, phone, password, is_active, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `
_, err := r.db.Exec(query, user.ID, user.Name, user.Email, user.Phone,
user.Password, user.IsActive, user.CreatedAt, user.UpdatedAt)
return err
}
```

**5. Database**

```sql
INSERT INTO users (id, name, email, phone, password, is_active, created_at, updated_at)
VALUES ('uuid-here', 'John Doe', 'john@example.com', '1234567890', '', true, NOW(), NOW());
```

**6. Response**

```json
HTTP/1.1 201 Created
Content-Type: application/json

{
"message": "User created successfully",
"success": true,
"data": {
"id": "uuid-here",
"name": "John Doe",
"email": "john@example.com",
"phone": "1234567890",
"is_active": true,
"created_at": "2026-01-18T10:00:00Z",
"updated_at": "2026-01-18T10:00:00Z"
},
"code": "SUCCESS",
"timestamp": "2026-01-18T10:00:00Z"
}
```

## Benefits

### 1. Separation of Concerns

- Each layer has a single responsibility
- Changes in one layer don't affect others

### 2. Testability

```go
// Mock repository for testing service layer
type MockUserRepository struct{}

func (m *MockUserRepository) GetByID(id string) (*models.User, error) {
return &models.User{ID: id, Name: "Test User"}, nil
}

// Test service without database
func TestGetUserByID(t *testing.T) {
mockRepo := &MockUserRepository{}
service := services.NewUserService(mockRepo)

user, err := service.GetUserByID("123")
assert.NoError(t, err)
assert.Equal(t, "Test User", user.Name)
}
```

### 3. Maintainability

- Easy to find and modify code
- Clear structure for new developers

### 4. Reusability

- Services can be used by multiple handlers
- Repositories can be used by multiple services

### 5. Flexibility

- Easy to swap implementations (e.g., PostgreSQL → MongoDB)
- Easy to add caching, logging, etc.

## Error Handling

All errors are handled through custom error types:

```go
// In service layer
if !exists {
return nil, customErrors.NotFound("User not found")
}

if emailExists {
return nil, customErrors.Conflict("Email already exists")
}

// In repository layer
if err == sql.ErrNoRows {
return nil, nil // Return nil for not found
}
return customErrors.DatabaseError("Failed to fetch user", err)
```

The error handler middleware catches these errors and formats them:

```json
{
  "message": "Email already exists",
  "success": false,
  "data": null,
  "code": "CONFLICT",
  "timestamp": "2026-01-18T10:00:00Z"
}
```

## API Endpoints

### User Management (Repository Pattern)

| Method | Endpoint                  | Description      |
|--------|---------------------------|------------------|
| GET    | `/api/v1/users`           | Get all users    |
| GET    | `/api/v1/users/:id`       | Get user by ID   |
| POST   | `/api/v1/users`           | Create new user  |
| PUT    | `/api/v1/users/:id`       | Update user      |
| DELETE | `/api/v1/users/:id`       | Soft delete user |
| DELETE | `/api/v1/users/admin/:id` | Permanent delete |

### Authentication (JWT)

| Method | Endpoint                | Description          |
|--------|-------------------------|----------------------|
| POST   | `/api/v1/auth/register` | Register new account |
| POST   | `/api/v1/auth/login`    | Login                |
| POST   | `/api/v1/auth/refresh`  | Refresh access token |
| POST   | `/api/v1/auth/logout`   | Logout (protected)   |
| GET    | `/api/v1/auth/profile`  | Get profile          |
| PUT    | `/api/v1/auth/profile`  | Update profile       |

## Setup in main.go

```go
func setupUserRoutes(api fiber.Router, db *database.PostgresDB) {
// Layer 1: Repository (Data Access)
userRepo := repository.NewUserRepository(db.GetDB())

// Layer 2: Service (Business Logic)
userService := services.NewUserService(userRepo)

// Layer 3: Handler (HTTP Layer)
userHandler := handlers.NewUserHandler(userService)

// Layer 4: Routes
users := api.Group("/users")
users.Get("/", userHandler.GetUsers())
users.Post("/", userHandler.CreateUser())
// ... more routes
}
```

## Adding New Features

### Example: Add "Skills" feature

**1. Create Model** (`pkg/models/skill.go`)

```go
type Skill struct {
ID     string `json:"id" db:"id"`
Name   string `json:"name" db:"name"`
UserID string `json:"user_id" db:"user_id"`
}
```

**2. Create Repository** (`pkg/repository/skill_repository.go`)

```go
type SkillRepository interface {
GetByUserID(userID string) ([]models.Skill, error)
Create(skill *models.Skill) error
}
```

**3. Create Service** (`pkg/services/skill_service.go`)

```go
type SkillService interface {
GetUserSkills(userID string) ([]models.Skill, error)
AddSkill(req *models.CreateSkillRequest) (*models.Skill, error)
}
```

**4. Create Handler** (`pkg/handlers/skill_handler.go`)

```go
type SkillHandler struct {
service services.SkillService
}
```

**5. Add Routes** (`cmd/server/main.go`)

```go
func setupSkillRoutes(api fiber.Router, db *database.PostgresDB) {
skillRepo := repository.NewSkillRepository(db.GetDB())
skillService := services.NewSkillService(skillRepo)
skillHandler := handlers.NewSkillHandler(skillService)

skills := api.Group("/skills")
skills.Get("/user/:userId", skillHandler.GetUserSkills())
skills.Post("/", skillHandler.AddSkill())
}
```

## Best Practices

1. **Keep layers independent**: Don't skip layers (Handler → Repository)
2. **Use interfaces**: Makes testing easier
3. **Handle errors at the right layer**:
    - Repository: Database errors
    - Service: Business logic errors
    - Handler: HTTP errors
4. **Use DTOs**: Separate request/response models from database models
5. **Validate early**: Validate in handler, business rules in service
6. **Return appropriate types**: Repository returns models, Service returns responses

## Comparison with Direct Database Access

### ❌ Without Repository Pattern

```go
// Handler directly accessing database
func CreateUser(db *sqlx.DB) fiber.Handler {
return func (c *fiber.Ctx) error {
var req CreateUserRequest
c.BodyParser(&req)

// Check email exists
var exists bool
db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", req.Email)

// Insert user
db.Exec("INSERT INTO users ...", ...)

return c.JSON(user)
}
}
```

**Problems**: Tight coupling, hard to test, duplicated code, mixed concerns

### ✅ With Repository Pattern

```go
// Clean separation of concerns
Handler → Service → Repository → Database

// Easy to test each layer independently
// Business logic centralized in service
// Database logic centralized in repository
```

## Conclusion

The Repository Pattern provides:

- ✅ Clean code organization
- ✅ Easy testing
- ✅ Better maintainability
- ✅ Flexible architecture
- ✅ Team collaboration friendly

Start with this pattern from the beginning to avoid refactoring later!
