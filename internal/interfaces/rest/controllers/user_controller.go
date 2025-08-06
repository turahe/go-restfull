package controllers

import (
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/turahe/go-restfull/internal/application/commands"
	"github.com/turahe/go-restfull/internal/application/handlers"
	"github.com/turahe/go-restfull/internal/application/queries"
	"github.com/turahe/go-restfull/internal/interfaces/rest/dto"
	"github.com/turahe/go-restfull/internal/interfaces/rest/middleware"
	"github.com/turahe/go-restfull/internal/shared/errors"
)

// UserController handles HTTP requests for user operations
type UserController struct {
	commandHandlers *handlers.UserCommandHandlers
	queryHandlers   *handlers.UserQueryHandlers
}

// NewUserController creates a new user controller
func NewUserController(
	commandHandlers *handlers.UserCommandHandlers,
	queryHandlers *handlers.UserQueryHandlers,
) *UserController {
	return &UserController{
		commandHandlers: commandHandlers,
		queryHandlers:   queryHandlers,
	}
}

// CreateUser creates a new user
// @Summary Create a new user
// @Description Create a new user with the provided information
// @Tags users
// @Accept json
// @Produce json
// @Param user body dto.CreateUserRequest true "User creation data"
// @Success 201 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users [post]
func (c *UserController) CreateUser(ctx *fiber.Ctx) error {
	var req dto.CreateUserRequest
	if err := ctx.BodyParser(&req); err != nil {
		return c.handleError(ctx, errors.NewDomainError("INVALID_REQUEST", "Invalid request body"))
	}

	if err := middleware.ValidateStruct(&req); err != nil {
		return c.handleError(ctx, err)
	}

	cmd := commands.CreateUserCommand{
		Username: req.Username,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: req.Password,
	}

	user, err := c.commandHandlers.CreateUser.Handle(ctx.Context(), cmd)
	if err != nil {
		return c.handleError(ctx, err)
	}

	response := dto.NewUserResponse(user)
	return ctx.Status(http.StatusCreated).JSON(response)
}

// GetUser gets a user by ID
// @Summary Get user by ID
// @Description Get a user by their unique identifier
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/{id} [get]
func (c *UserController) GetUser(ctx *fiber.Ctx) error {
	userID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return c.handleError(ctx, errors.NewDomainError("INVALID_USER_ID", "Invalid user ID format"))
	}

	query := queries.GetUserByIDQuery{
		UserID: userID,
	}

	user, err := c.queryHandlers.GetUserByID.Handle(ctx.Context(), query)
	if err != nil {
		return c.handleError(ctx, err)
	}

	response := dto.NewUserResponse(user)
	return ctx.JSON(response)
}

// UpdateUser updates a user
// @Summary Update user
// @Description Update user information
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body dto.UpdateUserRequest true "User update data"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/{id} [put]
func (c *UserController) UpdateUser(ctx *fiber.Ctx) error {
	userID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return c.handleError(ctx, errors.NewDomainError("INVALID_USER_ID", "Invalid user ID format"))
	}

	var req dto.UpdateUserRequest
	if err := ctx.BodyParser(&req); err != nil {
		return c.handleError(ctx, errors.NewDomainError("INVALID_REQUEST", "Invalid request body"))
	}

	if err := middleware.ValidateStruct(&req); err != nil {
		return c.handleError(ctx, err)
	}

	cmd := commands.UpdateUserCommand{
		UserID:   userID,
		Username: req.Username,
		Email:    req.Email,
		Phone:    req.Phone,
	}

	err = c.commandHandlers.UpdateUser.Handle(ctx.Context(), cmd)
	if err != nil {
		return c.handleError(ctx, err)
	}

	// Get updated user
	query := queries.GetUserByIDQuery{UserID: userID}
	user, err := c.queryHandlers.GetUserByID.Handle(ctx.Context(), query)
	if err != nil {
		return c.handleError(ctx, err)
	}

	response := dto.NewUserResponse(user)
	return ctx.JSON(response)
}

// DeleteUser deletes a user
// @Summary Delete user
// @Description Soft delete a user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 204
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/{id} [delete]
func (c *UserController) DeleteUser(ctx *fiber.Ctx) error {
	userID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return c.handleError(ctx, errors.NewDomainError("INVALID_USER_ID", "Invalid user ID format"))
	}

	cmd := commands.DeleteUserCommand{
		UserID: userID,
	}

	err = c.commandHandlers.DeleteUser.Handle(ctx.Context(), cmd)
	if err != nil {
		return c.handleError(ctx, err)
	}

	return ctx.SendStatus(http.StatusNoContent)
}

