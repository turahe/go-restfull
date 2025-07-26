package controllers

import (
	"net/http"
	"strconv"

	"webapi/internal/application/ports"
	"webapi/internal/interfaces/http/requests"
	"webapi/internal/interfaces/http/responses"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// UserController handles HTTP requests for user operations
type UserController struct {
	userService ports.UserService
}

// NewUserController creates a new user controller
func NewUserController(userService ports.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

// CreateUser handles POST /users
func (c *UserController) CreateUser(ctx *fiber.Ctx) error {
	var req requests.CreateUserRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid request body",
		})
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	// Create user
	user, err := c.userService.CreateUser(ctx.Context(), req.Username, req.Email, req.Phone, req.Password)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusCreated).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   user,
	})
}

// GetUserByID handles GET /users/:id
func (c *UserController) GetUserByID(ctx *fiber.Ctx) error {
	idParam := ctx.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid user ID",
		})
	}

	user, err := c.userService.GetUserByID(ctx.Context(), id)
	if err != nil {
		return ctx.Status(http.StatusNotFound).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "User not found",
		})
	}

	return ctx.JSON(responses.SuccessResponse{
		Status: "success",
		Data:   user,
	})
}

// GetUsers handles GET /users
func (c *UserController) GetUsers(ctx *fiber.Ctx) error {
	limitStr := ctx.Query("limit", "10")
	offsetStr := ctx.Query("offset", "0")
	query := ctx.Query("query", "")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	var users interface{}
	var err2 error

	if query != "" {
		users, err2 = c.userService.SearchUsers(ctx.Context(), query, limit, offset)
	} else {
		users, err2 = c.userService.GetAllUsers(ctx.Context(), limit, offset)
	}

	if err2 != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err2.Error(),
		})
	}

	return ctx.JSON(responses.SuccessResponse{
		Status: "success",
		Data:   users,
	})
}

// UpdateUser handles PUT /users/:id
func (c *UserController) UpdateUser(ctx *fiber.Ctx) error {
	idParam := ctx.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid user ID",
		})
	}

	var req requests.UpdateUserRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid request body",
		})
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	// Update user
	user, err := c.userService.UpdateUser(ctx.Context(), id, req.Username, req.Email, req.Phone)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.JSON(responses.SuccessResponse{
		Status: "success",
		Data:   user,
	})
}

// DeleteUser handles DELETE /users/:id
func (c *UserController) DeleteUser(ctx *fiber.Ctx) error {
	idParam := ctx.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid user ID",
		})
	}

	err = c.userService.DeleteUser(ctx.Context(), id)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.JSON(responses.SuccessResponse{
		Status:  "success",
		Message: "User deleted successfully",
	})
}

// ChangePassword handles PUT /users/:id/password
func (c *UserController) ChangePassword(ctx *fiber.Ctx) error {
	idParam := ctx.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid user ID",
		})
	}

	var req requests.ChangePasswordRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid request body",
		})
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	// Change password
	err = c.userService.ChangePassword(ctx.Context(), id, req.OldPassword, req.NewPassword)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.JSON(responses.SuccessResponse{
		Status:  "success",
		Message: "Password changed successfully",
	})
}

// AuthenticateUser handles POST /auth/login
func (c *UserController) AuthenticateUser(ctx *fiber.Ctx) error {
	var req requests.LoginRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid request body",
		})
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	// Authenticate user
	user, err := c.userService.AuthenticateUser(ctx.Context(), req.Username, req.Password)
	if err != nil {
		return ctx.Status(http.StatusUnauthorized).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid credentials",
		})
	}

	return ctx.JSON(responses.SuccessResponse{
		Status: "success",
		Data:   user,
	})
}
