# Integration Testing Guide

This project uses `testcontainers-go` to provide a robust, isolated environment for integration testing. Real Docker containers for PostgreSQL are dynamically spun up, tested against, and torn down automatically.

## Prerequisites
- Docker must be installed and running on your machine.
- Go 1.21+

## Running the Tests

To run the integration tests, you need to use the `integration` build tag. This prevents the slower integration tests from running during normal unit test execution.

```bash
# Run integration tests
go test -tags=integration ./...

# Run with verbose output
go test -v -tags=integration ./...
```

## How It Works
1. **Global Setup**: The global setup resides in `internal/app/main_test.go`. The `TestMain` function spins up a PostgreSQL Testcontainer before executing any tests.
2. **Database Migrations**: The `TestMain` function applies all database migrations in the `migrations/` folder directly to the test container.
3. **Environment Injection**: The connection string for the test container is injected into the application's configuration, so the standard `app.Start()` function connects to the isolated database.

## Writing New Integration Tests

When writing new integration tests:
1. Ensure the file has the `//go:build integration` build tag at the very top.
2. The tests should be placed alongside standard tests, typically ending in `_test.go`.
3. You can test your handlers using Fiber's `app.Test(req)` method. Since `TestMain` handles the database, your test can make actual queries to the test database.

Example:
```go
//go:build integration

package app_test

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestMyEndpoint(t *testing.T) {
    req := httptest.NewRequest("GET", "/api/v1/my-endpoint", nil)
    resp, err := appInstance.Test(req, -1)
    
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}
```
