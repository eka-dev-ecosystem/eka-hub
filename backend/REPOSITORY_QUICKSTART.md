# Repository Pattern - Quick Start Guide

## 🎯 What You Get

The codebase has been refactored to use the **Repository Pattern** for clean architecture:

```
Handler → Service → Repository → Database
```

## 📁 File Structure

```
pkg/
├── models/user.go           # Data models & DTOs
├── repository/              # Data access layer
│   └── user_repository.go
├── services/                # Business logic layer
│   └── user_service.go
└── handlers/                # HTTP handlers
    └── user_handler.go
```

## 🚀 API Endpoints

### User Management (CRUD)

| Method | Endpoint                  | Description      |
|--------|---------------------------|------------------|
| GET    | `/api/v1/users`           | Get all users    |
| GET    | `/api/v1/users/:id`       | Get user by ID   |
| POST   | `/api/v1/users`           | Create user      |
| PUT    | `/api/v1/users/:id`       | Update user      |
| DELETE | `/api/v1/users/:id`       | Soft delete user |
| DELETE | `/api/v1/users/admin/:id` | Permanent delete |

### Examples

**Create User:**

```bash
curl -X POST http://localhost:3000/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "phone": "1234567890"
  }'
```

**Response:**

```json
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

## 🏗️ Architecture Layers

### 1. Model Layer (`pkg/models/user.go`)

Defines data structures:

```go
type User struct {
ID        string    `json:"id" db:"id"`
Name      string    `json:"name" db:"name"`
Email     string    `json:"email" db:"email"`
// ...
}

type CreateUserRequest struct {
Name  string `json:"name" validate:"required,min=3,max=100"`
Email string `json:"email" validate:"required,email"`
// ...
}
```

### 2. Repository Layer (`pkg/repository/user_repository.go`)

Database operations:

```go
type UserRepository interface {
GetAll() ([]models.User, error)
GetByID(id string) (*models.User, error)
Create(user *models.User) error
Update(id string, updates map[string]interface{}) (*models.User, error)
Delete(id string) error
}
```

### 3. Service Layer (`pkg/services/user_service.go`)

Business logic:

```go
type UserService interface {
GetAllUsers() ([]models.UserResponse, error)
GetUserByID(id string) (*models.UserResponse, error)
CreateUser(req *models.CreateUserRequest) (*models.UserResponse, error)
UpdateUser(id string, req *models.UpdateUserRequest) (*models.UserResponse, error)
DeleteUser(id string) error
}
```

### 4. Handler Layer (`pkg/handlers/user_handler.go`)

HTTP request/response:

```go
type UserHandler struct {
service services.UserService
}

func (h *UserHandler) CreateUser() fiber.Handler {
// Parse, validate, call service, return response
}
```

## 🔧 How to Add New Features

### Example: Add "Posts" feature

**Step 1: Create Model** (`pkg/models/post.go`)

```go
package models

type Post struct {
	ID        string    `json:"id" db:"id"`
	Title     string    `json:"title" db:"title"`
	Content   string    `json:"content" db:"content"`
	UserID    string    `json:"user_id" db:"user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type CreatePostRequest struct {
	Title   string `json:"title" validate:"required,min=3,max=200"`
	Content string `json:"content" validate:"required"`
}
```

**Step 2: Create Repository** (`pkg/repository/post_repository.go`)

```go
package repository

type PostRepository interface {
	GetAll() ([]models.Post, error)
	GetByID(id string) (*models.Post, error)
	Create(post *models.Post) error
}

type postRepository struct {
	db *sqlx.DB
}

func NewPostRepository(db *sqlx.DB) PostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) GetAll() ([]models.Post, error) {
	var posts []models.Post
	query := "SELECT * FROM posts ORDER BY created_at DESC"
	err := r.db.Select(&posts, query)
	return posts, err
}

func (r *postRepository) Create(post *models.Post) error {
	query := `INSERT INTO posts (id, title, content, user_id, created_at) 
              VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(query, post.ID, post.Title, post.Content,
		post.UserID, post.CreatedAt)
	return err
}
```

**Step 3: Create Service** (`pkg/services/post_service.go`)

```go
package services

type PostService interface {
	GetAllPosts() ([]models.Post, error)
	CreatePost(req *models.CreatePostRequest, userID string) (*models.Post, error)
}

type postService struct {
	repo repository.PostRepository
}

