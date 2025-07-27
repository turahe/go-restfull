package controllers

import (
	"strconv"
	"webapi/internal/application/ports"
	"webapi/internal/http/response"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type MenuRoleController struct {
	menuRoleService ports.MenuRoleService
}

func NewMenuRoleController(menuRoleService ports.MenuRoleService) *MenuRoleController {
	return &MenuRoleController{
		menuRoleService: menuRoleService,
	}
}

// AssignRoleToMenu handles POST /v1/menus/:menu_id/roles/:role_id requests
func (c *MenuRoleController) AssignRoleToMenu(ctx *fiber.Ctx) error {
	menuID, err := uuid.Parse(ctx.Params("menu_id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid menu ID",
			Data:            map[string]interface{}{},
		})
	}

	roleID, err := uuid.Parse(ctx.Params("role_id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid role ID",
			Data:            map[string]interface{}{},
		})
	}

	err = c.menuRoleService.AssignRoleToMenu(ctx.Context(), menuID, roleID)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: err.Error(),
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(response.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Role assigned to menu successfully",
		Data:            map[string]interface{}{},
	})
}

// RemoveRoleFromMenu handles DELETE /v1/menus/:menu_id/roles/:role_id requests
func (c *MenuRoleController) RemoveRoleFromMenu(ctx *fiber.Ctx) error {
	menuID, err := uuid.Parse(ctx.Params("menu_id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid menu ID",
			Data:            map[string]interface{}{},
		})
	}

	roleID, err := uuid.Parse(ctx.Params("role_id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid role ID",
			Data:            map[string]interface{}{},
		})
	}

	err = c.menuRoleService.RemoveRoleFromMenu(ctx.Context(), menuID, roleID)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: err.Error(),
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(response.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Role removed from menu successfully",
		Data:            map[string]interface{}{},
	})
}

// GetMenuRoles handles GET /v1/menus/:menu_id/roles requests
func (c *MenuRoleController) GetMenuRoles(ctx *fiber.Ctx) error {
	menuID, err := uuid.Parse(ctx.Params("menu_id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid menu ID",
			Data:            map[string]interface{}{},
		})
	}

	roles, err := c.menuRoleService.GetMenuRoles(ctx.Context(), menuID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to retrieve menu roles",
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(response.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Menu roles retrieved successfully",
		Data:            roles,
	})
}

// GetRoleMenus handles GET /v1/roles/:role_id/menus requests
func (c *MenuRoleController) GetRoleMenus(ctx *fiber.Ctx) error {
	roleID, err := uuid.Parse(ctx.Params("role_id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid role ID",
			Data:            map[string]interface{}{},
		})
	}

	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))
	offset, _ := strconv.Atoi(ctx.Query("offset", "0"))

	menus, err := c.menuRoleService.GetRoleMenus(ctx.Context(), roleID, limit, offset)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to retrieve role menus",
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(response.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Role menus retrieved successfully",
		Data:            menus,
	})
}

// HasRole handles GET /v1/menus/:menu_id/roles/:role_id/check requests
func (c *MenuRoleController) HasRole(ctx *fiber.Ctx) error {
	menuID, err := uuid.Parse(ctx.Params("menu_id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid menu ID",
			Data:            map[string]interface{}{},
		})
	}

	roleID, err := uuid.Parse(ctx.Params("role_id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid role ID",
			Data:            map[string]interface{}{},
		})
	}

	hasRole, err := c.menuRoleService.HasRole(ctx.Context(), menuID, roleID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to check role assignment",
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(response.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Role check completed successfully",
		Data:            map[string]interface{}{"has_role": hasRole},
	})
}

// GetMenuRoleCount handles GET /v1/roles/:role_id/menus/count requests
func (c *MenuRoleController) GetMenuRoleCount(ctx *fiber.Ctx) error {
	roleID, err := uuid.Parse(ctx.Params("role_id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid role ID",
			Data:            map[string]interface{}{},
		})
	}

	count, err := c.menuRoleService.GetMenuRoleCount(ctx.Context(), roleID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to get menu count",
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(response.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Menu count retrieved successfully",
		Data:            map[string]interface{}{"count": count},
	})
}
