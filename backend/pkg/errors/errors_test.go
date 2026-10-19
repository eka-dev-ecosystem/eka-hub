package errors

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"testing"

	"golang/pkg/constants"

	"github.com/lib/pq"
)

func TestConstructorsSetCodesAndStatus(t *testing.T) {
	tests := []struct {
		name   string
		build  func() *AppError
		code   string
		status int
	}{
		{"BadRequest", func() *AppError { return BadRequest("bad") }, constants.CodeBadRequest, http.StatusBadRequest},
		{"Unauthorized", func() *AppError { return Unauthorized("no") }, constants.CodeUnauthorized, http.StatusUnauthorized},
		{"Forbidden", func() *AppError { return Forbidden("nope") }, constants.CodeForbidden, http.StatusForbidden},
		{"NotFound", func() *AppError { return NotFound("gone") }, constants.CodeNotFound, http.StatusNotFound},
		{"Conflict", func() *AppError { return Conflict("taken") }, constants.CodeConflict, http.StatusConflict},
		{"InternalError", func() *AppError { return InternalError("boom", errors.New("x")) }, constants.CodeInternalError, http.StatusInternalServerError},
		{"DatabaseError", func() *AppError { return DatabaseError("db", errors.New("x")) }, constants.CodeDatabaseError, http.StatusInternalServerError},
		{"ValidationError", func() *AppError { return ValidationError("bad input") }, constants.CodeValidationError, http.StatusUnprocessableEntity},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.build()
			if err.StatusCode != tt.status {
				t.Errorf("status = %d, want %d", err.StatusCode, tt.status)
			}
			if err.Code != tt.code {
				t.Errorf("code = %q, want %q", err.Code, tt.code)
			}
			if err.Error() == "" {
				t.Error("Error() must not be empty")
			}
		})
	}
}

func TestAppErrorWrapsCause(t *testing.T) {
	cause := errors.New("underlying")
	err := DatabaseError("query failed", cause)
	if !errors.Is(err, cause) {
		t.Error("AppError must wrap its underlying error")
	}
	if !strings.Contains(err.Error(), "underlying") {
		t.Errorf("Error() should include cause: %q", err.Error())
	}
}

func TestParseDatabaseErrorUniqueViolation(t *testing.T) {
	pqErr := &pq.Error{Code: "23505", Constraint: "users_email_key"}
	appErr := ParseDatabaseError(pqErr)
	if appErr == nil {
		t.Fatal("expected a parsed error")
	}
	if appErr.Code != constants.CodeDuplicateEntry {
		t.Errorf("code = %q, want DUPLICATE_ENTRY", appErr.Code)
	}
	if !strings.Contains(appErr.Message, "email") {
		t.Errorf("message should mention email field: %q", appErr.Message)
	}
}

func TestParseDatabaseErrorGeneric(t *testing.T) {
	appErr := ParseDatabaseError(errors.New("connection reset"))
	if appErr == nil {
		t.Fatal("expected a parsed error")
	}
	if appErr.Code != constants.CodeDatabaseError {
		t.Errorf("code = %q, want DATABASE_ERROR", appErr.Code)
	}
}

func TestParseDatabaseErrorNil(t *testing.T) {
	if got := ParseDatabaseError(nil); got != nil {
		t.Errorf("expected nil for nil input, got %#v", got)
	}
}

func TestUniqueViolationFallbackFieldExtraction(t *testing.T) {
	pqErr := &pq.Error{Code: "23505", Constraint: "orders_user_token_key"}
	appErr := ParseDatabaseError(pqErr)
	if appErr.Message == "" {
		t.Error("expected a message for unknown constraint")
	}
}

func TestSqlNoRowsIsNotFoundStyle(t *testing.T) {
	// sql.ErrNoRows is not a pq error; default is DATABASE_ERROR.
	appErr := ParseDatabaseError(sql.ErrNoRows)
	if appErr == nil {
		t.Fatal("expected a parsed error")
	}
	if appErr.Code != constants.CodeDatabaseError {
		t.Errorf("code = %q, want DATABASE_ERROR", appErr.Code)
	}
}
