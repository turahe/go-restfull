package controllers

import (
	"net/http"
	"strconv"

	"github.com/turahe/go-restfull/internal/application/ports"
	"github.com/turahe/go-restfull/internal/helper/pagination"
	"github.com/turahe/go-restfull/internal/interfaces/http/requests"
	"github.com/turahe/go-restfull/internal/interfaces/http/responses"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TaxonomyController struct {
	taxonomyService ports.TaxonomyService
}

func NewTaxonomyController(taxonomyService ports.TaxonomyService) *TaxonomyController {
	return &TaxonomyController{
		taxonomyService: taxonomyService,
	}
}

// CreateTaxonomy godoc
//
//	@Summary		Create a new taxonomy
//	@Description	Create a new taxonomy with optional parent taxonomy
//	@Tags			taxonomies
//	@Accept			json
//	@Produce		json
//	@Param			request	body		requests.CreateTaxonomyRequest						true	"Taxonomy creation request"
//	@Success		201		{object}	responses.TaxonomyResourceResponse	"Taxonomy created successfully"
//	@Failure		422		{object}	responses.ErrorResponse								"Validation errors"
//	@Failure		500		{object}	responses.ErrorResponse								"Internal server error"
//	@Router			/api/v1/taxonomies [post]
//	@Security		BearerAuth
func (c *TaxonomyController) CreateTaxonomy(ctx *fiber.Ctx) error {
	var req requests.CreateTaxonomyRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid request body",
		})
	}

	if err := req.Validate(); err != nil {
		return ctx.Status(http.StatusUnprocessableEntity).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	// Transform request to entity
	taxonomy, err := req.ToEntity()
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	// Create taxonomy using the entity
	createdTaxonomy, err := c.taxonomyService.CreateTaxonomy(ctx.Context(), taxonomy)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusCreated).JSON(responses.NewTaxonomyResource(createdTaxonomy))
}

// GetTaxonomyByID godoc
//
//	@Summary		Get taxonomy by ID
//	@Description	Retrieve a taxonomy by its unique identifier
//	@Tags			taxonomies
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string												true	"Taxonomy ID"	format(uuid)
//	@Success		200	{object}	responses.TaxonomyResourceResponse	"Taxonomy found"
//	@Failure		400	{object}	responses.ErrorResponse								"Bad request - Invalid taxonomy ID format"
//	@Failure		404	{object}	responses.ErrorResponse								"Taxonomy not found"
//	@Router			/api/v1/taxonomies/{id} [get]
//	@Security		BearerAuth
func (c *TaxonomyController) GetTaxonomyByID(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid taxonomy ID format",
		})
	}

	taxonomy, err := c.taxonomyService.GetTaxonomyByID(ctx.Context(), id)
	if err != nil {
		return ctx.Status(http.StatusNotFound).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Taxonomy not found",
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.NewTaxonomyResource(taxonomy))
}

// GetTaxonomyBySlug godoc
//
//	@Summary		Get taxonomy by slug
//	@Description	Retrieve a taxonomy by its slug identifier
//	@Tags			taxonomies
//	@Accept			json
//	@Produce		json
//	@Param			slug	path		string												true	"Taxonomy slug"
//	@Success		200		{object}	responses.TaxonomyResourceResponse	"Taxonomy found"
//	@Failure		400		{object}	responses.ErrorResponse								"Bad request - Slug is required"
//	@Failure		404		{object}	responses.ErrorResponse								"Taxonomy not found"
//	@Router			/api/v1/taxonomies/slug/{slug} [get]
//	@Security		BearerAuth
func (c *TaxonomyController) GetTaxonomyBySlug(ctx *fiber.Ctx) error {
	slug := ctx.Params("slug")
	taxonomy, err := c.taxonomyService.GetTaxonomyBySlug(ctx.Context(), slug)
	if err != nil {
		return ctx.Status(http.StatusNotFound).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Taxonomy not found",
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.NewTaxonomyResource(taxonomy))
}