// ListUsers lists users with pagination and filters
// @Summary List users
// @Description Get a paginated list of users with optional filters
// @Tags users
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param search query string false "Search term"
// @Param role_id query string false "Role ID filter"
// @Param sort_by query string false "Sort field" Enums(username, email, created_at, updated_at)
// @Param sort_dir query string false "Sort direction" Enums(asc, desc)
// @Success 200 {object} dto.PaginatedUsersResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users [get]
func (c *UserController) ListUsers(ctx *fiber.Ctx) error {
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.Query("page_size", "10"))
	
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	query := queries.ListUsersQuery{
		Page:     page,
		PageSize: pageSize,
	}

	if search := ctx.Query("search"); search != "" {
		query.Search = &search
	}

	if roleIDStr := ctx.Query("role_id"); roleIDStr != "" {
		if roleID, err := uuid.Parse(roleIDStr); err == nil {
			query.RoleID = &roleID
		}
	}

	if sortBy := ctx.Query("sort_by"); sortBy != "" {
		query.SortBy = &sortBy
	}

	if sortDir := ctx.Query("sort_dir"); sortDir != "" {
		query.SortDir = &sortDir
	}

	result, err := c.queryHandlers.ListUsers.Handle(ctx.Context(), query)
	if err != nil {
		return c.handleError(ctx, err)
	}

	response := dto.NewPaginatedUsersResponse(result)
	return ctx.JSON(response)
}

// SearchUsers searches users
// @Summary Search users
// @Description Search users by query string
// @Tags users
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} dto.PaginatedUsersResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/search [get]
func (c *UserController) SearchUsers(ctx *fiber.Ctx) error {
	searchQuery := ctx.Query("q")
	if searchQuery == "" {
		return c.handleError(ctx, errors.NewDomainError("MISSING_QUERY", "Search query is required"))
	}

	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.Query("page_size", "10"))
	
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	query := queries.SearchUsersQuery{
		Query:    searchQuery,
		Page:     page,
		PageSize: pageSize,
	}

	result, err := c.queryHandlers.SearchUsers.Handle(ctx.Context(), query)
	if err != nil {
		return c.handleError(ctx, err)
	}

	response := dto.NewPaginatedUsersResponse(result)
	return ctx.JSON(response)
}

// ChangePassword changes user password
// @Summary Change user password
// @Description Change a user's password
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param password body dto.ChangePasswordRequest true "Password change data"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/{id}/password [put]
func (c *UserController) ChangePassword(ctx *fiber.Ctx) error {
	userID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return c.handleError(ctx, errors.NewDomainError("INVALID_USER_ID", "Invalid user ID format"))
	}

	var req dto.ChangePasswordRequest
	if err := ctx.BodyParser(&req); err != nil {
		return c.handleError(ctx, errors.NewDomainError("INVALID_REQUEST", "Invalid request body"))
	}

	if err := middleware.ValidateStruct(&req); err != nil {
		return c.handleError(ctx, err)
	}

	cmd := commands.ChangePasswordCommand{
		UserID:      userID,
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	}

	err = c.commandHandlers.ChangePassword.Handle(ctx.Context(), cmd)
	if err != nil {
		return c.handleError(ctx, err)
	}

	return ctx.JSON(dto.SuccessResponse{
		Message: "Password changed successfully",
	})
}

// handleError handles errors and returns appropriate HTTP responses
func (c *UserController) handleError(ctx *fiber.Ctx, err error) error {
	if domainErr := errors.GetDomainError(err); domainErr != nil {
		status := c.getHTTPStatusFromDomainError(domainErr)
		return ctx.Status(status).JSON(dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    domainErr.Code,
				Message: domainErr.Message,
				Details: domainErr.Details,
			},
		})
	}

	// Generic error
	return ctx.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{
		Error: dto.ErrorDetail{
			Code:    "INTERNAL_ERROR",
			Message: "An internal error occurred",
		},
	})
}

// getHTTPStatusFromDomainError maps domain error codes to HTTP status codes
func (c *UserController) getHTTPStatusFromDomainError(err *errors.DomainError) int {
	switch err.Code {
	case errors.ValidationErrorCode, errors.InvalidEmailErrorCode, 
		 errors.InvalidPhoneErrorCode, errors.InvalidPasswordErrorCode,
		 errors.RequiredFieldErrorCode:
		return http.StatusBadRequest
	case errors.EmailAlreadyExistsCode, errors.UsernameAlreadyExistsCode,
		 errors.PhoneAlreadyExistsCode, errors.EmailAlreadyVerifiedCode,
		 errors.PhoneAlreadyVerifiedCode, errors.RoleAlreadyAssignedCode:
		return http.StatusConflict
	case errors.UserNotFoundCode, errors.RoleNotFoundCode, errors.NotFoundErrorCode:
		return http.StatusNotFound
	case errors.UnauthorizedErrorCode, errors.InvalidCredentialsCode:
		return http.StatusUnauthorized
	case errors.ForbiddenErrorCode:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}