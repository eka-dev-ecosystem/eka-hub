package usermodule

import (
	"golang/pkg/response"
	"golang/pkg/validation"

	"github.com/gofiber/fiber/v2"
)

// UserHandler exposes the user CRUD endpoints.
type UserHandler struct {
	service UserService
}

// NewUserHandler builds the user handler.
func NewUserHandler(service UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) GetUsers() fiber.Handler {
	return func(c *fiber.Ctx) error {
		users, err := h.service.GetAll(c.UserContext())
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, users, "Users retrieved successfully")
	}
}

func (h *UserHandler) GetUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, err := h.service.GetByID(c.UserContext(), c.Params("id"))
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, user, "User retrieved successfully")
	}
}

func (h *UserHandler) CreateUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req CreateUserRequest
		if err := validation.Default.BindAndValidate(c, &req); err != nil {
			return err
		}

		user, err := h.service.Create(c.UserContext(), &req)
		if err != nil {
			return err
		}
		return response.CreatedResponse(c, user, "User created successfully")
	}
}

func (h *UserHandler) UpdateUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req UpdateUserRequest
		if err := validation.Default.BindAndValidate(c, &req); err != nil {
			return err
		}

		user, err := h.service.Update(c.UserContext(), c.Params("id"), &req)
		if err != nil {
			return err
		}
		return response.SuccessResponse(c, user, "User updated successfully")
	}
}

func (h *UserHandler) DeleteUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := h.service.Delete(c.UserContext(), c.Params("id")); err != nil {
			return err
		}
		return response.SuccessMessageResponse(c, "User deleted successfully")
	}
}

func (h *UserHandler) HardDeleteUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := h.service.HardDelete(c.UserContext(), c.Params("id")); err != nil {
			return err
		}
		return response.SuccessMessageResponse(c, "User permanently deleted")
	}
}