// GetTaxonomies godoc
//
//	@Summary		Get all taxonomies
//	@Description	Retrieve all taxonomies with pagination
//	@Tags			taxonomies
//	@Accept			json
//	@Produce		json
//	@Param			limit	query		int													false	"Number of results to return (default: 10)"	default(10)
//	@Param			offset	query		int													false	"Number of results to skip (default: 0)"	default(0)
//	@Success		200		{object}	responses.TaxonomyCollectionResponse	"Taxonomies found"
//	@Failure		400		{object}	responses.ErrorResponse								"Bad request - Invalid pagination parameters"
//	@Failure		500		{object}	responses.ErrorResponse								"Internal server error"
//	@Router			/api/v1/taxonomies [get]
//	@Security		BearerAuth
func (c *TaxonomyController) GetTaxonomies(ctx *fiber.Ctx) error {
	limitStr := ctx.Query("limit", "10")
	offsetStr := ctx.Query("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid limit parameter",
		})
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid offset parameter",
		})
	}

	taxonomies, err := c.taxonomyService.GetAllTaxonomies(ctx.Context(), limit, offset)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	// Calculate pagination parameters
	page := (offset / limit) + 1
	if offset == 0 {
		page = 1
	}

	// Get base URL for pagination links
	baseURL := ctx.OriginalURL()

	// For now, use simple count. In real implementation, get total count
	total := int64(len(taxonomies))

	return ctx.Status(http.StatusOK).JSON(responses.NewPaginatedTaxonomyCollectionResponse(
		taxonomies, page, limit, total, baseURL,
	))
}

// GetRootTaxonomies godoc
//
//	@Summary		Get root taxonomies
//	@Description	Retrieve all root taxonomies (taxonomies without parent)
//	@Tags			taxonomies
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	responses.TaxonomyCollectionResponse	"Root taxonomies found"
//	@Failure		500	{object}	responses.ErrorResponse								"Internal server error"
//	@Router			/api/v1/taxonomies/root [get]
//	@Security		BearerAuth
func (c *TaxonomyController) GetRootTaxonomies(ctx *fiber.Ctx) error {
	taxonomies, err := c.taxonomyService.GetRootTaxonomies(ctx.Context())
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.NewTaxonomyCollectionResponse(taxonomies))
}

// GetTaxonomyHierarchy godoc
//
//	@Summary		Get taxonomy hierarchy
//	@Description	Retrieve the complete taxonomy hierarchy tree
//	@Tags			taxonomies
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	responses.TaxonomyCollectionResponse	"Taxonomy hierarchy found"
//	@Failure		500	{object}	responses.ErrorResponse								"Internal server error"
//	@Router			/api/v1/taxonomies/hierarchy [get]
//	@Security		BearerAuth
func (c *TaxonomyController) GetTaxonomyHierarchy(ctx *fiber.Ctx) error {
	taxonomies, err := c.taxonomyService.GetTaxonomyHierarchy(ctx.Context())
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.NewTaxonomyCollectionResponse(taxonomies))
}

// GetTaxonomyChildren godoc
//
//	@Summary		Get taxonomy children
//	@Description	Retrieve direct children of a taxonomy
//	@Tags			taxonomies
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string												true	"Taxonomy ID"	format(uuid)
//	@Success		200	{object}	responses.SuccessResponse{data=[]entities.Taxonomy}	"Taxonomy children found"
//	@Failure		400	{object}	responses.ErrorResponse								"Bad request - Invalid taxonomy ID format"
//	@Failure		500	{object}	responses.ErrorResponse								"Internal server error"
//	@Router			/api/v1/taxonomies/{id}/children [get]
//	@Security		BearerAuth
func (c *TaxonomyController) GetTaxonomyChildren(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid taxonomy ID format",
		})
	}

	taxonomies, err := c.taxonomyService.GetTaxonomyChildren(ctx.Context(), id)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.NewTaxonomyCollectionResponse(taxonomies))
}

// GetTaxonomyDescendants godoc
//
//	@Summary		Get taxonomy descendants
//	@Description	Retrieve all descendants of a taxonomy (children, grandchildren, etc.)
//	@Tags			taxonomies
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string												true	"Taxonomy ID"	format(uuid)
//	@Success		200	{object}	responses.SuccessResponse{data=[]entities.Taxonomy}	"Taxonomy descendants found"
//	@Failure		400	{object}	responses.ErrorResponse								"Bad request - Invalid taxonomy ID format"
//	@Failure		500	{object}	responses.ErrorResponse								"Internal server error"
//	@Router			/api/v1/taxonomies/{id}/descendants [get]
//	@Security		BearerAuth
func (c *TaxonomyController) GetTaxonomyDescendants(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid taxonomy ID format",
		})
	}

	taxonomies, err := c.taxonomyService.GetTaxonomyDescendants(ctx.Context(), id)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.NewTaxonomyCollectionResponse(taxonomies))
}

