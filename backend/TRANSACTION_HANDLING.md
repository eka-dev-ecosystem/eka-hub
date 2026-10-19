# Database Transaction Handling

Complete transaction management system similar to TypeScript's `withTransaction` helper.

## 🎯 Features

- ✅ Automatic BEGIN, COMMIT, ROLLBACK
- ✅ Panic recovery with automatic rollback
- ✅ Context support (timeout, cancellation)
- ✅ Generic return types
- ✅ PostgreSQL error parsing
- ✅ Nested transaction support

## 📦 Transaction Methods

### 1. WithTransaction - Basic Transaction

Simple transaction without return value.

```go
err := db.WithTransaction(func (tx *sqlx.Tx) error {
// Your database operations here
_, err := tx.Exec("INSERT INTO users ...")
if err != nil {
return err // Automatic rollback
}

_, err = tx.Exec("INSERT INTO profiles ...")
if err != nil {
return err // Automatic rollback
}

return nil // Automatic commit
})

if err != nil {
// Handle error
}
```

### 2. WithTransactionCtx - With Context Support

Transaction with timeout and cancellation support.

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := db.WithTransactionCtx(ctx, func (tx *sqlx.Tx) error {
// Your operations with context
_, err := tx.ExecContext(ctx, "INSERT INTO users ...")
return err
})
```

### 3. Transact - With Return Value

Transaction that returns data.

```go
user, err := database.Transact(db, func(tx *sqlx.Tx) (*User, error) {
user := &User{ID: uuid.New().String(), Name: "John"}

_, err := tx.Exec("INSERT INTO users (id, name) VALUES ($1, $2)",
user.ID, user.Name)
if err != nil {
return nil, err
}

return user, nil
})

if err != nil {
// Handle error
}

// Use user data
fmt.Println(user.Name)
```

### 4. TransactCtx - With Context and Return Value

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

result, err := database.TransactCtx(ctx, db, func (tx *sqlx.Tx) (map[string]interface{}, error) {
// Your operations
return map[string]interface{}{
"id": "123",
"status": "success",
}, nil
})
```

## 💻 Usage Examples

### Example 1: Create Order with Items

```go
func CreateOrder(db *database.PostgresDB) fiber.Handler {
return func (c *fiber.Ctx) error {
var req CreateOrderRequest

if err := c.BodyParser(&req); err != nil {
return customErrors.BadRequest("Invalid request")
}

// Transaction with return value
order, err := database.Transact(db, func (tx *sqlx.Tx) (*Order, error) {
// 1. Create order
order := &Order{
ID:     uuid.New().String(),
UserID: req.UserID,
Total:  0,
}

_, err := tx.Exec(
"INSERT INTO orders (id, user_id, total) VALUES ($1, $2, $3)",
order.ID, order.UserID, order.Total,
)
if err != nil {
return nil, customErrors.ParseDatabaseError(err)
}

// 2. Create order items
var total float64
for _, item := range req.Items {
// Get product price
var price float64
err := tx.Get(&price,
"SELECT price FROM products WHERE id = $1",
item.ProductID,
)
if err != nil {
if err == sql.ErrNoRows {
return nil, customErrors.NotFound("Product not found")
}
return nil, err
}

// Insert order item
_, err = tx.Exec(
"INSERT INTO order_items (id, order_id, product_id, quantity, price) VALUES ($1, $2, $3, $4, $5)",
uuid.New().String(), order.ID, item.ProductID, item.Quantity, price,
)
if err != nil {
return nil, customErrors.ParseDatabaseError(err)
}

total += price * float64(item.Quantity)
}

// 3. Update order total
_, err = tx.Exec(
"UPDATE orders SET total = $1 WHERE id = $2",
total, order.ID,
)
if err != nil {
return nil, err
}

order.Total = total
return order, nil
})

if err != nil {
return err // Error is already formatted
}

return response.CreatedResponse(c, order, "Order created successfully")
}
}
```

### Example 2: Transfer Funds

