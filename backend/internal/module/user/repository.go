package usermodule

import (
	"context"
	"database/sql"
	"fmt"
	"golang/database"
)

// getExecutor resolves the active transaction from context if one was started
// by the handler, middleware, or service. Otherwise, it falls back to the database connection pool.
func (r *userRepository) getExecutor(ctx context.Context) database.QueryerExecer {
	return database.GetExecutor(ctx, r.db)
}

// GetAll retrieves all active users.
func (r *userRepository) GetAll(ctx context.Context) ([]User, error) {
	var users []User
	query := "SELECT id, name, email, phone, is_active, created_at, updated_at FROM users WHERE is_active = true ORDER BY name"
	err := r.getExecutor(ctx).SelectContext(ctx, &users, query)
	return users, err
}

// GetByID retrieves an active user by ID.
func (r *userRepository) GetByID(ctx context.Context, id string) (*User, error) {
	var user User
	query := "SELECT id, name, email, phone, password, is_active, created_at, updated_at FROM users WHERE id = $1 AND is_active = true"
	err := r.getExecutor(ctx).GetContext(ctx, &user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetByEmail retrieves a user by email.
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	query := "SELECT id, name, email, phone, password, is_active, created_at, updated_at FROM users WHERE email = $1"
	err := r.getExecutor(ctx).GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// Create inserts a new user using the current executor.
func (r *userRepository) Create(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (id, name, email, phone, password, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.getExecutor(ctx).ExecContext(ctx, query, user.ID, user.Name, user.Email, user.Phone, user.Password, user.IsActive, user.CreatedAt, user.UpdatedAt)
	return err
}

// Update applies the given columns and returns the updated user.
func (r *userRepository) Update(ctx context.Context, id string, updates map[string]interface{}) (*User, error) {
	query := "UPDATE users SET updated_at = NOW()"
	args := []interface{}{id}

	if name, ok := updates["name"].(string); ok && name != "" {
		query += fmt.Sprintf(", name = $%d", len(args)+1)
		args = append(args, name)
	}
	if email, ok := updates["email"].(string); ok && email != "" {
		query += fmt.Sprintf(", email = $%d", len(args)+1)
		args = append(args, email)
	}
	if phone, ok := updates["phone"].(string); ok && phone != "" {
		query += fmt.Sprintf(", phone = $%d", len(args)+1)
		args = append(args, phone)
	}

	query += " WHERE id = $1 RETURNING id, name, email, phone, is_active, created_at, updated_at"

	var user User
	if err := r.getExecutor(ctx).QueryRowxContext(ctx, query, args...).StructScan(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

// Delete soft deletes a user.
func (r *userRepository) Delete(ctx context.Context, id string) error {
	result, err := r.getExecutor(ctx).ExecContext(ctx, "UPDATE users SET is_active = false, updated_at = NOW() WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// HardDelete permanently deletes a user.
func (r *userRepository) HardDelete(ctx context.Context, id string) error {
	result, err := r.getExecutor(ctx).ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ExistsByID checks whether a user exists.
func (r *userRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	var exists bool
	err := r.getExecutor(ctx).GetContext(ctx, &exists, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", id)
	return exists, err
}

// ExistsByEmail checks whether an email is taken.
func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.getExecutor(ctx).GetContext(ctx, &exists, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email)
	return exists, err
}