// GetTaxonomyAncestors godoc
//
//	@Summary		Get taxonomy ancestors
//	@Description	Retrieve all ancestors of a taxonomy (parent, grandparent, etc.)
//	@Tags			taxonomies
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string												true	"Taxonomy ID"	format(uuid)
//	@Success		200	{object}	responses.SuccessResponse{data=[]entities.Taxonomy}	"Taxonomy ancestors found"
//	@Failure		400	{object}	responses.ErrorResponse								"Bad request - Invalid taxonomy ID format"
//	@Failure		500	{object}	responses.ErrorResponse								"Internal server error"
//	@Router			/api/v1/taxonomies/{id}/ancestors [get]
//	@Security		BearerAuth
func (c *TaxonomyController) GetTaxonomyAncestors(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid taxonomy ID format",
		})
	}

	taxonomies, err := c.taxonomyService.GetTaxonomyAncestors(ctx.Context(), id)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.NewTaxonomyCollectionResponse(taxonomies))
}

// GetTaxonomySiblings godoc
//
//	@Summary		Get taxonomy siblings
//	@Description	Retrieve all siblings of a taxonomy (taxonomies with the same parent)
//	@Tags			taxonomies
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string												true	"Taxonomy ID"	format(uuid)
//	@Success		200	{object}	responses.SuccessResponse{data=[]entities.Taxonomy}	"Taxonomy siblings found"
//	@Failure		400	{object}	responses.ErrorResponse								"Bad request - Invalid taxonomy ID format"
//	@Failure		500	{object}	responses.ErrorResponse								"Internal server error"
//	@Router			/api/v1/taxonomies/{id}/siblings [get]
//	@Security		BearerAuth
func (c *TaxonomyController) GetTaxonomySiblings(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid taxonomy ID format",
		})
	}

	taxonomies, err := c.taxonomyService.GetTaxonomySiblings(ctx.Context(), id)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.NewTaxonomyCollectionResponse(taxonomies))
}

// SearchTaxonomies godoc
//
//	@Summary		Search taxonomies
//	@Description	Search taxonomies by name, slug, or description with pagination
//	@Tags			taxonomies
//	@Accept			json
//	@Produce		json
//	@Param			q		query		string												true	"Search query"
//	@Param			limit	query		int													false	"Number of results to return (default: 10)"	default(10)
//	@Param			offset	query		int													false	"Number of results to skip (default: 0)"	default(0)
//	@Success		200		{object}	responses.TaxonomyCollectionResponse	"Taxonomies found"
//	@Failure		400		{object}	responses.ErrorResponse								"Bad request - Search query is required"
//	@Failure		500		{object}	responses.ErrorResponse								"Internal server error"
//	@Router			/api/v1/taxonomies/search [get]
//	@Security		BearerAuth
func (c *TaxonomyController) SearchTaxonomies(ctx *fiber.Ctx) error {
	query := ctx.Query("q", "")
	if query == "" {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Search query is required",
		})
	}

	limitStr := ctx.Query("limit", "10")
	offsetStr := ctx.Query("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid limit parameter",
		})
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid offset parameter",
		})
	}

	taxonomies, err := c.taxonomyService.SearchTaxonomies(ctx.Context(), query, limit, offset)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	// Calculate pagination parameters
	page := (offset / limit) + 1
	if offset == 0 {
		page = 1
	}

	// Get base URL for pagination links
	baseURL := ctx.OriginalURL()

	// For now, use simple count. In real implementation, get total count
	total := int64(len(taxonomies))

	return ctx.Status(http.StatusOK).JSON(responses.NewPaginatedTaxonomyCollectionResponse(
		taxonomies, page, limit, total, baseURL,
	))
}

