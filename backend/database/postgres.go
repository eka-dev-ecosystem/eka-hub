package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"golang/config"
	customErrors "golang/pkg/errors"
	"golang/pkg/logging"

	"github.com/jmoiron/sqlx"
	// Register the postgres driver via its init side effect.
	_ "github.com/lib/pq"
)

type PostgresDB struct {
	DB *sqlx.DB
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(cfg *config.DatabaseConfig) (*PostgresDB, error) {
	dsn := cfg.DSN()

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Hour)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logging.Info("connected to PostgreSQL",
		slog.String("host", cfg.Host),
		slog.String("dbname", cfg.DBName),
	)

	return &PostgresDB{DB: db}, nil
}

// Close closes the database connection
func (p *PostgresDB) Close() error {
	return p.DB.Close()
}

// HealthCheck checks if database connection is alive
func (p *PostgresDB) HealthCheck() error {
	return p.DB.Ping()
}

// GetDB returns the underlying sqlx.DB instance
func (p *PostgresDB) GetDB() *sqlx.DB {
	return p.DB
}

// WithTransaction executes a function within a database transaction
// Automatically handles BEGIN, COMMIT, and ROLLBACK
// If the function returns an error, the transaction is rolled back
func (p *PostgresDB) WithTransaction(fn func(tx *sqlx.Tx) error) error {
	// Start transaction
	tx, err := p.DB.Beginx()
	if err != nil {
		return customErrors.DatabaseError("Failed to begin transaction", err)
	}

	// Ensure transaction is finalized
	defer func() {
		if r := recover(); r != nil {
			// Rollback on panic
			if rbErr := tx.Rollback(); rbErr != nil {
				logging.Error("postgres rollback failed after panic", logging.Err(rbErr))
			}
			panic(r) // Re-throw panic after rollback
		}
	}()

	// Execute the function
	err = fn(tx)
	if err != nil {
		// Rollback on error
		if rbErr := tx.Rollback(); rbErr != nil {
			logging.Error("postgres transaction rollback failed", logging.Err(rbErr))
			return customErrors.DatabaseError("Transaction rollback failed", rbErr)
		}

		// Return the original error (might be AppError or database error)
		if appErr, ok := err.(*customErrors.AppError); ok {
			return appErr
		}
		return customErrors.ParseDatabaseError(err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return customErrors.DatabaseError("Failed to commit transaction", err)
	}

	return nil
}

// WithTransactionCtx executes a function within a database transaction with context
// Supports context cancellation and timeout
func (p *PostgresDB) WithTransactionCtx(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
	// Start transaction with context
	tx, err := p.DB.BeginTxx(ctx, nil)
	if err != nil {
		return customErrors.DatabaseError("Failed to begin transaction", err)
	}

	// Ensure transaction is finalized
	defer func() {
		if r := recover(); r != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				logging.Error("postgres rollback failed after panic", logging.Err(rbErr))
			}
			panic(r)
		}
	}()

	// Check if context is already canceled
	select {
	case <-ctx.Done():
		_ = tx.Rollback()
		return ctx.Err()
	default:
	}

	// Execute the function
	err = fn(tx)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			logging.Error("postgres transaction rollback failed", logging.Err(rbErr))
			return customErrors.DatabaseError("Transaction rollback failed", rbErr)
		}

		if appErr, ok := err.(*customErrors.AppError); ok {
			return appErr
		}
		return customErrors.ParseDatabaseError(err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return customErrors.DatabaseError("Failed to commit transaction", err)
	}

	return nil
}

// Transact is a convenience wrapper that returns the result of a transaction
// Use this when you need to return data from the transaction
func Transact[T any](db *PostgresDB, fn func(tx *sqlx.Tx) (T, error)) (T, error) {
	var result T

	err := db.WithTransaction(func(tx *sqlx.Tx) error {
		var err error
		result, err = fn(tx)
		return err
	})

	return result, err
}

// TransactCtx is a convenience wrapper with context support
func TransactCtx[T any](ctx context.Context, db *PostgresDB, fn func(tx *sqlx.Tx) (T, error)) (T, error) {
	var result T

	err := db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		var err error
		result, err = fn(tx)
		return err
	})

	return result, err
}

// WithReadOnlyTransactionCtx executes a function within a read-only database transaction with context
func (p *PostgresDB) WithReadOnlyTransactionCtx(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
	tx, err := p.DB.BeginTxx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return customErrors.DatabaseError("Failed to begin read-only transaction", err)
	}

	defer func() {
		if r := recover(); r != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				logging.Error("postgres rollback failed after panic in read-only tx", logging.Err(rbErr))
			}
			panic(r)
		}
	}()

	select {
	case <-ctx.Done():
		_ = tx.Rollback()
		return ctx.Err()
	default:
	}

	err = fn(tx)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			logging.Error("postgres read-only transaction rollback failed", logging.Err(rbErr))
			return customErrors.DatabaseError("Transaction rollback failed", rbErr)
		}

		if appErr, ok := err.(*customErrors.AppError); ok {
			return appErr
		}
		return customErrors.ParseDatabaseError(err)
	}

	if err := tx.Commit(); err != nil {
		return customErrors.DatabaseError("Failed to commit read-only transaction", err)
	}

	return nil
}

// WithReadOnlyTransaction executes a function within a read-only database transaction
func (p *PostgresDB) WithReadOnlyTransaction(fn func(tx *sqlx.Tx) error) error {
	return p.WithReadOnlyTransactionCtx(context.Background(), fn)
}