```go
func TransferFunds(db *database.PostgresDB) fiber.Handler {
return func (c *fiber.Ctx) error {
var req TransferRequest

if err := c.BodyParser(&req); err != nil {
return customErrors.BadRequest("Invalid request")
}

ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
defer cancel()

result, err := database.TransactCtx(ctx, db, func(tx *sqlx.Tx) (map[string]interface{}, error) {
// 1. Deduct from sender
result, err := tx.Exec(
"UPDATE wallets SET balance = balance - $1 WHERE user_id = $2 AND balance >= $1",
req.Amount, req.FromUserID,
)
if err != nil {
return nil, customErrors.ParseDatabaseError(err)
}

rowsAffected, _ := result.RowsAffected()
if rowsAffected == 0 {
return nil, customErrors.BadRequest("Insufficient balance")
}

// 2. Add to receiver
_, err = tx.Exec(
"UPDATE wallets SET balance = balance + $1 WHERE user_id = $2",
req.Amount, req.ToUserID,
)
if err != nil {
return nil, customErrors.ParseDatabaseError(err)
}

// 3. Create transaction record
txID := uuid.New().String()
_, err = tx.Exec(
"INSERT INTO transactions (id, from_user_id, to_user_id, amount) VALUES ($1, $2, $3, $4)",
txID, req.FromUserID, req.ToUserID, req.Amount,
)
if err != nil {
return nil, err
}

return map[string]interface{}{
"transaction_id": txID,
"amount": req.Amount,
}, nil
})

if err != nil {
return err
}

return response.SuccessResponse(c, result, "Transfer completed")
}
}
```

### Example 3: Batch Update

```go
func BatchUpdateUsers(db *database.PostgresDB) fiber.Handler {
return func (c *fiber.Ctx) error {
var req BatchUpdateRequest

if err := c.BodyParser(&req); err != nil {
return customErrors.BadRequest("Invalid request")
}

err := db.WithTransaction(func (tx *sqlx.Tx) error {
// Update all users
query := `UPDATE users SET status = $1, updated_at = NOW() WHERE id = ANY($2)`

result, err := tx.Exec(query, req.Status, req.UserIDs)
if err != nil {
return customErrors.ParseDatabaseError(err)
}

rowsAffected, _ := result.RowsAffected()
if rowsAffected == 0 {
return customErrors.NotFound("No users found")
}

// Create audit log
_, err = tx.Exec(
"INSERT INTO audit_logs (action, details) VALUES ($1, $2)",
"batch_update", req,
)
return err
})

if err != nil {
return err
}

return response.SuccessResponse(c, fiber.Map{
"updated_count": len(req.UserIDs),
}, "Users updated successfully")
}
}
```

### Example 4: Create User with Profile

```go
func CreateUserWithProfile(db *database.PostgresDB) fiber.Handler {
return func (c *fiber.Ctx) error {
var req UserProfileRequest

if err := c.BodyParser(&req); err != nil {
return customErrors.BadRequest("Invalid request")
}

result, err := database.Transact(db, func (tx *sqlx.Tx) (map[string]interface{}, error) {
// 1. Create user
userID := uuid.New().String()
var user User

err := tx.Get(&user,
`INSERT INTO users (id, name, email) 
                 VALUES ($1, $2, $3) 
                 RETURNING id, name, email, created_at`,
userID, req.Name, req.Email,
)
if err != nil {
return nil, customErrors.ParseDatabaseError(err)
}

// 2. Create profile
_, err = tx.Exec(
`INSERT INTO user_profiles (id, user_id, bio, avatar) 
                 VALUES ($1, $2, $3, $4)`,
uuid.New().String(), user.ID, req.Bio, req.Avatar,
)
if err != nil {
return nil, customErrors.ParseDatabaseError(err)
}

return map[string]interface{}{
"id":    user.ID,
"name":  user.Name,
"email": user.Email,
"bio":   req.Bio,
}, nil
})

if err != nil {
return err
}

return response.CreatedResponse(c, result, "User created successfully")
}
}
```

## 🔄 Error Handling

### Automatic Rollback

Transactions automatically rollback on:

- ✅ Any returned error
- ✅ Panic/crash
- ✅ Context cancellation
- ✅ Timeout

```go
err := db.WithTransaction(func (tx *sqlx.Tx) error {
_, err := tx.Exec("INSERT INTO users ...")
if err != nil {
return err // ← Automatic rollback
}

// This will also trigger rollback
panic("something went wrong")
})
```

### Custom Error Handling

```go
err := db.WithTransaction(func (tx *sqlx.Tx) error {
_, err := tx.Exec("INSERT INTO users (email) VALUES ($1)", email)
if err != nil {
// Parse PostgreSQL errors (duplicate, FK violations, etc.)
return customErrors.ParseDatabaseError(err)
}

// Or return custom errors
if someCondition {
return customErrors.BadRequest("Invalid data")
}

return nil
})

// Error is already formatted, just return it
if err != nil {
return err
}
```

### Context Timeout Handling

```go
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()

err := db.WithTransactionCtx(ctx, func (tx *sqlx.Tx) error {
// If this takes longer than 3 seconds, automatic rollback
_, err := tx.ExecContext(ctx, "UPDATE users SET status = 'active'")
return err
})

if err != nil {
if err == context.DeadlineExceeded {
return customErrors.InternalError("Operation timeout", err)
}
return err
}
```

## 🎯 Best Practices

### 1. Always Use Transactions for Multiple Operations