// SearchTaxonomiesWithPagination godoc
//
//	@Summary		Search taxonomies with pagination
//	@Description	Search taxonomies with advanced pagination and sorting options
//	@Tags			taxonomies
//	@Accept			json
//	@Produce		json
//	@Param			query		query		string																false	"Search query"
//	@Param			page		query		int																	false	"Page number (default: 1)"					default(1)
//	@Param			per_page	query		int																	false	"Items per page (default: 10, max: 100)"	default(10)
//	@Param			sort_by		query		string																false	"Sort field (default: record_left)"			default(record_left)
//	@Param			sort_desc	query		bool																false	"Sort descending (default: false)"			default(false)
//	@Success		200			{object}	responses.SuccessResponse{data=pagination.TaxonomySearchResponse}	"Taxonomies with pagination"
//	@Failure		400			{object}	responses.ErrorResponse												"Bad request"
//	@Failure		500			{object}	responses.ErrorResponse												"Internal server error"
//	@Router			/api/v1/taxonomies/search/advanced [get]
//	@Security		BearerAuth
func (c *TaxonomyController) SearchTaxonomiesWithPagination(ctx *fiber.Ctx) error {
	// Parse query parameters
	query := ctx.Query("query", "")
	pageStr := ctx.Query("page", "1")
	perPageStr := ctx.Query("per_page", "10")
	sortBy := ctx.Query("sort_by", "record_left")
	sortDescStr := ctx.Query("sort_desc", "false")

	// Parse pagination parameters
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	perPage, err := strconv.Atoi(perPageStr)
	if err != nil || perPage < 1 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}

	// Parse sort direction
	sortDesc := sortDescStr == "true" || sortDescStr == "1"

	// Create search request
	searchRequest := &pagination.TaxonomySearchRequest{
		Query:    query,
		Page:     page,
		PerPage:  perPage,
		SortBy:   sortBy,
		SortDesc: sortDesc,
	}

	// Call service method
	response, err := c.taxonomyService.SearchTaxonomiesWithPagination(ctx.Context(), searchRequest)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Failed to search taxonomies: " + err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(responses.SuccessResponse{
		Status:  "success",
		Message: "Taxonomies retrieved successfully",
		Data:    response,
	})
}

// UpdateTaxonomy godoc
//
//	@Summary		Update a taxonomy
//	@Description	Update an existing taxonomy with new information
//	@Tags			taxonomies
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string								true	"Taxonomy ID"	format(uuid)
//	@Param			request	body		requests.UpdateTaxonomyRequest						true	"Taxonomy update request"
//	@Success		200		{object}	responses.SuccessResponse{data=entities.Taxonomy}	"Taxonomy updated successfully"
//	@Failure		422		{object}	responses.ErrorResponse								"Validation errors"
//	@Failure		500		{object}	responses.ErrorResponse								"Internal server error"
//	@Router			/api/v1/taxonomies/{id} [put]
//	@Security		BearerAuth
func (c *TaxonomyController) UpdateTaxonomy(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid taxonomy ID format",
		})
	}

	var req requests.UpdateTaxonomyRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid request body",
		})
	}

	if err := req.Validate(); err != nil {
		return ctx.Status(http.StatusUnprocessableEntity).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	// Get existing taxonomy
	existingTaxonomy, err := c.taxonomyService.GetTaxonomyByID(ctx.Context(), id)
	if err != nil {
		return ctx.Status(http.StatusNotFound).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Taxonomy not found",
		})
	}

	// Transform request to entity
	updatedTaxonomy, err := req.ToEntity(existingTaxonomy)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	// Update taxonomy using the entity
	taxonomy, err := c.taxonomyService.UpdateTaxonomy(ctx.Context(), updatedTaxonomy)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   taxonomy,
	})
}

// DeleteTaxonomy godoc
//
//	@Summary		Delete a taxonomy
//	@Description	Delete a taxonomy by its ID
//	@Tags			taxonomies
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string						true	"Taxonomy ID"	format(uuid)
//	@Success		200	{object}	responses.SuccessResponse	"Taxonomy deleted successfully"
//	@Failure		400	{object}	responses.ErrorResponse		"Bad request - Invalid taxonomy ID format"
//	@Failure		500	{object}	responses.ErrorResponse		"Internal server error"
//	@Router			/api/v1/taxonomies/{id} [delete]
//	@Security		BearerAuth
func (c *TaxonomyController) DeleteTaxonomy(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid taxonomy ID format",
		})
	}

	err = c.taxonomyService.DeleteTaxonomy(ctx.Context(), id)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status:  "success",
		Message: "Taxonomy deleted successfully",
	})
}
