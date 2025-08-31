package controllers

import (
	"strconv"

	"github.com/turahe/go-restfull/internal/application/ports"
	"github.com/turahe/go-restfull/internal/interfaces/http/responses"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/turahe/go-restfull/pkg/logger"
	"go.uber.org/zap")

type UserRoleController struct {
	userRoleService ports.UserRoleService
}

func NewUserRoleController(userRoleService ports.UserRoleService) *UserRoleController {
	return &UserRoleController{
		userRoleService: userRoleService,
	}
}

// AssignRoleToUser godoc
// @Summary Assign role to user
// @Description Assign a specific role to a user
// @Tags Authentication & Authorization
// @Accept json
// @Produce json
// @Param user_id path string true "User ID" format(uuid)
// @Param role_id path string true "Role ID" format(uuid)
// @Success 200 {object} responses.UserRoleResourceResponse "Role assigned to user successfully"
// @Failure 400 {object} responses.ErrorResponse "Bad request - Invalid user ID or role ID"
// @Router /api/v1/users/{user_id}/roles/{role_id} [post]
// @Security BearerAuth
func (c *UserRoleController) AssignRoleToUser(ctx *fiber.Ctx) error {
	userID, err := uuid.Parse(ctx.Params("user_id"))
	if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid user ID",
		})
	}

	roleID, err := uuid.Parse(ctx.Params("role_id"))
	if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid role ID",
		})
	}

	err = c.userRoleService.AssignRoleToUser(ctx.Context(), userID, roleID)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.JSON(responses.NewUserRoleResourceResponse(userID.String(), roleID.String(), nil, nil))
}

// RemoveRoleFromUser godoc
// @Summary Remove role from user
// @Description Remove a specific role from a user
// @Tags Authentication & Authorization
// @Accept json
// @Produce json
// @Param user_id path string true "User ID" format(uuid)
// @Param role_id path string true "Role ID" format(uuid)
// @Success 200 {object} responses.UserRoleResourceResponse "Role removed from user successfully"
// @Failure 400 {object} responses.ErrorResponse "Bad request - Invalid user ID or role ID"
// @Router /api/v1/users/{user_id}/roles/{role_id} [delete]
// @Security BearerAuth
func (c *UserRoleController) RemoveRoleFromUser(ctx *fiber.Ctx) error {
	userID, err := uuid.Parse(ctx.Params("user_id"))
	if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid user ID",
		})
	}

	roleID, err := uuid.Parse(ctx.Params("role_id"))
	if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid role ID",
		})
	}

	err = c.userRoleService.RemoveRoleFromUser(ctx.Context(), userID, roleID)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.JSON(responses.NewUserRoleResourceResponse(userID.String(), roleID.String(), nil, nil))
}

// GetUserRoles godoc
// @Summary Get user roles
// @Description Retrieve all roles assigned to a specific user
// @Tags Authentication & Authorization
// @Accept json
// @Produce json
// @Param user_id path string true "User ID" format(uuid)
// @Success 200 {object} responses.UserRoleCollectionResponse "User roles retrieved successfully"
// @Failure 400 {object} responses.ErrorResponse "Bad request - Invalid user ID"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /api/v1/users/{user_id}/roles [get]
// @Security BearerAuth
func (c *UserRoleController) GetUserRoles(ctx *fiber.Ctx) error {
	userID, err := uuid.Parse(ctx.Params("user_id"))
	if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid user ID",
		})
	}

	roles, err := c.userRoleService.GetUserRoles(ctx.Context(), userID)
	if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Failed to retrieve user roles",
		})
	}

	// Convert roles to UserRoleResource format
	userRoleResources := make([]responses.UserRoleResource, len(roles))
	for i, role := range roles {
		userRoleResources[i] = responses.NewUserRoleResource(userID.String(), role.ID.String(), nil, role)
	}

	return ctx.JSON(responses.NewUserRoleCollectionResponse(userRoleResources))
}

// GetRoleUsers godoc
// @Summary Get role users
// @Description Retrieve all users assigned to a specific role with pagination
// @Tags Authentication & Authorization
// @Accept json
// @Produce json
// @Param role_id path string true "Role ID" format(uuid)
// @Param limit query int false "Number of results to return (default: 10)" default(10)
// @Param offset query int false "Number of results to skip (default: 0)" default(0)
// @Success 200 {object} responses.RoleUserCollectionResponse "Role users retrieved successfully"
// @Failure 400 {object} responses.ErrorResponse "Bad request - Invalid role ID"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /api/v1/roles/{role_id}/users [get]
// @Security BearerAuth
func (c *UserRoleController) GetRoleUsers(ctx *fiber.Ctx) error {
	roleID, err := uuid.Parse(ctx.Params("role_id"))
	if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid role ID",
		})
	}

	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))
	offset, _ := strconv.Atoi(ctx.Query("offset", "0"))

	users, err := c.userRoleService.GetRoleUsers(ctx.Context(), roleID, limit, offset)
	if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Failed to retrieve role users",
		})
	}

	// Convert users to RoleUserResource format
	roleUserResources := make([]responses.RoleUserResource, len(users))
	for i, user := range users {
		roleUserResources[i] = responses.NewRoleUserResource(roleID.String(), user.ID.String(), nil, user)
	}

	return ctx.JSON(responses.NewRoleUserCollectionResponse(roleUserResources))
}

// HasRole godoc
// @Summary Check if user has role
// @Description Check if a specific user has a specific role
// @Tags Authentication & Authorization
// @Accept json
// @Produce json
// @Param user_id path string true "User ID" format(uuid)
// @Param role_id path string true "Role ID" format(uuid)
// @Success 200 {object} responses.ErrorResponse "Role check completed successfully"
// @Failure 400 {object} responses.ErrorResponse "Bad request - Invalid user ID or role ID"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /api/v1/users/{user_id}/roles/{role_id}/check [get]
// @Security BearerAuth
func (c *UserRoleController) HasRole(ctx *fiber.Ctx) error {
	userID, err := uuid.Parse(ctx.Params("user_id"))
	if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid user ID",
		})
	}

	roleID, err := uuid.Parse(ctx.Params("role_id"))
	if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid role ID",
		})
	}

	hasRole, err := c.userRoleService.HasRole(ctx.Context(), userID, roleID)
	if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Failed to check user role",
		})
	}

	return ctx.JSON(responses.SuccessResponse{
		Status:  "success",
		Message: "Role check completed successfully",
		Data:    map[string]interface{}{"has_role": hasRole},
	})
}

// GetUserRoleCount godoc
// @Summary Get user count for role
// @Description Get the total number of users assigned to a specific role
// @Tags Authentication & Authorization
// @Accept json
// @Produce json
// @Param role_id path string true "Role ID" format(uuid)
// @Success 200 {object} responses.SuccessResponse "User count for role retrieved successfully"
// @Failure 400 {object} responses.ErrorResponse "Bad request - Invalid role ID"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /api/v1/roles/{role_id}/users/count [get]
// @Security BearerAuth
func (c *UserRoleController) GetUserRoleCount(ctx *fiber.Ctx) error {
	roleID, err := uuid.Parse(ctx.Params("role_id"))
	if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid role ID",
		})
	}

	count, err := c.userRoleService.GetUserRoleCount(ctx.Context(), roleID)
	if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Failed to get user count for role",
		})
	}

	return ctx.JSON(responses.SuccessResponse{
		Status:  "success",
		Message: "User count for role retrieved successfully",
		Data:    map[string]interface{}{"count": count},
	})
}
