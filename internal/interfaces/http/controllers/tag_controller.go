package controllers

import (
	"strconv"

	"github.com/turahe/go-restfull/internal/application/ports"
	"github.com/turahe/go-restfull/internal/interfaces/http/requests"
	"github.com/turahe/go-restfull/internal/interfaces/http/responses"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// TagController handles HTTP requests for tag operations
//
//	@title						Tag Management API
//	@version					1.0
//	@description				This is a tag management API for creating, reading, updating, and deleting tags
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
type TagController struct {
	tagService ports.TagService
}

func NewTagController(tagService ports.TagService) *TagController {
	return &TagController{
		tagService: tagService,
	}
}

// GetTags handles GET /v1/tags requests
//
//	@Summary		Get all tags
//	@Description	Retrieve a list of all tags with pagination
//	@Tags			tags
//	@Accept			json
//	@Produce		json
//	@Param			limit	query		int												false	"Number of tags to return (default: 10, max: 100)"
//	@Param			offset	query		int												false	"Number of tags to skip (default: 0)"
//	@Success		200		{object}	responses.TagCollectionResponse					"List of tags"
//	@Failure		400		{object}	responses.ErrorResponse							"Bad request"
//	@Failure		500		{object}	responses.ErrorResponse							"Internal server error"
//	@Security		BearerAuth
//	@Router			/tags [get]
func (c *TagController) GetTags(ctx *fiber.Ctx) error {
	// Parse query parameters
	limitStr := ctx.Query("limit", "10")
	offsetStr := ctx.Query("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid limit parameter",
		})
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid offset parameter",
		})
	}

	// Set reasonable limits
	if limit > 100 {
		limit = 100
	}

	// Get tags from service
	tags, err := c.tagService.GetAllTags(ctx.Context(), limit, offset)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Failed to retrieve tags: " + err.Error(),
		})
	}

	// Get total count for pagination
	total, err := c.tagService.GetTagCount(ctx.Context())
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Failed to retrieve tag count",
		})
	}

	// Calculate page from offset
	page := (offset / limit) + 1
	if offset == 0 {
		page = 1
	}

	// Build base URL for pagination links
	baseURL := ctx.BaseURL() + ctx.Path()

	// Return paginated tag collection response (Laravel style)
	return ctx.Status(fiber.StatusOK).JSON(responses.NewPaginatedTagCollection(
		tags, page, limit, int(total), baseURL,
	))
}

// GetTagByID handles GET /v1/tags/:id requests
//
//	@Summary		Get tag by ID
//	@Description	Retrieve a tag by its ID
//	@Tags			tags
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string										true	"Tag ID"	format(uuid)
//	@Success		200	{object}	responses.TagResourceResponse					"Tag retrieved successfully"
//	@Failure		400	{object}	responses.ErrorResponse						"Bad request - Invalid tag ID"
//	@Failure		404	{object}	responses.ErrorResponse						"Not found - Tag does not exist"
//	@Failure		500	{object}	responses.ErrorResponse						"Internal server error"
//	@Security		BearerAuth
//	@Router			/tags/{id} [get]
func (c *TagController) GetTagByID(ctx *fiber.Ctx) error {
	// Parse tag ID
	tagIDStr := ctx.Params("id")
	tagID, err := uuid.Parse(tagIDStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid tag ID format",
		})
	}

	// Get tag from service
	tag, err := c.tagService.GetTagByID(ctx.Context(), tagID)
	if err != nil {
		// Check if it's a not found error
		if err.Error() == "tag not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(responses.ErrorResponse{
				Status:  "error",
				Message: "Tag not found",
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Failed to retrieve tag: " + err.Error(),
		})
	}

	// Return tag resource response
	return ctx.Status(fiber.StatusOK).JSON(responses.NewTagResource(tag))
}

// CreateTag handles POST /v1/tags requests
//
//	@Summary		Create tag
//	@Description	Create a new tag
//	@Tags			tags
//	@Accept			json
//	@Produce		json
//	@Param			tag	body		requests.CreateTagRequest	true	"Tag object"
//	@Success		201	{object}	responses.TagResourceResponse	"Tag created successfully"
//	@Failure		422	{object}	responses.ValidationErrorResponse	"Validation errors"
//	@Failure		409	{object}	responses.ErrorResponse		"Conflict - Tag with same slug already exists"
//	@Failure		500	{object}	responses.ErrorResponse		"Internal server error"
//	@Security		BearerAuth
//	@Router			/tags [post]
func (c *TagController) CreateTag(ctx *fiber.Ctx) error {
	// Parse request body
	var req requests.CreateTagRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid request body: " + err.Error(),
		})
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Validation failed: " + err.Error(),
		})
	}

	// Transform request to entity
	tag := req.ToEntity()

	// Get user ID from context (set by JWT middleware)
	userIDInterface := ctx.Locals("user_id")
	if userIDInterface != nil {
		if userID, ok := userIDInterface.(uuid.UUID); ok {
			tag.CreatedBy = userID
			tag.UpdatedBy = userID
		}
	}

	// Create tag using service
	createdTag, err := c.tagService.CreateTag(ctx.Context(), tag)
	if err != nil {
		// Check for specific errors
		if err.Error() == "tag with this slug already exists" {
			return ctx.Status(fiber.StatusConflict).JSON(responses.ErrorResponse{
				Status:  "error",
				Message: "Tag with this slug already exists",
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Failed to create tag: " + err.Error(),
		})
	}

	// Return tag resource response
	return ctx.Status(fiber.StatusCreated).JSON(responses.NewTagResource(createdTag))
}