```go
// ❌ BAD: No transaction
_, err := db.DB.Exec("INSERT INTO orders ...")
_, err = db.DB.Exec("INSERT INTO order_items ...") // ← Might fail, order already created

// ✅ GOOD: With transaction
err := db.WithTransaction(func (tx *sqlx.Tx) error {
_, err := tx.Exec("INSERT INTO orders ...")
if err != nil {
return err
}

_, err = tx.Exec("INSERT INTO order_items ...")
return err // ← Both succeed or both rollback
})
```

### 2. Use Context for Long Operations

```go
// ✅ GOOD: Timeout protection
ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
defer cancel()

err := db.WithTransactionCtx(ctx, func (tx *sqlx.Tx) error {
// Your operations
return nil
})
```

### 3. Return Data from Transactions

```go
// ✅ GOOD: Use Transact to return data
user, err := database.Transact(db, func(tx *sqlx.Tx) (*User, error) {
var user User
err := tx.Get(&user, "INSERT INTO users ... RETURNING *")
return &user, err
})
```

### 4. Parse Database Errors

```go
err := db.WithTransaction(func (tx *sqlx.Tx) error {
_, err := tx.Exec("INSERT INTO users ...")
if err != nil {
// ✅ This handles duplicate email, FK violations, etc.
return customErrors.ParseDatabaseError(err)
}
return nil
})
```

### 5. Don't Forget to Check Rows Affected

```go
err := db.WithTransaction(func (tx *sqlx.Tx) error {
result, err := tx.Exec("UPDATE users SET name = $1 WHERE id = $2", name, id)
if err != nil {
return err
}

// ✅ Check if update actually happened
rowsAffected, _ := result.RowsAffected()
if rowsAffected == 0 {
return customErrors.NotFound("User not found")
}

return nil
})
```

## 🔍 Comparison with TypeScript

### TypeScript Version

```typescript
export async function withTransaction<T>(
    fn: (client: PoolClient) => Promise<T>
): Promise<T> {
    const client = await pool.connect();
    try {
        await client.query('BEGIN');
        const result = await fn(client);
        await client.query('COMMIT');
        return result;
    } catch (err) {
        try {
            await client.query('ROLLBACK');
        } catch (rollbackErr) {
            console.error('[postgres] rollback failed', rollbackErr);
        }
        if (err instanceof HttpError) {
            throw err;
        }
        formatPgError(err);
    } finally {
        client.release();
    }
}
```

### Go Version

```go
// Same functionality!
func Transact[T any](db *PostgresDB, fn func (tx *sqlx.Tx) (T, error)) (T, error) {
var result T

err := db.WithTransaction(func (tx *sqlx.Tx) error {
var err error
result, err = fn(tx)
return err
})

return result, err
}
```

### Usage Comparison

**TypeScript:**

```typescript
const user = await withTransaction(async (client) => {
    const result = await client.query('INSERT INTO users ...');
    return result.rows[0];
});
```

**Go:**

```go
user, err := database.Transact(db, func(tx *sqlx.Tx) (*User, error) {
var user User
err := tx.Get(&user, "INSERT INTO users ... RETURNING *")
return &user, err
})
```

## ✅ Complete Example

```go
package handlers

import (
	"github.com/gofiber/fiber/v2"
	"golang/database"
	customErrors "golang/pkg/errors"
	"golang/pkg/response"
)

func CreateOrderHandler(db *database.PostgresDB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req CreateOrderRequest

		// 1. Parse and validate
		if err := c.BodyParser(&req); err != nil {
			return customErrors.BadRequest("Invalid request")
		}

		if err := validate.Struct(req); err != nil {
			return err
		}

		// 2. Execute transaction
		order, err := database.Transact(db, func(tx *sqlx.Tx) (*Order, error) {
			// Create order
			order := &Order{ID: uuid.New().String()}
			_, err := tx.Exec("INSERT INTO orders ...", order.ID)
			if err != nil {
				return nil, customErrors.ParseDatabaseError(err)
			}

			// Create items
			for _, item := range req.Items {
				_, err := tx.Exec("INSERT INTO order_items ...", item)
				if err != nil {
					return nil, customErrors.ParseDatabaseError(err)
				}
			}

			return order, nil
		})

		// 3. Handle error
		if err != nil {
			return err
		}

		// 4. Return response
		return response.CreatedResponse(c, order, "Order created")
	}
}
```

## 🎉 Summary

Transaction system yang mirip dengan TypeScript:

- ✅ Automatic BEGIN/COMMIT/ROLLBACK
- ✅ Panic recovery
- ✅ Context support
- ✅ Generic return types
- ✅ Error parsing
- ✅ Easy to use

**Ready to use!** 🚀

