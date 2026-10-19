package errors

import (
	"fmt"
	"strings"

	"golang/pkg/constants"

	"github.com/lib/pq"
)

// AppError represents a custom application error
type AppError struct {
	Message    string
	Code       string
	StatusCode int
	Err        error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the wrapped cause so the error tree works with the standard
// library errors.Is/As helpers.
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new application error
func NewAppError(message, code string, statusCode int, err error) *AppError {
	return &AppError{
		Message:    message,
		Code:       code,
		StatusCode: statusCode,
		Err:        err,
	}
}

// BadRequest creates a 400 error
func BadRequest(message string) *AppError {
	return &AppError{
		Message:    message,
		Code:       constants.CodeBadRequest,
		StatusCode: 400,
	}
}

// Unauthorized creates a 401 error
func Unauthorized(message string) *AppError {
	return &AppError{
		Message:    message,
		Code:       constants.CodeUnauthorized,
		StatusCode: 401,
	}
}

// Forbidden creates a 403 error
func Forbidden(message string) *AppError {
	return &AppError{
		Message:    message,
		Code:       constants.CodeForbidden,
		StatusCode: 403,
	}
}

// NotFound creates a 404 error
func NotFound(message string) *AppError {
	return &AppError{
		Message:    message,
		Code:       constants.CodeNotFound,
		StatusCode: 404,
	}
}

// Conflict creates a 409 error
func Conflict(message string) *AppError {
	return &AppError{
		Message:    message,
		Code:       constants.CodeConflict,
		StatusCode: 409,
	}
}

// InternalError creates a 500 error
func InternalError(message string, err error) *AppError {
	return &AppError{
		Message:    message,
		Code:       constants.CodeInternalError,
		StatusCode: 500,
		Err:        err,
	}
}

// DatabaseError creates a database error
func DatabaseError(message string, err error) *AppError {
	return &AppError{
		Message:    message,
		Code:       constants.CodeDatabaseError,
		StatusCode: 500,
		Err:        err,
	}
}

// ValidationError creates a validation error
func ValidationError(message string) *AppError {
	return &AppError{
		Message:    message,
		Code:       constants.CodeValidationError,
		StatusCode: 422,
	}
}

// ParseDatabaseError parses PostgreSQL errors and returns user-friendly messages
func ParseDatabaseError(err error) *AppError {
	if err == nil {
		return nil
	}

	// Check if it's a PostgreSQL error
	if pqErr, ok := err.(*pq.Error); ok {
		return handlePostgresError(pqErr)
	}

	// Default database error
	return DatabaseError("Database operation failed", err)
}

// handlePostgresError handles specific PostgreSQL error codes
func handlePostgresError(pqErr *pq.Error) *AppError {
	switch pqErr.Code {
	case "23505": // unique_violation
		return handleUniqueViolation(pqErr)
	case "23503": // foreign_key_violation
		return handleForeignKeyViolation(pqErr)
	case "23502": // not_null_violation
		return handleNotNullViolation(pqErr)
	case "23514": // check_violation
		return handleCheckViolation(pqErr)
	case "42P01": // undefined_table
		return InternalError("Database table not found", pqErr)
	case "42703": // undefined_column
		return InternalError("Database column not found", pqErr)
	case "08006": // connection_failure
		return InternalError("Database connection failed", pqErr)
	case "08003": // connection_does_not_exist
		return InternalError("Database connection does not exist", pqErr)
	case "08001": // sqlclient_unable_to_establish_sqlconnection
		return InternalError("Unable to establish database connection", pqErr)
	default:
		return DatabaseError(fmt.Sprintf("Database error: %s", pqErr.Message), pqErr)
	}
}

// handleUniqueViolation handles unique constraint violations
func handleUniqueViolation(pqErr *pq.Error) *AppError {
	constraintName := pqErr.Constraint

	// Try to find the field name from constraint map
	if fieldName, exists := constants.UniqueConstraintFieldMap[constraintName]; exists {
		return &AppError{
			Message:    fmt.Sprintf("The %s already exists", fieldName),
			Code:       constants.CodeDuplicateEntry,
			StatusCode: 409,
			Err:        pqErr,
		}
	}

	// Extract field name from constraint name if not in map
	// Example: "users_email_key" -> "email"
	parts := strings.Split(constraintName, "_")
	if len(parts) >= 2 {
		fieldName := parts[len(parts)-2] // Get second to last part
		return &AppError{
			Message:    fmt.Sprintf("The %s already exists", fieldName),
			Code:       constants.CodeDuplicateEntry,
			StatusCode: 409,
			Err:        pqErr,
		}
	}

	// Generic duplicate message
	return &AppError{
		Message:    "This record already exists",
		Code:       constants.CodeDuplicateEntry,
		StatusCode: 409,
		Err:        pqErr,
	}
}

// handleForeignKeyViolation handles foreign key constraint violations
func handleForeignKeyViolation(pqErr *pq.Error) *AppError {
	constraintName := pqErr.Constraint

	// Try to find the relation name from constraint map
	if relationName, exists := constants.ForeignKeyConstraintMap[constraintName]; exists {
		// Check if it's a delete/update violation
		if strings.Contains(pqErr.Detail, "still referenced") {
			return &AppError{
				Message:    fmt.Sprintf("Cannot delete this record because it is still referenced by %s", relationName),
				Code:       constants.CodeForeignKeyViolation,
				StatusCode: 409,
				Err:        pqErr,
			}
		}
		return &AppError{
			Message:    fmt.Sprintf("The referenced %s does not exist", relationName),
			Code:       constants.CodeForeignKeyViolation,
			StatusCode: 422,
			Err:        pqErr,
		}
	}

	// Generic foreign key message
	if strings.Contains(pqErr.Detail, "still referenced") {
		return &AppError{
			Message:    "Cannot delete this record because it is still being used",
			Code:       constants.CodeForeignKeyViolation,
			StatusCode: 409,
			Err:        pqErr,
		}
	}

	return &AppError{
		Message:    "Referenced record does not exist",
		Code:       constants.CodeForeignKeyViolation,
		StatusCode: 422,
		Err:        pqErr,
	}
}

// handleNotNullViolation handles not null constraint violations
func handleNotNullViolation(pqErr *pq.Error) *AppError {
	column := pqErr.Column
	if column != "" {
		return &AppError{
			Message:    fmt.Sprintf("Field '%s' is required", column),
			Code:       constants.CodeValidationError,
			StatusCode: 422,
			Err:        pqErr,
		}
	}
	return &AppError{
		Message:    "Required field is missing",
		Code:       constants.CodeValidationError,
		StatusCode: 422,
		Err:        pqErr,
	}
}

// handleCheckViolation handles check constraint violations
func handleCheckViolation(pqErr *pq.Error) *AppError {
	constraintName := pqErr.Constraint
	return &AppError{
		Message:    fmt.Sprintf("Data validation failed: %s", constraintName),
		Code:       constants.CodeValidationError,
		StatusCode: 422,
		Err:        pqErr,
	}
}
