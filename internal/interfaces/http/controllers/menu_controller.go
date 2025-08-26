package controllers

import (
	"strconv"

	"github.com/turahe/go-restfull/internal/application/ports"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/interfaces/http/responses"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// MenuController handles HTTP requests for menu operations
//
//	@title						Menu Management API
//	@version					1.0
//	@description				This is a menu management API for creating, reading, updating, and deleting hierarchical menu structures
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
type MenuController struct {
	menuService ports.MenuService
}

func NewMenuController(menuService ports.MenuService) *MenuController {
	return &MenuController{
		menuService: menuService,
	}
}

// GetMenus handles GET /v1/menus requests
//
//	@Summary		Get all menus
//	@Description	Retrieve a paginated list of menus with optional filtering
//	@Tags			menus
//	@Accept			json
//	@Produce		json
//	@Param			limit	query		int												false	"Number of menus to return (default: 10, max: 100)"	default(10)	minimum(1)	maximum(100)
//	@Param			offset	query		int												false	"Number of menus to skip (default: 0)"				default(0)	minimum(0)
//	@Param			active	query		string											false	"Filter by active status (true/false)"				Enums(true, false)
//	@Param			visible	query		string											false	"Filter by visible status (true/false)"				Enums(true, false)
//	@Success		200		{object}	responses.MenuCollectionResponse				"List of menus"
//	@Failure		400		{object}	responses.CommonResponse						"Bad request - Invalid parameters"
//	@Failure		500		{object}	responses.CommonResponse						"Internal server error"
//	@Security		BearerAuth
//	@Router			/menus [get]
func (c *MenuController) GetMenus(ctx *fiber.Ctx) error {
	// Get pagination parameters
	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))
	offset, _ := strconv.Atoi(ctx.Query("offset", "0"))

	// Set reasonable limits
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	active := ctx.Query("active", "false")
	visible := ctx.Query("visible", "false")

	var menus []*entities.Menu
	var err error

	if visible == "true" {
		menus, err = c.menuService.GetVisibleMenus(ctx.Context(), limit, offset)
	} else if active == "true" {
		menus, err = c.menuService.GetActiveMenus(ctx.Context(), limit, offset)
	} else {
		menus, err = c.menuService.GetAllMenus(ctx.Context(), limit, offset)
	}

	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to retrieve menus",
			Data:            map[string]interface{}{},
		})
	}

	// Get total count for pagination
	total, err := c.menuService.GetMenuCount(ctx.Context())
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to retrieve menu count",
			Data:            map[string]interface{}{},
		})
	}

	// Calculate page from offset
	page := (offset / limit) + 1
	if offset == 0 {
		page = 1
	}

	// Build base URL for pagination links
	baseURL := ctx.BaseURL() + ctx.Path()

	// Return paginated menu collection response
	return ctx.Status(fiber.StatusOK).JSON(responses.NewPaginatedMenuCollectionResponse(
		menus, page, limit, int(total), baseURL,
	))
}

// GetMenuByID handles GET /v1/menus/:id requests
//
//	@Summary		Get menu by ID
//	@Description	Retrieve a specific menu by its unique identifier
//	@Tags			menus
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string													true	"Menu ID"	format(uuid)
//	@Success		200	{object}	responses.MenuResourceResponse					"Menu details"
//	@Failure		400	{object}	responses.CommonResponse								"Bad request - Invalid menu ID"
//	@Failure		404	{object}	responses.CommonResponse								"Not found - Menu does not exist"
//	@Failure		500	{object}	responses.CommonResponse								"Internal server error"
//	@Security		BearerAuth
//	@Router			/menus/{id} [get]
func (c *MenuController) GetMenuByID(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid menu ID",
			Data:            map[string]interface{}{},
		})
	}

	menu, err := c.menuService.GetMenuByID(ctx.Context(), id)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusNotFound,
			ResponseMessage: "Menu not found",
			Data:            map[string]interface{}{},
		})
	}

	// Return menu resource response
	return ctx.Status(fiber.StatusOK).JSON(responses.NewMenuResourceResponse(menu))
}

