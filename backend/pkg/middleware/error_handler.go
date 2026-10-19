package middleware

import (
	"database/sql"
	"errors"
	"log/slog"

	"golang/pkg/constants"
	customErrors "golang/pkg/errors"
	"golang/pkg/logging"
	"golang/pkg/response"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// ErrorHandler is a global error handling middleware for Fiber
func ErrorHandler() fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		// Log the error for debugging, correlated with the request ID
		// (LoggerFromFiber already carries the request_id attribute).
		logging.LoggerFromFiber(c).Error("request failed",
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", statusCodeFrom(err)),
			logging.Err(err),
		)

		// Handle custom AppError
		if appErr, ok := err.(*customErrors.AppError); ok {
			return response.ErrorResponse(c, appErr.StatusCode, appErr.Message, appErr.Code)
		}

		// Handle Fiber errors
		if fiberErr, ok := err.(*fiber.Error); ok {
			code := getCodeFromStatus(fiberErr.Code)
			return response.ErrorResponse(c, fiberErr.Code, fiberErr.Message, code)
		}

		// Handle validator errors
		if validationErrs, ok := err.(validator.ValidationErrors); ok {
			errors := formatValidationErrors(validationErrs)
			return response.ValidationErrorResponse(c, errors)
		}

		// Handle sql.ErrNoRows (not found)
		if err == sql.ErrNoRows {
			return response.ErrorResponse(c, fiber.StatusNotFound, "Record not found", constants.CodeNotFound)
		}

		// Handle database errors (try to parse PostgreSQL errors)
		if dbErr := customErrors.ParseDatabaseError(err); dbErr != nil {
			return response.ErrorResponse(c, dbErr.StatusCode, dbErr.Message, dbErr.Code)
		}

		// Default internal server error
		return response.ErrorResponse(
			c,
			fiber.StatusInternalServerError,
			constants.MsgInternalError,
			constants.CodeInternalError,
		)
	}
}

// statusCodeFrom derives the intended HTTP status of an error so the log
// entry carries the real code (at this point Fiber hasn't written it yet).
func statusCodeFrom(err error) int {
	switch e := err.(type) {
	case *customErrors.AppError:
		return e.StatusCode
	case *fiber.Error:
		return e.Code
	case validator.ValidationErrors:
		return fiber.StatusBadRequest
	}
	if errors.Is(err, sql.ErrNoRows) {
		return fiber.StatusNotFound
	}
	return fiber.StatusInternalServerError
}

// getCodeFromStatus maps HTTP status codes to error codes
func getCodeFromStatus(status int) string {
	switch status {
	case fiber.StatusBadRequest:
		return constants.CodeBadRequest
	case fiber.StatusUnauthorized:
		return constants.CodeUnauthorized
	case fiber.StatusForbidden:
		return constants.CodeForbidden
	case fiber.StatusNotFound:
		return constants.CodeNotFound
	case fiber.StatusConflict:
		return constants.CodeConflict
	case fiber.StatusUnprocessableEntity:
		return constants.CodeValidationError
	case fiber.StatusInternalServerError:
		return constants.CodeInternalError
	default:
		return constants.CodeInternalError
	}
}

// formatValidationErrors formats validator errors into a readable format
func formatValidationErrors(validationErrs validator.ValidationErrors) []map[string]string {
	errors := make([]map[string]string, 0)

	for _, err := range validationErrs {
		fieldError := map[string]string{
			"field":   err.Field(),
			"tag":     err.Tag(),
			"value":   err.Param(),
			"message": getValidationErrorMessage(err),
		}
		errors = append(errors, fieldError)
	}

	return errors
}

// getValidationErrorMessage returns a user-friendly validation error message
func getValidationErrorMessage(err validator.FieldError) string {
	field := err.Field()

	switch err.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email address"
	case "min":
		return field + " must be at least " + err.Param() + " characters"
	case "max":
		return field + " must be at most " + err.Param() + " characters"
	case "len":
		return field + " must be exactly " + err.Param() + " characters"
	case "eq":
		return field + " must be equal to " + err.Param()
	case "ne":
		return field + " must not be equal to " + err.Param()
	case "gt":
		return field + " must be greater than " + err.Param()
	case "gte":
		return field + " must be greater than or equal to " + err.Param()
	case "lt":
		return field + " must be less than " + err.Param()
	case "lte":
		return field + " must be less than or equal to " + err.Param()
	case "alpha":
		return field + " must contain only alphabetic characters"
	case "alphanum":
		return field + " must contain only alphanumeric characters"
	case "numeric":
		return field + " must be a number"
	case "url":
		return field + " must be a valid URL"
	case "uuid":
		return field + " must be a valid UUID"
	case "oneof":
		return field + " must be one of: " + err.Param()
	default:
		return field + " validation failed on " + err.Tag()
	}
}

// RecoverMiddleware recovers from panics and returns error response
func RecoverMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				logging.Error("panic recovered",
					slog.String("method", c.Method()),
					slog.String("path", c.Path()),
					slog.Any("panic", r),
				)
				if err := response.ErrorResponse(
					c,
					fiber.StatusInternalServerError,
					"An unexpected error occurred",
					constants.CodeInternalError,
				); err != nil {
					logging.Error("failed to write panic response", logging.Err(err))
				}
			}
		}()
		return c.Next()
	}
}