// UpdateTag handles PUT /v1/tags/:id requests
//
//	@Summary		Update tag
//	@Description	Update an existing tag by its ID
//	@Tags			tags
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string										true	"Tag ID"	format(uuid)
//	@Param			tag		body		requests.UpdateTagRequest					true	"Tag object"
//	@Success		200		{object}	responses.TagResourceResponse				"Tag updated successfully"
//	@Failure		400		{object}	responses.ErrorResponse						"Bad request - Invalid tag ID or input data"
//	@Failure		404		{object}	responses.ErrorResponse						"Not found - Tag does not exist"
//	@Failure		409		{object}	responses.ErrorResponse						"Conflict - Tag with same slug already exists"
//	@Failure		500		{object}	responses.ErrorResponse						"Internal server error"
//	@Security		BearerAuth
//	@Router			/tags/{id} [put]
func (c *TagController) UpdateTag(ctx *fiber.Ctx) error {
	// Parse tag ID
	tagIDStr := ctx.Params("id")
	tagID, err := uuid.Parse(tagIDStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid tag ID format",
		})
	}

	// Parse request body
	var req requests.UpdateTagRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid request body: " + err.Error(),
		})
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Validation failed: " + err.Error(),
		})
	}

	// Get existing tag
	existingTag, err := c.tagService.GetTagByID(ctx.Context(), tagID)
	if err != nil {
		if err.Error() == "tag not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(responses.ErrorResponse{
				Status:  "error",
				Message: "Tag not found",
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Failed to retrieve tag: " + err.Error(),
		})
	}

	// Transform request to entity
	updatedTag := req.ToEntity(existingTag)

	// Get user ID from context (set by JWT middleware) for updated_by
	userIDInterface := ctx.Locals("user_id")
	if userIDInterface != nil {
		if userID, ok := userIDInterface.(uuid.UUID); ok {
			updatedTag.UpdatedBy = userID
		}
	}

	// Update tag using service
	tag, err := c.tagService.UpdateTag(ctx.Context(), updatedTag)
	if err != nil {
		// Check for specific errors
		if err.Error() == "tag not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(responses.ErrorResponse{
				Status:  "error",
				Message: "Tag not found",
			})
		}
		if err.Error() == "tag with this slug already exists" {
			return ctx.Status(fiber.StatusConflict).JSON(responses.ErrorResponse{
				Status:  "error",
				Message: "Tag with this slug already exists",
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Failed to update tag: " + err.Error(),
		})
	}

	// Return tag resource response
	return ctx.Status(fiber.StatusOK).JSON(responses.NewTagResource(tag))
}

// DeleteTag handles DELETE /v1/tags/:id requests
//
//	@Summary		Delete tag
//	@Description	Delete a tag by its ID
//	@Tags			tags
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string						true	"Tag ID"	format(uuid)
//	@Success		200	{object}	responses.SuccessResponse	"Tag deleted successfully"
//	@Failure		400	{object}	responses.ErrorResponse		"Bad request - Invalid tag ID"
//	@Failure		404	{object}	responses.ErrorResponse		"Not found - Tag does not exist"
//	@Failure		500	{object}	responses.ErrorResponse		"Internal server error"
//	@Security		BearerAuth
//	@Router			/tags/{id} [delete]
func (c *TagController) DeleteTag(ctx *fiber.Ctx) error {
	// Parse tag ID
	tagIDStr := ctx.Params("id")
	tagID, err := uuid.Parse(tagIDStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid tag ID format",
		})
	}

	// Delete tag using service
	err = c.tagService.DeleteTag(ctx.Context(), tagID)
	if err != nil {
		// Check for specific errors
		if err.Error() == "tag not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(responses.ErrorResponse{
				Status:  "error",
				Message: "Tag not found",
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Failed to delete tag: " + err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(responses.SuccessResponse{
		Status:  "success",
		Message: "Tag deleted successfully",
		Data:    nil,
	})
}

// SearchTags handles GET /v1/tags/search requests
//
//	@Summary		Search tags
//	@Description	Search tags by query with pagination
//	@Tags			tags
//	@Accept			json
//	@Produce		json
//	@Param			query	query		string											true	"Search query"
//	@Param			limit	query		int												false	"Number of tags to return (default: 10, max: 100)"
//	@Param			offset	query		int												false	"Number of tags to skip (default: 0)"
//	@Success		200		{object}	responses.TagCollectionResponse				"Search results"
//	@Failure		400		{object}	responses.ErrorResponse							"Bad request"
//	@Failure		500		{object}	responses.ErrorResponse							"Internal server error"
//	@Security		BearerAuth
//	@Router			/tags/search [get]
func (c *TagController) SearchTags(ctx *fiber.Ctx) error {
	// Parse query parameters
	query := ctx.Query("query")
	if query == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Query parameter is required",
		})
	}

	limitStr := ctx.Query("limit", "10")
	offsetStr := ctx.Query("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid limit parameter",
		})
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid offset parameter",
		})
	}

	// Set reasonable limits
	if limit > 100 {
		limit = 100
	}

	// Search tags using service
	tags, err := c.tagService.SearchTags(ctx.Context(), query, limit, offset)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Failed to search tags: " + err.Error(),
		})
	}

	// For search results, we'll return a simple collection response (Laravel style)
	// since we don't have a total count for search results
	return ctx.Status(fiber.StatusOK).JSON(responses.NewTagCollection(tags))
}