// GetMenuBySlug handles GET /v1/menus/slug/:slug requests
//
//	@Summary		Get menu by slug
//	@Description	Retrieve a menu by its URL-friendly slug
//	@Tags			menus
//	@Accept			json
//	@Produce		json
//	@Param			slug	path		string													true	"Menu slug"
//	@Success		200		{object}	responses.MenuResourceResponse					"Menu details"
//	@Failure		400		{object}	responses.CommonResponse								"Bad request - Slug is required"
//	@Failure		404		{object}	responses.CommonResponse								"Not found - Menu does not exist"
//	@Failure		500		{object}	responses.CommonResponse								"Internal server error"
//	@Security		BearerAuth
//	@Router			/menus/slug/{slug} [get]
func (c *MenuController) GetMenuBySlug(ctx *fiber.Ctx) error {
	slug := ctx.Params("slug")
	if slug == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Menu slug is required",
			Data:            map[string]interface{}{},
		})
	}

	menu, err := c.menuService.GetMenuBySlug(ctx.Context(), slug)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusNotFound,
			ResponseMessage: "Menu not found",
			Data:            map[string]interface{}{},
		})
	}

	// Return menu resource response
	return ctx.Status(fiber.StatusOK).JSON(responses.NewMenuResourceResponse(menu))
}

// GetRootMenus handles GET /v1/menus/root requests
//
//	@Summary		Get root menus
//	@Description	Retrieve all root-level menus (top-level menu items)
//	@Tags			menus
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	responses.MenuCollectionResponse				"List of root menus"
//	@Failure		500	{object}	responses.CommonResponse						"Internal server error"
//	@Security		BearerAuth
//	@Router			/menus/root [get]
func (c *MenuController) GetRootMenus(ctx *fiber.Ctx) error {
	menus, err := c.menuService.GetRootMenus(ctx.Context())
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to retrieve root menus",
			Data:            map[string]interface{}{},
		})
	}

	// Return menu collection response
	return ctx.Status(fiber.StatusOK).JSON(responses.NewMenuCollectionResponse(menus))
}

// GetMenuHierarchy handles GET /v1/menus/hierarchy requests
//
//	@Summary		Get menu hierarchy
//	@Description	Retrieve the complete menu hierarchy with nested structure
//	@Tags			menus
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	responses.MenuCollectionResponse				"Menu hierarchy"
//	@Failure		500	{object}	responses.CommonResponse						"Internal server error"
//	@Security		BearerAuth
//	@Router			/menus/hierarchy [get]
func (c *MenuController) GetMenuHierarchy(ctx *fiber.Ctx) error {
	menus, err := c.menuService.GetMenuHierarchy(ctx.Context())
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to retrieve menu hierarchy",
			Data:            menus,
		})
	}

	// Return menu collection response
	return ctx.Status(fiber.StatusOK).JSON(responses.NewMenuCollectionResponse(menus))
}

// GetUserMenus handles GET /v1/users/:user_id/menus requests
//
//	@Summary		Get user menus
//	@Description	Retrieve all menus accessible to a specific user based on their roles
//	@Tags			menus
//	@Accept			json
//	@Produce		json
//	@Param			user_id	path		string											true	"User ID"	format(uuid)
//	@Success		200		{object}	responses.MenuCollectionResponse				"List of user menus"
//	@Failure		400		{object}	responses.CommonResponse						"Bad request - Invalid user ID"
//	@Failure		404		{object}	responses.CommonResponse						"Not found - User does not exist"
//	@Failure		500		{object}	responses.CommonResponse						"Internal server error"
//	@Security		BearerAuth
//	@Router			/users/{user_id}/menus [get]
func (c *MenuController) GetUserMenus(ctx *fiber.Ctx) error {
	userID, err := uuid.Parse(ctx.Params("user_id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid user ID",
			Data:            map[string]interface{}{},
		})
	}

	menus, err := c.menuService.GetUserMenus(ctx.Context(), userID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to retrieve user menus",
			Data:            map[string]interface{}{},
		})
	}

	// Return menu collection response
	return ctx.Status(fiber.StatusOK).JSON(responses.NewMenuCollectionResponse(menus))
}

// SearchMenus handles GET /v1/menus/search requests
//
//	@Summary		Search menus
//	@Description	Search menus by name or description
//	@Tags			menus
//	@Accept			json
//	@Produce		json
//	@Param			q		query		string											true	"Search query"
//	@Param			limit	query		int												false	"Number of menus to return (default: 10, max: 100)"	default(10)	minimum(1)	maximum(100)
//	@Param			offset	query		int												false	"Number of menus to skip (default: 0)"				default(0)	minimum(0)
//	@Success		200		{object}	responses.MenuCollectionResponse				"List of matching menus"
//	@Failure		400		{object}	responses.CommonResponse						"Bad request - Search query is required"
//	@Failure		500		{object}	responses.CommonResponse						"Internal server error"
//	@Security		BearerAuth
//	@Router			/menus/search [get]
func (c *MenuController) SearchMenus(ctx *fiber.Ctx) error {
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

	menus, err := c.menuService.SearchMenus(ctx.Context(), query, limit, offset)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to search menus",
			Data:            map[string]interface{}{},
		})
	}

	// Return menu collection response
	return ctx.Status(fiber.StatusOK).JSON(responses.NewMenuCollectionResponse(menus))
}

