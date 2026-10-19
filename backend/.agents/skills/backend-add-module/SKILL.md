---
name: backend-add-module
description: >-
  Step-by-step procedure to scaffold and implement a new domain feature module in the Go Fiber backend (golang). Use when adding new API modules, entities, or business features.
---

# Scaffolding a New Go Backend Module

This skill guides you through implementing a new domain feature module in the **Go (Fiber)** backend following Clean Architecture.

---

## Architecture Pattern

Every domain feature lives in its own self-contained directory:
```text
internal/module/<feature>/
├── dto.go          # Request & Response structs with validation tags
├── repository.go   # Database queries using sqlx (parameterized $1, $2)
├── service.go      # Business logic interface and implementation
└── handler.go      # Fiber HTTP handlers with validation & response envelopes
```

---

## Step-by-Step Implementation Workflow

### Step 1: Create Directory
```bash
mkdir internal/module/<feature>
```

### Step 2: Define DTOs (`dto.go`)
Define all request payloads and response DTOs. Always include `json` and `validate` tags (`go-playground/validator`).

```go
package <feature>

import (
	"time"
)

// Request DTOs
type Create<Entity>Request struct {
	Name        string `json:"name" validate:"required,min=3,max=100"`
	Description string `json:"description" validate:"max=500"`
	Status      string `json:"status" validate:"required,oneof=active inactive pending"`
}

type Update<Entity>Request struct {
	Name        *string `json:"name" validate:"omitempty,min=3,max=100"`
	Description *string `json:"description" validate:"omitempty,max=500"`
	Status      *string `json:"status" validate:"omitempty,oneof=active inactive pending"`
}

type <Entity>Filter struct {
	Status string `query:"status"`
	Limit  int    `query:"limit" validate:"omitempty,min=1,max=100"`
	Offset int    `query:"offset" validate:"omitempty,min=0"`
}

// Response DTO
type <Entity>Response struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
```

### Step 3: Define Repository Layer (`repository.go`)
Use sqlx with parameterized queries. Never concatenate SQL strings.

> **ATURAN WAJIB DATABASE TRANSACTIONS (WAJIB PAKAI HELPER)**:
> **SEMUA** operasi write/mutasi data (`Create`, `Update`, `Delete`) **WAJIB** dieksekusi menggunakan transaction helper terpusat `database.TransactCtx` atau `r.db.WithTransactionCtx`.
> **DILARANG KERAS manual begin** (`BeginTxx`, manual `Commit`) agar penulisan kode seragam dan konsisten.
> Jika alur proses melibatkan pembacaan sebelum mutasi (read-then-write) atau relasi parent-child, semua query WAJIB dieksekusi di dalam callback helper yang sama.
> Operasi read-only murni (seperti `FindByID`, `FindAll`) **boleh langsung tanpa transaksi** via `r.db.GetDB()`.

```go
package <feature>

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"golang/database"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	Create(ctx context.Context, req Create<Entity>Request) (*<Entity>Response, error)
	FindByID(ctx context.Context, id int64) (*<Entity>Response, error)
	FindAll(ctx context.Context, filter <Entity>Filter) ([]<Entity>Response, int, error)
	Update(ctx context.Context, id int64, req Update<Entity>Request) (*<Entity>Response, error)
	Delete(ctx context.Context, id int64) error
}

type repository struct {
	db *database.PostgresDB
}

func NewRepository(db *database.PostgresDB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, req Create<Entity>Request) (*<Entity>Response, error) {
	return database.TransactCtx(ctx, r.db, func(tx *sqlx.Tx) (*<Entity>Response, error) {
		query := `
			INSERT INTO <table_name> (name, description, status, created_at, updated_at)
			VALUES ($1, $2, $3, NOW(), NOW())
			RETURNING id, name, description, status, created_at, updated_at
		`
		var res <Entity>Response
		err := tx.QueryRowxContext(ctx, query, req.Name, req.Description, req.Status).StructScan(&res)
		if err != nil {
			return nil, fmt.Errorf("failed to create: %w", err)
		}

		// Jika ada relasi child/batch inserts, eksekusi menggunakan tx di sini:
		// for _, child := range req.Children { ... tx.ExecContext(...) ... }

		return &res, nil
	})
}

func (r *repository) FindByID(ctx context.Context, id int64) (*<Entity>Response, error) {
	return database.TransactReadOnlyCtx(ctx, r.db, func(tx *sqlx.Tx) (*<Entity>Response, error) {
		query := `SELECT id, name, description, status, created_at, updated_at FROM <table_name> WHERE id = $1`
		var res <Entity>Response
		err := tx.GetContext(ctx, &res, query, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, nil
			}
			return nil, fmt.Errorf("failed to find by id: %w", err)
		}
		return &res, nil
	})
}
```

### Step 4: Define Service Layer (`service.go`)
Implement business logic and return typed errors from `pkg/errors`.

```go
package <feature>

import (
	"context"
	"golang/pkg/errors"
)

type Service interface {
	Create(ctx context.Context, req Create<Entity>Request) (*<Entity>Response, error)
	GetByID(ctx context.Context, id int64) (*<Entity>Response, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req Create<Entity>Request) (*<Entity>Response, error) {
	return s.repo.Create(ctx, req)
}

func (s *service) GetByID(ctx context.Context, id int64) (*<Entity>Response, error) {
	entity, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if entity == nil {
		return nil, errors.NotFound("<Entity> not found")
	}
	return entity, nil
}
```

### Step 5: Define Handler (`handler.go`)
Handle HTTP requests, validate incoming payloads via `pkg/validation`, and return responses using `pkg/response`.

```go
package <feature>

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"golang/pkg/logging"
	"golang/pkg/response"
	"golang/pkg/validation"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req Create<Entity>Request
	if err := validation.Default.BindAndValidate(c, &req); err != nil {
		return err
	}

	result, err := h.service.Create(c.UserContext(), req)
	if err != nil {
		return err
	}

	return response.CreatedResponse(c, result)
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid ID")
	}

	result, err := h.service.GetByID(c.UserContext(), id)
	if err != nil {
		return err
	}

	return response.SuccessResponse(c, result)
}
```

### Step 6: Register Route in `internal/module/router.go`
Wire repository, service, and handler, then mount the route group:

```go
// In internal/module/router.go:
import "<feature>"

// Inside SetupRoutes(app *fiber.App, db *sqlx.DB, ...):
<feature>Repo := <feature>.NewRepository(db)
<feature>Service := <feature>.NewService(<feature>Repo)
<feature>Handler := <feature>.NewHandler(<feature>Service)

api := app.Group("/api/v1/<feature_plural>")
api.Post("", <feature>Handler.Create)
api.Get("/:id", <feature>Handler.GetByID)
```

---

## Quality Gate Checklist

Before marking module creation complete:
- [ ] `make fmt` formats all Go files cleanly.
- [ ] `make check` passes without any linter issues.
- [ ] `go test ./internal/module/<feature>/...` runs cleanly.
- [ ] No `fmt.Println` statements exist (use `logging.LoggerFromFiber(c)`).
