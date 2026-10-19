package transaction

import (
	"context"
	"golang/database"

	"github.com/gofiber/fiber/v2"
)

// Middleware wraps the request handler execution inside an active database transaction.
// If the handler completes without error, the transaction is committed.
// If the handler returns an error or panics, the transaction is automatically rolled back.
// Any downstream service or repository calls that inspect context will automatically participate
// in this transaction.
func Middleware(transactor database.Transactor) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return transactor.WithTransactionContext(c.UserContext(), func(txCtx context.Context) error {
			c.SetUserContext(txCtx)
			return c.Next()
		})
	}
}