// CreateMenu handles POST /v1/menus requests
//
//	@Summary		Create a new menu
//	@Description	Create a new menu item with the provided information
//	@Tags			menus
//	@Accept			json
//	@Produce		json
//	@Param			menu	body		object													true	"Menu creation request"	SchemaExample({"name": "Dashboard", "slug": "dashboard", "url": "/dashboard", "parent_id": "00000000-0000-0000-0000-000000000000"})
//	@Success		201		{object}	responses.MenuResourceResponse					"Menu created successfully"
//	@Failure		400		{object}	responses.CommonResponse								"Bad request - Invalid input data"
//	@Failure		409		{object}	responses.CommonResponse								"Conflict - Menu with same slug already exists"
//	@Failure		500		{object}	responses.CommonResponse								"Internal server error"
//	@Security		BearerAuth
//	@Router			/menus [post]
func (c *MenuController) CreateMenu(ctx *fiber.Ctx) error {
	var request struct {
		Name           string     `json:"name"`
		Slug           string     `json:"slug"`
		Description    string     `json:"description"`
		URL            string     `json:"url"`
		Icon           string     `json:"icon"`
		RecordOrdering int64      `json:"record_ordering"`
		ParentID       *uuid.UUID `json:"parent_id"`
		Target         string     `json:"target"`
	}

	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid request body",
			Data:            map[string]interface{}{},
		})
	}

	menu, err := c.menuService.CreateMenu(ctx.Context(), request.Name, request.Slug, request.Description, request.URL, request.Icon, request.RecordOrdering, request.ParentID)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: err.Error(),
			Data:            map[string]interface{}{},
		})
	}

	// Return menu resource response
	response := responses.NewMenuResourceResponse(menu)
	response.ResponseCode = fiber.StatusCreated
	response.ResponseMessage = "Menu created successfully"
	return ctx.Status(fiber.StatusCreated).JSON(response)
}

// UpdateMenu handles PUT /v1/menus/:id requests
//
//	@Summary		Update menu
//	@Description	Update an existing menu's information
//	@Tags			menus
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string													true	"Menu ID"				format(uuid)
//	@Param			menu	body		object													true	"Menu update request"	SchemaExample({"name": "Dashboard", "slug": "dashboard", "url": "/dashboard", "parent_id": "00000000-0000-0000-0000-000000000000"})
//	@Success		200		{object}	responses.MenuResourceResponse					"Menu updated successfully"
//	@Failure		400		{object}	responses.CommonResponse								"Bad request - Invalid input data"
//	@Failure		404		{object}	responses.CommonResponse								"Not found - Menu does not exist"
//	@Failure		500		{object}	responses.CommonResponse								"Internal server error"
//	@Security		BearerAuth
//	@Router			/menus/{id} [put]
func (c *MenuController) UpdateMenu(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid menu ID",
			Data:            map[string]interface{}{},
		})
	}

	var request struct {
		Name           string     `json:"name"`
		Slug           string     `json:"slug"`
		Description    string     `json:"description"`
		URL            string     `json:"url"`
		Icon           string     `json:"icon"`
		RecordOrdering int64      `json:"record_ordering"`
		ParentID       *uuid.UUID `json:"parent_id"`
		Target         string     `json:"target"`
	}

	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid request body",
			Data:            map[string]interface{}{},
		})
	}

	menu, err := c.menuService.UpdateMenu(ctx.Context(), id, request.Name, request.Slug, request.Description, request.URL, request.Icon, request.RecordOrdering, request.ParentID)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: err.Error(),
			Data:            map[string]interface{}{},
		})
	}

	// Return menu resource response
	return ctx.Status(fiber.StatusOK).JSON(responses.NewMenuResourceResponse(menu))
}

// DeleteMenu handles DELETE /v1/menus/:id requests
//
//	@Summary		Delete menu
//	@Description	Delete a menu item (soft delete)
//	@Tags			menus
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string						true	"Menu ID"	format(uuid)
//	@Success		200	{object}	responses.CommonResponse	"Menu deleted successfully"
//	@Failure		400	{object}	responses.CommonResponse	"Bad request - Invalid menu ID"
//	@Failure		404	{object}	responses.CommonResponse	"Not found - Menu does not exist"
//	@Failure		500	{object}	responses.CommonResponse	"Internal server error"
//	@Security		BearerAuth
//	@Router			/menus/{id} [delete]
func (c *MenuController) DeleteMenu(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid menu ID",
			Data:            map[string]interface{}{},
		})
	}

	err = c.menuService.DeleteMenu(ctx.Context(), id)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: err.Error(),
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(responses.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Menu deleted successfully",
		Data:            map[string]interface{}{},
	})
}

