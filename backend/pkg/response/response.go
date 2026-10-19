package response

import (
	"time"

	"golang/pkg/constants"

	"github.com/gofiber/fiber/v2"
)

// BaseResponse represents the standard API response structure.
// Non-success responses always carry a machine-readable error code sourced
// from pkg/constants.
type BaseResponse[T any] struct {
	Message   string `json:"message"`
	Success   bool   `json:"success"`
	Data      *T     `json:"data"`
	Code      string `json:"code"`
	Timestamp string `json:"timestamp"`
}

// SuccessResponse creates a successful response.
func SuccessResponse[T any](c *fiber.Ctx, data T, message string) error {
	return c.JSON(BaseResponse[T]{
		Message:   message,
		Success:   true,
		Data:      &data,
		Code:      constants.CodeSuccess,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// SuccessMessageResponse creates a successful response without payload data,
// useful for delete/logout-style endpoints.
func SuccessMessageResponse(c *fiber.Ctx, message string) error {
	return c.JSON(BaseResponse[any]{
		Message:   message,
		Success:   true,
		Data:      nil,
		Code:      constants.CodeSuccess,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// SuccessResponseWithCode creates a successful response with custom code.
func SuccessResponseWithCode[T any](c *fiber.Ctx, data T, message string, code string) error {
	return c.JSON(BaseResponse[T]{
		Message:   message,
		Success:   true,
		Data:      &data,
		Code:      code,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// ErrorResponse creates an error response. The code must be one of the
// constants in pkg/constants; empty values fall back to the generic internal
// error code.
func ErrorResponse(c *fiber.Ctx, statusCode int, message string, code string) error {
	if code == "" {
		code = constants.CodeInternalError
	}
	return c.Status(statusCode).JSON(BaseResponse[any]{
		Message:   message,
		Success:   false,
		Data:      nil,
		Code:      code,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// ValidationErrorResponse creates a validation error response.
func ValidationErrorResponse[T any](c *fiber.Ctx, errors T) error {
	return c.Status(fiber.StatusBadRequest).JSON(BaseResponse[T]{
		Message:   "Validation failed",
		Success:   false,
		Data:      &errors,
		Code:      constants.CodeValidationError,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// CreatedResponse creates a 201 Created response.
func CreatedResponse[T any](c *fiber.Ctx, data T, message string) error {
	return c.Status(fiber.StatusCreated).JSON(BaseResponse[T]{
		Message:   message,
		Success:   true,
		Data:      &data,
		Code:      constants.CodeCreated,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// NoContentResponse creates a 204 No Content response.
func NoContentResponse(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// PaginatedResponse represents a paginated response.
type PaginatedResponse[T any] struct {
	Items      []T   `json:"items"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PerPage    int   `json:"perPage"`
	TotalPages int   `json:"totalPages"`
}

// PaginatedSuccessResponse creates a paginated success response.
func PaginatedSuccessResponse[T any](c *fiber.Ctx, data PaginatedResponse[T], message string) error {
	return c.JSON(BaseResponse[PaginatedResponse[T]]{
		Message:   message,
		Success:   true,
		Data:      &data,
		Code:      constants.CodeSuccess,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}
