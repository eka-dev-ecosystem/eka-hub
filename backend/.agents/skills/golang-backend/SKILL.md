---
name: golang-backend
description: >-
  Clean architecture, fiber routing, validation with pkg/validation, structured slog logging, sqlx raw queries, and quality gates (make check) for the Go backend template. Use when creating or modifying Go modules, handlers, services, repositories, or migrations.
---

# Go Fiber Backend Skill Guide

This skill provides standard procedures, rules, and best practices for developing and maintaining the **Go (Fiber)** backend template.

---

## 1. Stack & Architecture

- **Language & Framework**: Go 1.25.0, Fiber v2.
- **Pattern**: Clean Architecture (Handler → Service → Repository).
- **Project Structure**:
  ```text
  ├── cmd/
  │   └── server/             # Entry point binary (main.go)
  ├── internal/
  │   ├── app/                # Bootstrapping & DI container
  │   └── module/             # Feature modules (self-contained)
  │       ├── router.go       # Route registry
  │       └── <feature>/
  │           ├── handler.go  # Fiber route handlers
  │           ├── service.go  # Business logic interface & implementation
  │           ├── repository.go # Database queries (sqlx)
  │           └── dto.go      # Request/Response structs with validator tags
  ├── pkg/                    # Reusable libraries
  │   ├── auth/               # JWT middleware
  │   ├── cache/              # Redis wrapper (supports REDIS_USERNAME)
  │   ├── errors/             # AppError definitions
  │   ├── logging/            # slog structured JSON logger
  │   ├── queue/              # RabbitMQ client
  │   ├── realtime/           # WebSocket & SSE hub
  │   ├── response/           # BaseResponse envelope helpers
  │   └── validation/         # Unified request binding & validation
  ├── database/               # sqlx connection & migration runner
  └── migrations/             # golang-migrate UP & DOWN SQL files
  ```

---

## 2. Core Rules & Conventions

### A. Input Binding & Validation
* Always use `pkg/validation` in handlers:
  * Body (JSON, Form, Multipart): `validation.Default.BindAndValidate(c, &dto)`
  * Query parameters: `validation.Default.BindQueryAndValidate(c, &dto)`
* Use `json` and `validate` struct tags (`go-playground/validator`).
```go
type CreateUserRequest struct {
    Name   string `json:"name" validate:"required,min=3,max=100"`
    Email  string `json:"email" validate:"required,email"`
}
```

### B. Standardized Responses & Errors
* **Success**: Return using `pkg/response`:
  * `response.SuccessResponse(c, data)`
  * `response.CreatedResponse(c, data)`
  * `response.PaginatedSuccessResponse(c, items, pagination)`
* **Errors**: Return error types from `pkg/errors`:
  * `errors.BadRequest("invalid input", err)`
  * `errors.NotFound("user not found")`
  * `errors.Unauthorized("invalid credentials")`

### C. Logging
* Use `pkg/logging` based on `log/slog`.
* In handlers, use `logging.LoggerFromFiber(c)`:
  ```go
  log := logging.LoggerFromFiber(c)
  log.Info("processing order", "order_id", id)
  ```
* **Never use `fmt.Println`** ad-hoc.

### D. Database & Migrations
* Repositories interact via `database` (`*sqlx.DB`).
* Always parameterize queries (`$1`, `$2`) to avoid SQL injection.
* **100% WAJIB Database Transactions untuk SEMUA Operasi**:
  - **Mutasi (`INSERT`, `UPDATE`, `DELETE`)**: **WAJIB** menggunakan Regular Transaction helper (`database.TransactCtx` / `db.WithTransactionCtx` atau `database.TransactSqlxCtx` / `database.WithTransactionSqlxCtx`).
  - **Pembacaan (`SELECT`, query, get, list)**: **WAJIB** menggunakan Read-Only Transaction helper (`database.TransactReadOnlyCtx` / `db.WithReadOnlyTransactionCtx` atau `database.TransactReadOnlySqlxCtx` / `database.WithReadOnlyTransactionSqlxCtx`).
  - **Read-then-Write**: Jika ada alur baca lalu tulis, bungkus seluruh alur dalam satu Regular Transaction.
  - **DILARANG keras manual begin**: Dilarang menulis `BeginTxx` dan `Commit()` manual di kode bisnis agar pola kode seragam.
  - **DILARANG query tanpa transaksi**: Seluruh interaksi DB wajib lewat transaction helper.
* For migrations, always create a matching pair:
  `migrations/<timestamp>_<name>.up.sql` and `migrations/<timestamp>_<name>.down.sql`.

---

## 3. Quality Gate & Verification

Before reporting completion on any Go task:
```bash
# Quality gate: fmt-check + vet + lint (MUST BE GREEN)
make check

# Format code
make fmt

# Run unit tests
make test-unit

# Run server
make run
```