func NewPostService(repo repository.PostRepository) PostService {
	return &postService{repo: repo}
}

func (s *postService) CreatePost(req *models.CreatePostRequest, userID string) (*models.Post, error) {
	post := &models.Post{
		ID:        uuid.New().String(),
		Title:     req.Title,
		Content:   req.Content,
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	err := s.repo.Create(post)
	if err != nil {
		return nil, customErrors.DatabaseError("Failed to create post", err)
	}

	return post, nil
}
```

**Step 4: Create Handler** (`pkg/handlers/post_handler.go`)

```go
package handlers

type PostHandler struct {
	service services.PostService
}

func NewPostHandler(service services.PostService) *PostHandler {
	return &PostHandler{service: service}
}

func (h *PostHandler) CreatePost() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.CreatePostRequest
		if err := c.BodyParser(&req); err != nil {
			return customErrors.BadRequest("Invalid request body")
		}

		if err := validate.Struct(req); err != nil {
			return err
		}

		userID, _ := authMiddleware.GetUserID(c)
		post, err := h.service.CreatePost(&req, userID)
		if err != nil {
			return err
		}

		return response.CreatedResponse(c, post, "Post created successfully")
	}
}
```

**Step 5: Wire Up in main.go**

```go
func setupPostRoutes(api fiber.Router, db *database.PostgresDB) {
// Wire up layers
postRepo := repository.NewPostRepository(db.GetDB())
postService := services.NewPostService(postRepo)
postHandler := handlers.NewPostHandler(postService)

// Setup routes
posts := api.Group("/posts")
posts.Get("/", postHandler.GetPosts())
posts.Post("/", postHandler.CreatePost())
}

// In main()
func main() {
// ... existing code ...

setupUserRoutes(api, db)
setupPostRoutes(api, db) // Add this line
}
```

## ✅ Benefits

1. **Testability**: Mock repositories for unit testing
2. **Maintainability**: Each layer has single responsibility
3. **Flexibility**: Easy to swap database implementations
4. **Reusability**: Services can be reused across handlers
5. **Clean Code**: Separation of concerns

## 🧪 Testing Example

```go
// Mock repository for testing
type MockUserRepository struct{}

func (m *MockUserRepository) GetByID(id string) (*models.User, error) {
return &models.User{
ID:    id,
Name:  "Test User",
Email: "test@example.com",
}, nil
}

// Test service without database
func TestUserService_GetUserByID(t *testing.T) {
mockRepo := &MockUserRepository{}
service := services.NewUserService(mockRepo)

user, err := service.GetUserByID("123")

assert.NoError(t, err)
assert.Equal(t, "Test User", user.Name)
}
```

## 📝 Error Handling

All errors follow the standard response format:

```json
{
  "message": "User not found",
  "success": false,
  "data": null,
  "code": "NOT_FOUND",
  "timestamp": "2026-01-18T10:00:00Z"
}
```

Error types handled:

- `BAD_REQUEST` - Invalid input (400)
- `NOT_FOUND` - Resource not found (404)
- `CONFLICT` - Duplicate/constraint violation (409)
- `DATABASE_ERROR` - Database errors (500)
- `INTERNAL_SERVER_ERROR` - Unexpected errors (500)

## 🗄️ Database Constraints

PostgreSQL unique constraints are automatically handled:

```go
// Service layer automatically detects duplicate emails
exists, _ := s.repo.ExistsByEmail(req.Email)
if exists {
return customErrors.Conflict("Email already exists")
}

// Foreign key violations are caught by error handler
return customErrors.ParseDatabaseError(err)
```

## 🔐 Authentication Integration

User routes can be protected with JWT middleware:

```go
func setupUserRoutes(api fiber.Router, db *database.PostgresDB, jwtService *auth.JWTService) {
// ... setup layers ...

users := api.Group("/users")

// Protect routes
users.Use(authMiddleware.AuthMiddleware(jwtService))

users.Get("/", userHandler.GetUsers())
users.Post("/", userHandler.CreateUser())
}
```

## 📚 Further Reading

See `REPOSITORY_PATTERN.md` for detailed documentation.

## 🎉 Summary

You now have a clean, maintainable architecture with:

- ✅ Repository pattern implementation
- ✅ Separation of concerns
- ✅ Easy testing
- ✅ Standard error handling
- ✅ Clean API responses
- ✅ JWT authentication support

Start building your features following this pattern!