// TransactReadOnly is a convenience wrapper for read-only transactions
func TransactReadOnly[T any](db *PostgresDB, fn func(tx *sqlx.Tx) (T, error)) (T, error) {
	return TransactReadOnlyCtx(context.Background(), db, fn)
}

// TransactReadOnlyCtx is a convenience wrapper for read-only transactions with context
func TransactReadOnlyCtx[T any](ctx context.Context, db *PostgresDB, fn func(tx *sqlx.Tx) (T, error)) (T, error) {
	var result T

	err := db.WithReadOnlyTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		var err error
		result, err = fn(tx)
		return err
	})

	return result, err
}

// TransactSqlxCtx executes a function within a transaction on raw *sqlx.DB with context
func TransactSqlxCtx[T any](ctx context.Context, db *sqlx.DB, fn func(tx *sqlx.Tx) (T, error)) (T, error) {
	pgDB := &PostgresDB{DB: db}
	return TransactCtx(ctx, pgDB, fn)
}

// WithTransactionSqlxCtx executes a function within a transaction on raw *sqlx.DB
func WithTransactionSqlxCtx(ctx context.Context, db *sqlx.DB, fn func(tx *sqlx.Tx) error) error {
	pgDB := &PostgresDB{DB: db}
	return pgDB.WithTransactionCtx(ctx, fn)
}

// TransactSqlx executes a function within a database transaction on raw *sqlx.DB
func TransactSqlx[T any](db *sqlx.DB, fn func(tx *sqlx.Tx) (T, error)) (T, error) {
	pgDB := &PostgresDB{DB: db}
	return Transact(pgDB, fn)
}

// WithTransactionSqlx executes a function within a transaction on raw *sqlx.DB
func WithTransactionSqlx(db *sqlx.DB, fn func(tx *sqlx.Tx) error) error {
	pgDB := &PostgresDB{DB: db}
	return pgDB.WithTransaction(fn)
}

// TransactReadOnlySqlx executes a function within a read-only transaction on raw *sqlx.DB
func TransactReadOnlySqlx[T any](db *sqlx.DB, fn func(tx *sqlx.Tx) (T, error)) (T, error) {
	pgDB := &PostgresDB{DB: db}
	return TransactReadOnly(pgDB, fn)
}

// TransactReadOnlySqlxCtx executes a function within a read-only transaction on raw *sqlx.DB with context
func TransactReadOnlySqlxCtx[T any](ctx context.Context, db *sqlx.DB, fn func(tx *sqlx.Tx) (T, error)) (T, error) {
	pgDB := &PostgresDB{DB: db}
	return TransactReadOnlyCtx(ctx, pgDB, fn)
}

// WithReadOnlyTransactionSqlx executes a function within a read-only transaction on raw *sqlx.DB
func WithReadOnlyTransactionSqlx(db *sqlx.DB, fn func(tx *sqlx.Tx) error) error {
	pgDB := &PostgresDB{DB: db}
	return pgDB.WithReadOnlyTransaction(fn)
}

// WithReadOnlyTransactionSqlxCtx executes a function within a read-only transaction on raw *sqlx.DB with context
func WithReadOnlyTransactionSqlxCtx(ctx context.Context, db *sqlx.DB, fn func(tx *sqlx.Tx) error) error {
	pgDB := &PostgresDB{DB: db}
	return pgDB.WithReadOnlyTransactionCtx(ctx, fn)
}

type txContextKey struct{}

// ContextWithTx injects an active *sqlx.Tx into context.
func ContextWithTx(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}

// TxFromContext extracts an active *sqlx.Tx from context if present.
func TxFromContext(ctx context.Context) *sqlx.Tx {
	if ctx == nil {
		return nil
	}
	if tx, ok := ctx.Value(txContextKey{}).(*sqlx.Tx); ok {
		return tx
	}
	return nil
}

// QueryerExecer defines the interface implemented by both *sqlx.DB and *sqlx.Tx.
type QueryerExecer interface {
	sqlx.ExtContext
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
}

// GetExecutor returns the active transaction from context if available, otherwise the provided db.
func GetExecutor(ctx context.Context, db *sqlx.DB) QueryerExecer {
	if tx := TxFromContext(ctx); tx != nil {
		return tx
	}
	return db
}

// Transactor defines the interface for running operations within transaction boundaries.
type Transactor interface {
	RunInTx(ctx context.Context, fn func(txCtx context.Context) error) error
	WithTransactionContext(ctx context.Context, fn func(txCtx context.Context) error) error
}

// WithTransactionContext executes a function within a database transaction,
// automatically injecting the transaction into the context so downstream services/repos use it.
// If the context already contains an active transaction, it is reused.
func (p *PostgresDB) WithTransactionContext(ctx context.Context, fn func(txCtx context.Context) error) error {
	if TxFromContext(ctx) != nil {
		return fn(ctx)
	}

	return p.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		txCtx := ContextWithTx(ctx, tx)
		return fn(txCtx)
	})
}

// RunInTx is an alias for WithTransactionContext.
func (p *PostgresDB) RunInTx(ctx context.Context, fn func(txCtx context.Context) error) error {
	return p.WithTransactionContext(ctx, fn)
}

// WithTransactionContext executes a function within a transaction using raw *sqlx.DB.
func WithTransactionContext(ctx context.Context, db *sqlx.DB, fn func(txCtx context.Context) error) error {
	pgDB := &PostgresDB{DB: db}
	return pgDB.WithTransactionContext(ctx, fn)
}
