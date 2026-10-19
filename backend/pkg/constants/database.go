package constants

// PostgreSQL constraint names mapped to field names
// This helps provide user-friendly error messages for database constraint violations
var UniqueConstraintFieldMap = map[string]string{
	"skills_name_key":      "name",
	"users_email_key":      "email",
	"users_username_key":   "username",
	"users_phone_key":      "phone",
	"products_sku_key":     "sku",
	"products_slug_key":    "slug",
	"categories_name_key":  "name",
	"categories_slug_key":  "slug",
	"roles_name_key":       "name",
	"permissions_name_key": "name",
}

// Foreign key constraint names mapped to readable names
var ForeignKeyConstraintMap = map[string]string{
	"users_role_id_fkey":                  "role",
	"orders_user_id_fkey":                 "user",
	"order_items_order_id_fkey":           "order",
	"order_items_product_id_fkey":         "product",
	"products_category_id_fkey":           "category",
	"user_permissions_user_id_fkey":       "user",
	"user_permissions_permission_id_fkey": "permission",
}

// Error codes
const (
	// Success codes
	CodeSuccess = "SUCCESS"
	CodeCreated = "CREATED"

	// Client error codes (4xx)
	CodeBadRequest          = "BAD_REQUEST"
	CodeValidationError     = "VALIDATION_ERROR"
	CodeUnauthorized        = "UNAUTHORIZED"
	CodeForbidden           = "FORBIDDEN"
	CodeNotFound            = "NOT_FOUND"
	CodeConflict            = "CONFLICT"
	CodeDuplicateEntry      = "DUPLICATE_ENTRY"
	CodeForeignKeyViolation = "FOREIGN_KEY_VIOLATION"
	CodeUnprocessableEntity = "UNPROCESSABLE_ENTITY"

	// Server error codes (5xx)
	CodeInternalError = "INTERNAL_SERVER_ERROR"
	CodeDatabaseError = "DATABASE_ERROR"
	CodeServiceError  = "SERVICE_ERROR"
)

// HTTP Status messages
const (
	MsgSuccess            = "Operation completed successfully"
	MsgCreated            = "Resource created successfully"
	MsgUpdated            = "Resource updated successfully"
	MsgDeleted            = "Resource deleted successfully"
	MsgBadRequest         = "Bad request"
	MsgValidationError    = "Validation error"
	MsgUnauthorized       = "Unauthorized access"
	MsgForbidden          = "Access forbidden"
	MsgNotFound           = "Resource not found"
	MsgConflict           = "Resource conflict"
	MsgInternalError      = "Internal server error"
	MsgDatabaseError      = "Database error occurred"
	MsgServiceUnavailable = "Service temporarily unavailable"
)
