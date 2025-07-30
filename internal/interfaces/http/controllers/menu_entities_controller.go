package controllers

import (
	"strconv"
	"github.com/turahe/go-restfull/internal/application/ports"
	"github.com/turahe/go-restfull/internal/interfaces/http/responses"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type MenuEntitiesController struct {
	menuRoleService ports.MenuEntitiesService
}

func NewMenuRoleController(menuRoleService ports.MenuEntitiesService) *MenuEntitiesController {
	return &MenuEntitiesController{
		menuRoleService: menuRoleService,
	}
}

// AssignRoleToMenu handles POST /v1/menus/:menu_id/roles/:role_id requests
func (c *MenuEntitiesController) AssignRoleToMenu(ctx *fiber.Ctx) error {
	menuID, err := uuid.Parse(ctx.Params("menu_id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid menu ID",
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

	err = c.menuRoleService.AssignRoleToMenu(ctx.Context(), menuID, roleID)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: err.Error(),
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(responses.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Role assigned to menu successfully",
		Data:            map[string]interface{}{},
	})
}

// RemoveRoleFromMenu handles DELETE /v1/menus/:menu_id/roles/:role_id requests
func (c *MenuEntitiesController) RemoveRoleFromMenu(ctx *fiber.Ctx) error {
	menuID, err := uuid.Parse(ctx.Params("menu_id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid menu ID",
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

	err = c.menuRoleService.RemoveRoleFromMenu(ctx.Context(), menuID, roleID)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: err.Error(),
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(responses.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Role removed from menu successfully",
		Data:            map[string]interface{}{},
	})
}

// GetMenuRoles handles GET /v1/menus/:menu_id/roles requests
func (c *MenuEntitiesController) GetMenuRoles(ctx *fiber.Ctx) error {
	menuID, err := uuid.Parse(ctx.Params("menu_id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid menu ID",
			Data:            map[string]interface{}{},
		})
	}

	roles, err := c.menuRoleService.GetMenuRoles(ctx.Context(), menuID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to retrieve menu roles",
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(responses.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Menu roles retrieved successfully",
		Data:            roles,
	})
}

// GetRoleMenus handles GET /v1/roles/:role_id/menus requests
func (c *MenuEntitiesController) GetRoleMenus(ctx *fiber.Ctx) error {
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

	menus, err := c.menuRoleService.GetRoleMenus(ctx.Context(), roleID, limit, offset)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to retrieve role menus",
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(responses.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Role menus retrieved successfully",
		Data:            menus,
	})
}

// HasRole handles GET /v1/menus/:menu_id/roles/:role_id/check requests
func (c *MenuEntitiesController) HasRole(ctx *fiber.Ctx) error {
	menuID, err := uuid.Parse(ctx.Params("menu_id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid menu ID",
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

	hasRole, err := c.menuRoleService.HasRole(ctx.Context(), menuID, roleID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to check role assignment",
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(responses.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Role check completed successfully",
		Data:            map[string]interface{}{"has_role": hasRole},
	})
}

// GetMenuRoleCount handles GET /v1/roles/:role_id/menus/count requests
func (c *MenuEntitiesController) GetMenuRoleCount(ctx *fiber.Ctx) error {
	roleID, err := uuid.Parse(ctx.Params("role_id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid role ID",
			Data:            map[string]interface{}{},
		})
	}

	count, err := c.menuRoleService.GetMenuRoleCount(ctx.Context(), roleID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to get menu count",
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(responses.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Menu count retrieved successfully",
		Data:            map[string]interface{}{"count": count},
	})
}