// ActivateMenu handles PATCH /v1/menus/:id/activate requests
//
//	@Summary		Activate menu
//	@Description	Activate a deactivated menu item
//	@Tags			menus
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string													true	"Menu ID"	format(uuid)
//	@Success		200	{object}	responses.MenuResourceResponse					"Menu activated successfully"
//	@Failure		400	{object}	responses.CommonResponse								"Bad request - Invalid menu ID"
//	@Failure		404	{object}	responses.CommonResponse								"Not found - Menu does not exist"
//	@Failure		500	{object}	responses.CommonResponse								"Internal server error"
//	@Security		BearerAuth
//	@Router			/menus/{id}/activate [patch]
func (c *MenuController) ActivateMenu(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid menu ID",
			Data:            map[string]interface{}{},
		})
	}

	menu, err := c.menuService.ActivateMenu(ctx.Context(), id)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: err.Error(),
			Data:            map[string]interface{}{},
		})
	}

	// Return menu resource response
	return ctx.Status(fiber.StatusOK).JSON(responses.NewMenuResourceResponse(menu))
}

// DeactivateMenu handles PATCH /v1/menus/:id/deactivate requests
//
//	@Summary		Deactivate menu
//	@Description	Deactivate an active menu item
//	@Tags			menus
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string													true	"Menu ID"	format(uuid)
//	@Success		200	{object}	responses.MenuResourceResponse					"Menu deactivated successfully"
//	@Failure		400	{object}	responses.CommonResponse								"Bad request - Invalid menu ID"
//	@Failure		404	{object}	responses.CommonResponse								"Not found - Menu does not exist"
//	@Failure		500	{object}	responses.CommonResponse								"Internal server error"
//	@Security		BearerAuth
//	@Router			/menus/{id}/deactivate [patch]
func (c *MenuController) DeactivateMenu(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid menu ID",
			Data:            map[string]interface{}{},
		})
	}

	menu, err := c.menuService.DeactivateMenu(ctx.Context(), id)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: err.Error(),
			Data:            map[string]interface{}{},
		})
	}

	// Return menu resource response
	return ctx.Status(fiber.StatusOK).JSON(responses.NewMenuResourceResponse(menu))
}

// ShowMenu handles PATCH /v1/menus/:id/show requests
//
//	@Summary		Show menu
//	@Description	Make a hidden menu item visible
//	@Tags			menus
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string													true	"Menu ID"	format(uuid)
//	@Success		200	{object}	responses.MenuResourceResponse					"Menu shown successfully"
//	@Failure		400	{object}	responses.CommonResponse								"Bad request - Invalid menu ID"
//	@Failure		404	{object}	responses.CommonResponse								"Not found - Menu does not exist"
//	@Failure		500	{object}	responses.CommonResponse								"Internal server error"
//	@Security		BearerAuth
//	@Router			/menus/{id}/show [patch]
func (c *MenuController) ShowMenu(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid menu ID",
			Data:            map[string]interface{}{},
		})
	}

	menu, err := c.menuService.ShowMenu(ctx.Context(), id)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: err.Error(),
			Data:            map[string]interface{}{},
		})
	}

	// Return menu resource response
	return ctx.Status(fiber.StatusOK).JSON(responses.NewMenuResourceResponse(menu))
}

// HideMenu handles PATCH /v1/menus/:id/hide requests
//
//	@Summary		Hide menu
//	@Description	Make a visible menu item hidden
//	@Tags			menus
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string													true	"Menu ID"	format(uuid)
//	@Success		200	{object}	responses.MenuResourceResponse					"Menu hidden successfully"
//	@Failure		400	{object}	responses.CommonResponse								"Bad request - Invalid menu ID"
//	@Failure		404	{object}	responses.CommonResponse								"Not found - Menu does not exist"
//	@Failure		500	{object}	responses.CommonResponse								"Internal server error"
//	@Security		BearerAuth
//	@Router			/menus/{id}/hide [patch]
func (c *MenuController) HideMenu(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid menu ID",
			Data:            map[string]interface{}{},
		})
	}

	menu, err := c.menuService.HideMenu(ctx.Context(), id)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: err.Error(),
			Data:            map[string]interface{}{},
		})
	}

	// Return menu resource response
	return ctx.Status(fiber.StatusOK).JSON(responses.NewMenuResourceResponse(menu))
}
