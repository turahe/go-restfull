package controllers

import (
	"strconv"
	"webapi/internal/application/ports"
	"webapi/internal/interfaces/http/responses"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// RoleController handles HTTP requests for role operations
// @title Role Management API
// @version 1.0
// @description This is a role management API for creating, reading, updating, and deleting user roles
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
type RoleController struct {
	roleService ports.RoleService
}

func NewRoleController(roleService ports.RoleService) *RoleController {
	return &RoleController{
		roleService: roleService,
	}
}

// GetRoles handles GET /v1/roles requests
// @Summary Get all roles
// @Description Retrieve a paginated list of roles with optional filtering
// @Tags roles
// @Accept json
// @Produce json
// @Param limit query int false "Number of roles to return (default: 10, max: 100)" default(10) minimum(1) maximum(100)
// @Param offset query int false "Number of roles to skip (default: 0)" default(0) minimum(0)
// @Param active query string false "Filter by active status (true/false)" Enums(true, false)
// @Success 200 {object} responses.CommonResponse{data=[]interface{}} "List of roles"
// @Failure 400 {object} responses.CommonResponse "Bad request - Invalid parameters"
// @Failure 500 {object} responses.CommonResponse "Internal server error"
// @Security BearerAuth
// @Router /roles [get]
func (c *RoleController) GetRoles(ctx *fiber.Ctx) error {
	// Get pagination parameters
	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))
	offset, _ := strconv.Atoi(ctx.Query("offset", "0"))
	active := ctx.Query("active", "false")

	var roles interface{}
	var err error

	if active == "true" {
		roles, err = c.roleService.GetActiveRoles(ctx.Context(), limit, offset)
	} else {
		roles, err = c.roleService.GetAllRoles(ctx.Context(), limit, offset)
	}

	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to retrieve roles",
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(responses.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Roles retrieved successfully",
		Data:            roles,
	})
}

// GetRoleByID handles GET /v1/roles/:id requests
// @Summary Get role by ID
// @Description Retrieve a specific role by its unique identifier
// @Tags roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID" format(uuid)
// @Success 200 {object} responses.CommonResponse{data=map[string]interface{}} "Role details"
// @Failure 400 {object} responses.CommonResponse "Bad request - Invalid role ID"
// @Failure 404 {object} responses.CommonResponse "Not found - Role does not exist"
// @Failure 500 {object} responses.CommonResponse "Internal server error"
// @Security BearerAuth
// @Router /roles/{id} [get]
func (c *RoleController) GetRoleByID(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid role ID",
			Data:            map[string]interface{}{},
		})
	}

	role, err := c.roleService.GetRoleByID(ctx.Context(), id)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusNotFound,
			ResponseMessage: "Role not found",
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(responses.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Role retrieved successfully",
		Data:            role,
	})
}

// GetRoleBySlug handles GET /v1/roles/slug/:slug requests
// @Summary Get role by slug
// @Description Retrieve a role by its URL-friendly slug
// @Tags roles
// @Accept json
// @Produce json
// @Param slug path string true "Role slug"
// @Success 200 {object} responses.CommonResponse{data=map[string]interface{}} "Role details"
// @Failure 400 {object} responses.CommonResponse "Bad request - Slug is required"
// @Failure 404 {object} responses.CommonResponse "Not found - Role does not exist"
// @Failure 500 {object} responses.CommonResponse "Internal server error"
// @Security BearerAuth
// @Router /roles/slug/{slug} [get]
func (c *RoleController) GetRoleBySlug(ctx *fiber.Ctx) error {
	slug := ctx.Params("slug")
	if slug == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Role slug is required",
			Data:            map[string]interface{}{},
		})
	}

	role, err := c.roleService.GetRoleBySlug(ctx.Context(), slug)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusNotFound,
			ResponseMessage: "Role not found",
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(responses.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Role retrieved successfully",
		Data:            role,
	})
}

// CreateRole handles POST /v1/roles requests
// @Summary Create a new role
// @Description Create a new role with the provided information
// @Tags roles
// @Accept json
// @Produce json
// @Param role body object true "Role creation request" SchemaExample({"name": "Admin", "slug": "admin", "description": "Administrator role"})
// @Success 201 {object} responses.CommonResponse{data=map[string]interface{}} "Role created successfully"
// @Failure 400 {object} responses.CommonResponse "Bad request - Invalid input data"
// @Failure 409 {object} responses.CommonResponse "Conflict - Role with same slug already exists"
// @Failure 500 {object} responses.CommonResponse "Internal server error"
// @Security BearerAuth
// @Router /roles [post]
func (c *RoleController) CreateRole(ctx *fiber.Ctx) error {
	var request struct {
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
	}

	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid request body",
			Data:            map[string]interface{}{},
		})
	}

	role, err := c.roleService.CreateRole(ctx.Context(), request.Name, request.Slug, request.Description)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: err.Error(),
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(responses.CommonResponse{
		ResponseCode:    fiber.StatusCreated,
		ResponseMessage: "Role created successfully",
		Data:            role,
	})
}

