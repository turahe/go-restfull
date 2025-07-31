package controllers

import (
	"strconv"

	"github.com/turahe/go-restfull/internal/application/ports"
	"github.com/turahe/go-restfull/internal/interfaces/http/responses"

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

// AssignRoleToUser godoc
//
//	@Summary		Assign role to user
//	@Description	Assign a specific role to a user
//	@Tags			Authentication & Authorization
//	@Accept			json
//	@Produce		json
//	@Param			user_id	path		string						true	"User ID"	format(uuid)
//	@Param			role_id	path		string						true	"Role ID"	format(uuid)
//	@Success		200		{object}	responses.CommonResponse	"Role assigned to user successfully"
//	@Failure		400		{object}	responses.CommonResponse	"Bad request - Invalid user ID or role ID"
//	@Router			/api/v1/users/{user_id}/roles/{role_id} [post]
//	@Security		BearerAuth
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

// RemoveRoleFromUser godoc
//
//	@Summary		Remove role from user
//	@Description	Remove a specific role from a user
//	@Tags			Authentication & Authorization
//	@Accept			json
//	@Produce		json
//	@Param			user_id	path		string						true	"User ID"	format(uuid)
//	@Param			role_id	path		string						true	"Role ID"	format(uuid)
//	@Success		200		{object}	responses.CommonResponse	"Role removed from user successfully"
//	@Failure		400		{object}	responses.CommonResponse	"Bad request - Invalid user ID or role ID"
//	@Router			/api/v1/users/{user_id}/roles/{role_id} [delete]
//	@Security		BearerAuth
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

// GetUserRoles godoc
//
//	@Summary		Get user roles
//	@Description	Retrieve all roles assigned to a specific user
//	@Tags			Authentication & Authorization
//	@Accept			json
//	@Produce		json
//	@Param			user_id	path		string											true	"User ID"	format(uuid)
//	@Success		200		{object}	responses.CommonResponse{data=[]entities.Role}	"User roles retrieved successfully"
//	@Failure		400		{object}	responses.CommonResponse						"Bad request - Invalid user ID"
//	@Failure		500		{object}	responses.CommonResponse						"Internal server error"
//	@Router			/api/v1/users/{user_id}/roles [get]
//	@Security		BearerAuth
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

// GetRoleUsers godoc
//
//	@Summary		Get role users
//	@Description	Retrieve all users assigned to a specific role with pagination
//	@Tags			Authentication & Authorization
//	@Accept			json
//	@Produce		json
//	@Param			role_id	path		string											true	"Role ID"									format(uuid)
//	@Param			limit	query		int												false	"Number of results to return (default: 10)"	default(10)
//	@Param			offset	query		int												false	"Number of results to skip (default: 0)"	default(0)
//	@Success		200		{object}	responses.CommonResponse{data=[]entities.User}	"Role users retrieved successfully"
//	@Failure		400		{object}	responses.CommonResponse						"Bad request - Invalid role ID"
//	@Failure		500		{object}	responses.CommonResponse						"Internal server error"
//	@Router			/api/v1/roles/{role_id}/users [get]
//	@Security		BearerAuth
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

// HasRole godoc
//
//	@Summary		Check if user has role
//	@Description	Check if a specific user has a specific role
//	@Tags			Authentication & Authorization
//	@Accept			json
//	@Produce		json
//	@Param			user_id	path		string													true	"User ID"	format(uuid)
//	@Param			role_id	path		string													true	"Role ID"	format(uuid)
//	@Success		200		{object}	responses.CommonResponse{data=object{has_role=bool}}	"Role check completed successfully"
//	@Failure		400		{object}	responses.CommonResponse								"Bad request - Invalid user ID or role ID"
//	@Failure		500		{object}	responses.CommonResponse								"Internal server error"
//	@Router			/api/v1/users/{user_id}/roles/{role_id}/check [get]
//	@Security		BearerAuth
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

// GetUserRoleCount godoc
//
//	@Summary		Get user count for role
//	@Description	Get the total number of users assigned to a specific role
//	@Tags			Authentication & Authorization
//	@Accept			json
//	@Produce		json
//	@Param			role_id	path		string												true	"Role ID"	format(uuid)
//	@Success		200		{object}	responses.CommonResponse{data=object{count=int}}	"User count for role retrieved successfully"
//	@Failure		400		{object}	responses.CommonResponse							"Bad request - Invalid role ID"
//	@Failure		500		{object}	responses.CommonResponse							"Internal server error"
//	@Router			/api/v1/roles/{role_id}/users/count [get]
//	@Security		BearerAuth
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
