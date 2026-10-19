---
name: write-unit-test
description: >-
  Standard pattern for writing table-driven unit tests for Go services and logic using mock repositories and subtests. Use when writing or updating Go test suites.
---

# Writing Unit Tests in Go (Table-Driven Pattern)

This skill guides you through writing robust, idiomatic **table-driven unit tests** for service layers in Go.

---

## Pattern Overview

Unit tests in `internal/module/<feature>/service_test.go`:
- Test business logic in isolation without touching a real database.
- Use a lightweight mock repository implementing the feature's `Repository` interface.
- Follow Go's idiomatic table-driven test pattern (`[]struct`).

---

## Step-by-Step Implementation

### Step 1: Create Mock Repository
Define a mock struct implementing the `Repository` interface:

```go
package <feature>

import (
	"context"
	"testing"
)

type mockRepository struct {
	createFunc   func(ctx context.Context, req Create<Entity>Request) (*<Entity>Response, error)
	findByIDFunc func(ctx context.Context, id int64) (*<Entity>Response, error)
}

func (m *mockRepository) Create(ctx context.Context, req Create<Entity>Request) (*<Entity>Response, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, req)
	}
	return nil, nil
}

func (m *mockRepository) FindByID(ctx context.Context, id int64) (*<Entity>Response, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, nil
}
```

### Step 2: Write Table-Driven Service Tests
Use subtests (`t.Run`) and cover both happy path and edge/error cases:

```go
func TestService_GetByID(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		id          int64
		mockRepo    Repository
		expectErr   bool
		expectedErr string
	}{
		{
			name: "Success - Entity found",
			id:   1,
			mockRepo: &mockRepository{
				findByIDFunc: func(ctx context.Context, id int64) (*<Entity>Response, error) {
					return &<Entity>Response{ID: 1, Name: "Test Entity", Status: "active"}, nil
				},
			},
			expectErr: false,
		},
		{
			name: "Error - Entity not found",
			id:   999,
			mockRepo: &mockRepository{
				findByIDFunc: func(ctx context.Context, id int64) (*<Entity>Response, error) {
					return nil, nil // Nil indicates not found
				},
			},
			expectErr:   true,
			expectedErr: "<Entity> not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(tt.mockRepo)
			res, err := svc.GetByID(ctx, tt.id)

			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.expectedErr != "" && err.Error() != tt.expectedErr {
					t.Errorf("expected error message %q, got %q", tt.expectedErr, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if res == nil || res.ID != tt.id {
					t.Errorf("expected entity with ID %d, got %v", tt.id, res)
				}
			}
		})
	}
}
```

---

## Running & Verifying Tests

```bash
# Run unit tests for specific package
go test -v ./internal/module/<feature>/...

# Run all unit tests
make test-unit
```
