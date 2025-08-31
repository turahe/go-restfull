package controllers

import (
	"fmt"
	"strconv"

	"github.com/turahe/go-restfull/internal/application/ports"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/interfaces/http/requests"
	"github.com/turahe/go-restfull/internal/interfaces/http/responses"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// RoleController handles HTTP requests for role operations
//
//	@title						Role Management API
//	@version					1.0
//	@description				This is a role management API for creating, reading, updating, and deleting user roles
//	@termsOfService				http://swagger.io/terms/
//	@contact.name				API Support
//	@contact.email				support@example.com
//	@license.name				MIT
//	@license.url				https://opensource.org/licenses/MIT
//	@host						localhost:8000
//	@BasePath					/api/v1
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.
type RoleController struct {
	roleService ports.RoleService
	baseURL     string
}

func NewRoleController(roleService ports.RoleService) *RoleController {
	return &RoleController{
		roleService: roleService,
		baseURL:     "/api/v1",
	}
}

// GetRoles godoc
// @Summary Get all roles
// @Description Get all roles with pagination
// @Tags roles
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(10)
// @Param active query string false "Filter by active status (true/false)" Enums(true, false)
// @Success 200 {object} responses.RoleCollectionResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /roles [get]
func (c *RoleController) GetRoles(ctx *fiber.Ctx) error {
	// Get pagination parameters
	page := ctx.QueryInt("page", 1)
	limit := ctx.QueryInt("limit", 10)
	active := ctx.Query("active", "false")

	var roles []*entities.Role
	var err error

	if active == "true" {
		roles, err = c.roleService.GetActiveRoles(ctx.Context(), limit, (page-1)*limit)
	} else {
		roles, err = c.roleService.GetAllRoles(ctx.Context(), limit, (page-1)*limit)
	}

	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Failed to retrieve roles",
		})
	}

	// Get total count for pagination
	total, err := c.roleService.GetRoleCount(ctx.Context())
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Failed to get role count",
		})
	}

	// Build base URL for pagination links
	baseURL := fmt.Sprintf("%s/roles", c.baseURL)

	return ctx.JSON(responses.NewPaginatedRoleCollection(roles, page, limit, int(total), baseURL))
}

// GetRoleByID godoc
// @Summary Get role by ID
// @Description Retrieve a specific role by its unique identifier
// @Tags roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID" format(uuid)
// @Success 200 {object} responses.RoleResourceResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /roles/{id} [get]
func (c *RoleController) GetRoleByID(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid role ID",
		})
	}

	role, err := c.roleService.GetRoleByID(ctx.Context(), id)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Role not found",
		})
	}

	return ctx.JSON(responses.NewRoleResource(role))
}

// GetRoleBySlug godoc
// @Summary Get role by slug
// @Description Retrieve a role by its URL-friendly slug
// @Tags roles
// @Accept json
// @Produce json
// @Param slug path string true "Role slug"
// @Success 200 {object} responses.RoleResourceResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /roles/slug/{slug} [get]
func (c *RoleController) GetRoleBySlug(ctx *fiber.Ctx) error {
	slug := ctx.Params("slug")
	if slug == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Role slug is required",
		})
	}

	role, err := c.roleService.GetRoleBySlug(ctx.Context(), slug)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Role not found",
		})
	}

	return ctx.JSON(responses.NewRoleResource(role))
}

// CreateRole godoc
// @Summary Create a new role
// @Description Create a new role with the provided information
// @Tags roles
// @Accept json
// @Produce json
// @Param role body object true "Role creation request" SchemaExample({"name": "Admin", "slug": "admin", "description": "Administrator role"})
// @Success 201 {object} responses.RoleResourceResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 409 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /roles [post]
func (c *RoleController) CreateRole(ctx *fiber.Ctx) error {
	var request requests.CreateRoleRequest

	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid request body",
		})
	}

	// Validate request
	if err := request.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Validation failed: " + err.Error(),
		})
	}

	// Transform request to entity
	role := request.ToEntity()

	// Create role using the entity
	createdRole, err := c.roleService.CreateRole(ctx.Context(), role.Name, role.Slug, role.Description)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(responses.NewRoleResource(createdRole))
}

// UpdateRole godoc
// @Summary Update role
// @Description Update an existing role's information
// @Tags roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID" format(uuid)
// @Param role body object true "Role update request" SchemaExample({"name": "Admin", "slug": "admin", "description": "Updated administrator role"})
// @Success 200 {object} responses.RoleResourceResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /roles/{id} [put]
func (c *RoleController) UpdateRole(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid role ID",
		})
	}

	var request requests.UpdateRoleRequest

	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid request body",
		})
	}

	// Validate request
	if err := request.Validate(); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Validation failed: " + err.Error(),
		})
	}

	// Get existing role
	existingRole, err := c.roleService.GetRoleByID(ctx.Context(), id)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Role not found",
		})
	}

	// Transform request to entity
	updatedRole := request.ToEntity(existingRole)

	// Update role using the entity
	role, err := c.roleService.UpdateRole(ctx.Context(), id, updatedRole.Name, updatedRole.Slug, updatedRole.Description)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.JSON(responses.NewRoleResource(role))
}

// DeleteRole godoc
// @Summary Delete role
// @Description Delete a role (soft delete)
// @Tags roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID" format(uuid)
// @Success 200 {object} responses.ErrorResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /roles/{id} [delete]
func (c *RoleController) DeleteRole(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid role ID",
		})
	}

	err = c.roleService.DeleteRole(ctx.Context(), id)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(responses.ErrorResponse{
		Status:  "success",
		Message: "Role deleted successfully",
	})
}

// ActivateRole godoc
// @Summary Activate role
// @Description Activate a deactivated role
// @Tags roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID" format(uuid)
// @Success 200 {object} responses.RoleResourceResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /roles/{id}/activate [put]
func (c *RoleController) ActivateRole(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid role ID",
		})
	}

	role, err := c.roleService.ActivateRole(ctx.Context(), id)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.JSON(responses.NewRoleResource(role))
}

// DeactivateRole godoc
// @Summary Deactivate role
// @Description Deactivate an active role
// @Tags roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID" format(uuid)
// @Success 200 {object} responses.RoleResourceResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /roles/{id}/deactivate [put]
func (c *RoleController) DeactivateRole(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid role ID",
		})
	}

	role, err := c.roleService.DeactivateRole(ctx.Context(), id)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.JSON(responses.NewRoleResource(role))
}

// SearchRoles godoc
// @Summary Search roles
// @Description Search roles by name or description
// @Tags roles
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param limit query int false "Number of roles to return (default: 10, max: 100)" default(10) minimum(1) maximum(100)
// @Param offset query int false "Number of roles to skip (default: 0)" default(0) minimum(0)
// @Success 200 {object} responses.RoleCollectionResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /roles/search [get]
func (c *RoleController) SearchRoles(ctx *fiber.Ctx) error {
	query := ctx.Query("q")
	if query == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Search query is required",
		})
	}

	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))
	offset, _ := strconv.Atoi(ctx.Query("offset", "0"))

	roles, err := c.roleService.SearchRoles(ctx.Context(), query, limit, offset)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Failed to search roles",
		})
	}

	return ctx.JSON(responses.NewRoleCollection(roles))
}
