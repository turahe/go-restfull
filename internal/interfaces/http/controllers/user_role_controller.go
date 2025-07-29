package controllers

import (
	"strconv"
	"webapi/internal/application/ports"
	"webapi/internal/interfaces/http/responses"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type UserRoleController struct {
	userRoleService ports.UserRoleService
}

func NewUserRoleController(userRoleService ports.UserRoleService) *UserRoleController {
	return &UserRoleController{
		userRoleService: userRoleService,
	}
}

// AssignRoleToUser handles POST /v1/users/:user_id/roles/:role_id requests
func (c *UserRoleController) AssignRoleToUser(ctx *fiber.Ctx) error {
	userID, err := uuid.Parse(ctx.Params("user_id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid user ID",
			Data:            map[string]interface{}{},
		})
	}

	roleID, err := uuid.Parse(ctx.Params("role_id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid role ID",
			Data:            map[string]interface{}{},
		})
	}

	err = c.userRoleService.AssignRoleToUser(ctx.Context(), userID, roleID)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: err.Error(),
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(responses.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Role assigned to user successfully",
		Data:            map[string]interface{}{},
	})
}

// RemoveRoleFromUser handles DELETE /v1/users/:user_id/roles/:role_id requests
func (c *UserRoleController) RemoveRoleFromUser(ctx *fiber.Ctx) error {
	userID, err := uuid.Parse(ctx.Params("user_id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid user ID",
			Data:            map[string]interface{}{},
		})
	}

	roleID, err := uuid.Parse(ctx.Params("role_id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid role ID",
			Data:            map[string]interface{}{},
		})
	}

	err = c.userRoleService.RemoveRoleFromUser(ctx.Context(), userID, roleID)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: err.Error(),
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(responses.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Role removed from user successfully",
		Data:            map[string]interface{}{},
	})
}

// GetUserRoles handles GET /v1/users/:user_id/roles requests
func (c *UserRoleController) GetUserRoles(ctx *fiber.Ctx) error {
	userID, err := uuid.Parse(ctx.Params("user_id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid user ID",
			Data:            map[string]interface{}{},
		})
	}

	roles, err := c.userRoleService.GetUserRoles(ctx.Context(), userID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to retrieve user roles",
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(responses.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "User roles retrieved successfully",
		Data:            roles,
	})
}

// GetRoleUsers handles GET /v1/roles/:role_id/users requests
func (c *UserRoleController) GetRoleUsers(ctx *fiber.Ctx) error {
	roleID, err := uuid.Parse(ctx.Params("role_id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid role ID",
			Data:            map[string]interface{}{},
		})
	}

	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))
	offset, _ := strconv.Atoi(ctx.Query("offset", "0"))

	users, err := c.userRoleService.GetRoleUsers(ctx.Context(), roleID, limit, offset)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to retrieve role users",
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(responses.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Role users retrieved successfully",
		Data:            users,
	})
}

// HasRole handles GET /v1/users/:user_id/roles/:role_id/check requests
func (c *UserRoleController) HasRole(ctx *fiber.Ctx) error {
	userID, err := uuid.Parse(ctx.Params("user_id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid user ID",
			Data:            map[string]interface{}{},
		})
	}

	roleID, err := uuid.Parse(ctx.Params("role_id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid role ID",
			Data:            map[string]interface{}{},
		})
	}

	hasRole, err := c.userRoleService.HasRole(ctx.Context(), userID, roleID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to check user role",
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(responses.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Role check completed successfully",
		Data:            map[string]interface{}{"has_role": hasRole},
	})
}

// GetUserRoleCount handles GET /v1/roles/:role_id/users/count requests
func (c *UserRoleController) GetUserRoleCount(ctx *fiber.Ctx) error {
	roleID, err := uuid.Parse(ctx.Params("role_id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid role ID",
			Data:            map[string]interface{}{},
		})
	}

	count, err := c.userRoleService.GetUserRoleCount(ctx.Context(), roleID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to get user count for role",
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(responses.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "User count for role retrieved successfully",
		Data:            map[string]interface{}{"count": count},
	})
}