// UpdateRole handles PUT /v1/roles/:id requests
// @Summary Update role
// @Description Update an existing role's information
// @Tags roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID" format(uuid)
// @Param role body object true "Role update request" SchemaExample({"name": "Admin", "slug": "admin", "description": "Updated administrator role"})
// @Success 200 {object} responses.CommonResponse{data=map[string]interface{}} "Role updated successfully"
// @Failure 400 {object} responses.CommonResponse "Bad request - Invalid input data"
// @Failure 404 {object} responses.CommonResponse "Not found - Role does not exist"
// @Failure 500 {object} responses.CommonResponse "Internal server error"
// @Security BearerAuth
// @Router /roles/{id} [put]
func (c *RoleController) UpdateRole(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid role ID",
			Data:            map[string]interface{}{},
		})
	}

	var request struct {
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
	}

	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid request body",
			Data:            map[string]interface{}{},
		})
	}

	role, err := c.roleService.UpdateRole(ctx.Context(), id, request.Name, request.Slug, request.Description)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: err.Error(),
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(responses.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Role updated successfully",
		Data:            role,
	})
}

// DeleteRole handles DELETE /v1/roles/:id requests
// @Summary Delete role
// @Description Delete a role (soft delete)
// @Tags roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID" format(uuid)
// @Success 200 {object} responses.CommonResponse "Role deleted successfully"
// @Failure 400 {object} responses.CommonResponse "Bad request - Invalid role ID"
// @Failure 404 {object} responses.CommonResponse "Not found - Role does not exist"
// @Failure 500 {object} responses.CommonResponse "Internal server error"
// @Security BearerAuth
// @Router /roles/{id} [delete]
func (c *RoleController) DeleteRole(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid role ID",
			Data:            map[string]interface{}{},
		})
	}

	err = c.roleService.DeleteRole(ctx.Context(), id)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: err.Error(),
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(responses.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Role deleted successfully",
		Data:            map[string]interface{}{},
	})
}

// ActivateRole handles PUT /v1/roles/:id/activate requests
// @Summary Activate role
// @Description Activate a deactivated role
// @Tags roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID" format(uuid)
// @Success 200 {object} responses.CommonResponse{data=map[string]interface{}} "Role activated successfully"
// @Failure 400 {object} responses.CommonResponse "Bad request - Invalid role ID"
// @Failure 404 {object} responses.CommonResponse "Not found - Role does not exist"
// @Failure 500 {object} responses.CommonResponse "Internal server error"
// @Security BearerAuth
// @Router /roles/{id}/activate [put]
func (c *RoleController) ActivateRole(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid role ID",
			Data:            map[string]interface{}{},
		})
	}

	role, err := c.roleService.ActivateRole(ctx.Context(), id)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: err.Error(),
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(responses.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Role activated successfully",
		Data:            role,
	})
}

// DeactivateRole handles PUT /v1/roles/:id/deactivate requests
// @Summary Deactivate role
// @Description Deactivate an active role
// @Tags roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID" format(uuid)
// @Success 200 {object} responses.CommonResponse{data=map[string]interface{}} "Role deactivated successfully"
// @Failure 400 {object} responses.CommonResponse "Bad request - Invalid role ID"
// @Failure 404 {object} responses.CommonResponse "Not found - Role does not exist"
// @Failure 500 {object} responses.CommonResponse "Internal server error"
// @Security BearerAuth
// @Router /roles/{id}/deactivate [put]
func (c *RoleController) DeactivateRole(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid role ID",
			Data:            map[string]interface{}{},
		})
	}

	role, err := c.roleService.DeactivateRole(ctx.Context(), id)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: err.Error(),
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(responses.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Role deactivated successfully",
		Data:            role,
	})
}

// SearchRoles handles GET /v1/roles/search requests
// @Summary Search roles
// @Description Search roles by name or description
// @Tags roles
// @Accept json
// @Produce json
// @Param query query string true "Search query"
// @Param limit query int false "Number of roles to return (default: 10, max: 100)" default(10) minimum(1) maximum(100)
// @Param offset query int false "Number of roles to skip (default: 0)" default(0) minimum(0)
// @Success 200 {object} responses.CommonResponse{data=[]interface{}} "List of matching roles"
// @Failure 400 {object} responses.CommonResponse "Bad request - Invalid parameters"
// @Failure 500 {object} responses.CommonResponse "Internal server error"
// @Security BearerAuth
// @Router /roles/search [get]
func (c *RoleController) SearchRoles(ctx *fiber.Ctx) error {
	query := ctx.Query("q")
	if query == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Search query is required",
			Data:            map[string]interface{}{},
		})
	}

	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))
	offset, _ := strconv.Atoi(ctx.Query("offset", "0"))

	roles, err := c.roleService.SearchRoles(ctx.Context(), query, limit, offset)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to search roles",
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(responses.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Roles search completed successfully",
		Data:            roles,
	})
}
