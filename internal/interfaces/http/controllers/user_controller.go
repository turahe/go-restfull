package controllers

import (
	"net/http"
	"strconv"

	"webapi/internal/application/ports"
	"webapi/internal/domain/entities"
	"webapi/internal/interfaces/http/requests"
	"webapi/internal/interfaces/http/responses"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// UserController handles HTTP requests for user operations
// @title User Management API
// @version 1.0
// @description This is a user management API with authentication and authorization
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email support@example.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8000
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
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
// @Summary Create a new user
// @Description Create a new user account with the provided information
// @Tags users
// @Accept json
// @Produce json
// @Param user body requests.CreateUserRequest true "User creation request"
// @Success 201 {object} responses.SuccessResponse{data=responses.UserResponse} "User created successfully"
// @Failure 400 {object} responses.ErrorResponse "Bad request - Invalid input data"
// @Failure 409 {object} responses.ErrorResponse "Conflict - User already exists"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /users [post]
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
		Data:   responses.NewUserResponse(user),
	})
}

// GetUserByID handles GET /users/:id
// @Summary Get user by ID
// @Description Retrieve a user by their unique identifier
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID" format(uuid)
// @Success 200 {object} responses.SuccessResponse{data=responses.UserResponse} "User found"
// @Failure 400 {object} responses.ErrorResponse "Bad request - Invalid user ID"
// @Failure 404 {object} responses.ErrorResponse "Not found - User does not exist"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /users/{id} [get]
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
		Data:   responses.NewUserResponse(user),
	})
}

// GetUsers handles GET /users
// @Summary Get all users
// @Description Retrieve a paginated list of users with optional search and filtering
// @Tags users
// @Accept json
// @Produce json
// @Param limit query int false "Number of users to return (default: 10, max: 100)" default(10) minimum(1) maximum(100)
// @Param offset query int false "Number of users to skip (default: 0)" default(0) minimum(0)
// @Param query query string false "Search query to filter users by username, email, or phone"
// @Success 200 {object} responses.SuccessResponse{data=[]responses.UserResponse} "List of users"
// @Failure 400 {object} responses.ErrorResponse "Bad request - Invalid parameters"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /users [get]
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

	var users []*entities.User
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

	// Convert to response DTOs
	userResponses := make([]responses.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = *responses.NewUserResponse(user)
	}

	return ctx.JSON(responses.SuccessResponse{
		Status: "success",
		Data:   userResponses,
	})
}

// UpdateUser handles PUT /users/:id
// @Summary Update user
// @Description Update an existing user's information
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID" format(uuid)
// @Param user body requests.UpdateUserRequest true "User update request"
// @Success 200 {object} responses.SuccessResponse{data=responses.UserResponse} "User updated successfully"
// @Failure 400 {object} responses.ErrorResponse "Bad request - Invalid input data"
// @Failure 404 {object} responses.ErrorResponse "Not found - User does not exist"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /users/{id} [put]
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
		Data:   responses.NewUserResponse(user),
	})
}

// DeleteUser handles DELETE /users/:id
// @Summary Delete user
// @Description Delete a user account (soft delete)
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID" format(uuid)
// @Success 200 {object} responses.SuccessResponse "User deleted successfully"
// @Failure 400 {object} responses.ErrorResponse "Bad request - Invalid user ID"
// @Failure 404 {object} responses.ErrorResponse "Not found - User does not exist"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /users/{id} [delete]
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
// @Summary Change user password
// @Description Change a user's password with old password verification
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID" format(uuid)
// @Param password body requests.ChangePasswordRequest true "Password change request"
// @Success 200 {object} responses.SuccessResponse "Password changed successfully"
// @Failure 400 {object} responses.ErrorResponse "Bad request - Invalid input data"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized - Invalid old password"
// @Failure 404 {object} responses.ErrorResponse "Not found - User does not exist"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /users/{id}/password [put]
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
