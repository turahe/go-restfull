package controllers

import (
	"net/http"
	"strconv"

	"webapi/internal/application/ports"
	"webapi/internal/helper/pagination"
	"webapi/internal/interfaces/http/requests"
	"webapi/internal/interfaces/http/responses"

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

func (c *TaxonomyController) CreateTaxonomy(ctx *fiber.Ctx) error {
	var req requests.CreateTaxonomyRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid request body",
		})
	}

	if err := req.Validate(); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	var parentID *uuid.UUID
	if req.ParentID != "" {
		parsedID, err := uuid.Parse(req.ParentID)
		if err != nil {
			return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
				Status:  "error",
				Message: "Invalid parent ID format",
			})
		}
		parentID = &parsedID
	}

	taxonomy, err := c.taxonomyService.CreateTaxonomy(
		ctx.Context(),
		req.Name,
		req.Slug,
		req.Code,
		req.Description,
		parentID,
	)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusCreated).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   taxonomy,
	})
}

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

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   taxonomy,
	})
}

func (c *TaxonomyController) GetTaxonomyBySlug(ctx *fiber.Ctx) error {
	slug := ctx.Params("slug")
	if slug == "" {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Slug is required",
		})
	}

	taxonomy, err := c.taxonomyService.GetTaxonomyBySlug(ctx.Context(), slug)
	if err != nil {
		return ctx.Status(http.StatusNotFound).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Taxonomy not found",
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   taxonomy,
	})
}

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

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   taxonomies,
	})
}

func (c *TaxonomyController) GetRootTaxonomies(ctx *fiber.Ctx) error {
	taxonomies, err := c.taxonomyService.GetRootTaxonomies(ctx.Context())
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   taxonomies,
	})
}

func (c *TaxonomyController) GetTaxonomyHierarchy(ctx *fiber.Ctx) error {
	taxonomies, err := c.taxonomyService.GetTaxonomyHierarchy(ctx.Context())
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   taxonomies,
	})
}

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

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   taxonomies,
	})
}

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

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   taxonomies,
	})
}

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

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   taxonomies,
	})
}

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

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   taxonomies,
	})
}

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

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   taxonomies,
	})
}

// SearchTaxonomiesWithPagination handles GET /v1/taxonomies/search requests with unified response
// @Summary Search taxonomies with pagination
// @Description Search taxonomies with pagination and return unified response
// @Tags taxonomies
// @Accept json
// @Produce json
// @Param query query string false "Search query"
// @Param page query int false "Page number (default: 1)"
// @Param per_page query int false "Items per page (default: 10, max: 100)"
// @Param sort_by query string false "Sort field (default: record_left)"
// @Param sort_desc query bool false "Sort descending (default: false)"
// @Success 200 {object} pagination.TaxonomySearchResponse "Taxonomies with pagination"
// @Failure 400 {object} responses.ErrorResponse "Bad request"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /taxonomies/search [get]
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
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	var parentID *uuid.UUID
	if req.ParentID != "" {
		parsedID, err := uuid.Parse(req.ParentID)
		if err != nil {
			return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
				Status:  "error",
				Message: "Invalid parent ID format",
			})
		}
		parentID = &parsedID
	}

	taxonomy, err := c.taxonomyService.UpdateTaxonomy(
		ctx.Context(),
		id,
		req.Name,
		req.Slug,
		req.Code,
		req.Description,
		parentID,
	)
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
