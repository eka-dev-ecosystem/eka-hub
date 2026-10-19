package usermodule

import (
	"context"
	"database/sql"
	"errors"
	"golang/database"
	"time"

	appErrors "golang/pkg/errors"

	"github.com/google/uuid"
)

// UserService defines the business logic for the user domain.
type UserService interface {
	GetAll(ctx context.Context) ([]UserResponse, error)
	GetByID(ctx context.Context, id string) (*UserResponse, error)
	Create(ctx context.Context, req *CreateUserRequest) (*UserResponse, error)
	Update(ctx context.Context, id string, req *UpdateUserRequest) (*UserResponse, error)
	Delete(ctx context.Context, id string) error
	HardDelete(ctx context.Context, id string) error
}

// userService implements UserService.
type userService struct {
	database.Transactor
	repo UserRepository
}

// NewUserService builds the user service with required database.Transactor.
func NewUserService(repo UserRepository, transactor database.Transactor) UserService {
	return &userService{
		Transactor: transactor,
		repo:       repo,
	}
}

// runInTx delegates to the embedded database.Transactor.RunInTx for backward compatibility.
func (s *userService) runInTx(ctx context.Context, fn func(txCtx context.Context) error) error {
	return s.RunInTx(ctx, fn)
}

func (s *userService) GetAll(ctx context.Context) ([]UserResponse, error) {
	users, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, appErrors.DatabaseError("Failed to fetch users", err)
	}

	res := make([]UserResponse, 0, len(users))
	for i := range users {
		res = append(res, users[i].toResponse())
	}
	return res, nil
}

func (s *userService) GetByID(ctx context.Context, id string) (*UserResponse, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, appErrors.BadRequest("Invalid user ID format")
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, appErrors.DatabaseError("Failed to fetch user", err)
	}
	if user == nil {
		return nil, appErrors.NotFound("User not found")
	}

	u := user.toResponse()
	return &u, nil
}

func (s *userService) Create(ctx context.Context, req *CreateUserRequest) (*UserResponse, error) {
	var res *UserResponse
	err := s.runInTx(ctx, func(txCtx context.Context) error {
		exists, err := s.repo.ExistsByEmail(txCtx, req.Email)
		if err != nil {
			return appErrors.DatabaseError("Failed to check email existence", err)
		}
		if exists {
			return appErrors.Conflict("Email already exists")
		}

		now := time.Now()
		user := &User{
			ID:        uuid.New().String(),
			Name:      req.Name,
			Email:     req.Email,
			Phone:     sql.NullString{String: req.Phone, Valid: req.Phone != ""},
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if err := s.repo.Create(txCtx, user); err != nil {
			return appErrors.ParseDatabaseError(err)
		}

		u := user.toResponse()
		res = &u
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *userService) Update(ctx context.Context, id string, req *UpdateUserRequest) (*UserResponse, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, appErrors.BadRequest("Invalid user ID format")
	}

	var res *UserResponse
	err := s.runInTx(ctx, func(txCtx context.Context) error {
		exists, err := s.repo.ExistsByID(txCtx, id)
		if err != nil {
			return appErrors.DatabaseError("Failed to check user existence", err)
		}
		if !exists {
			return appErrors.NotFound("User not found")
		}

		if req.Email != "" {
			existing, err := s.repo.GetByEmail(txCtx, req.Email)
			if err != nil {
				return appErrors.DatabaseError("Failed to check email existence", err)
			}
			if existing != nil && existing.ID != id {
				return appErrors.Conflict("Email already exists")
			}
		}

		updates := map[string]interface{}{}
		if req.Name != "" {
			updates["name"] = req.Name
		}
		if req.Email != "" {
			updates["email"] = req.Email
		}
		if req.Phone != "" {
			updates["phone"] = req.Phone
		}

		user, err := s.repo.Update(txCtx, id, updates)
		if err != nil {
			return appErrors.ParseDatabaseError(err)
		}

		u := user.toResponse()
		res = &u
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *userService) Delete(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return appErrors.BadRequest("Invalid user ID format")
	}

	return s.runInTx(ctx, func(txCtx context.Context) error {
		err := s.repo.Delete(txCtx, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return appErrors.NotFound("User not found")
			}
			return appErrors.ParseDatabaseError(err)
		}
		return nil
	})
}

func (s *userService) HardDelete(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return appErrors.BadRequest("Invalid user ID format")
	}

	return s.runInTx(ctx, func(txCtx context.Context) error {
		err := s.repo.HardDelete(txCtx, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return appErrors.NotFound("User not found")
			}
			return appErrors.ParseDatabaseError(err)
		}
		return nil
	})
}

func (u *User) toResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Phone:     u.Phone.String,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
