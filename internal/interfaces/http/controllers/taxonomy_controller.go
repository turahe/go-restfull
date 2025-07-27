package controllers

import (
	"strconv"
	"webapi/internal/application/ports"
	"webapi/internal/http/response"

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

// GetTaxonomies handles GET /v1/taxonomies requests
func (c *TaxonomyController) GetTaxonomies(ctx *fiber.Ctx) error {
	// Get pagination parameters
	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))
	offset, _ := strconv.Atoi(ctx.Query("offset", "0"))

	taxonomies, err := c.taxonomyService.GetAllTaxonomies(ctx.Context(), limit, offset)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to retrieve taxonomies",
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(response.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Taxonomies retrieved successfully",
		Data:            taxonomies,
	})
}

// GetTaxonomyByID handles GET /v1/taxonomies/:id requests
func (c *TaxonomyController) GetTaxonomyByID(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid taxonomy ID",
			Data:            map[string]interface{}{},
		})
	}

	taxonomy, err := c.taxonomyService.GetTaxonomyByID(ctx.Context(), id)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusNotFound,
			ResponseMessage: "Taxonomy not found",
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(response.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Taxonomy retrieved successfully",
		Data:            taxonomy,
	})
}

// GetTaxonomyBySlug handles GET /v1/taxonomies/slug/:slug requests
func (c *TaxonomyController) GetTaxonomyBySlug(ctx *fiber.Ctx) error {
	slug := ctx.Params("slug")
	if slug == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Taxonomy slug is required",
			Data:            map[string]interface{}{},
		})
	}

	taxonomy, err := c.taxonomyService.GetTaxonomyBySlug(ctx.Context(), slug)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusNotFound,
			ResponseMessage: "Taxonomy not found",
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(response.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Taxonomy retrieved successfully",
		Data:            taxonomy,
	})
}

// GetRootTaxonomies handles GET /v1/taxonomies/root requests
func (c *TaxonomyController) GetRootTaxonomies(ctx *fiber.Ctx) error {
	taxonomies, err := c.taxonomyService.GetRootTaxonomies(ctx.Context())
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to retrieve root taxonomies",
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(response.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Root taxonomies retrieved successfully",
		Data:            taxonomies,
	})
}

// GetTaxonomyHierarchy handles GET /v1/taxonomies/hierarchy requests
func (c *TaxonomyController) GetTaxonomyHierarchy(ctx *fiber.Ctx) error {
	taxonomies, err := c.taxonomyService.GetTaxonomyHierarchy(ctx.Context())
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to retrieve taxonomy hierarchy",
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(response.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Taxonomy hierarchy retrieved successfully",
		Data:            taxonomies,
	})
}

// GetTaxonomyChildren handles GET /v1/taxonomies/:id/children requests
func (c *TaxonomyController) GetTaxonomyChildren(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid taxonomy ID",
			Data:            map[string]interface{}{},
		})
	}

	taxonomies, err := c.taxonomyService.GetTaxonomyChildren(ctx.Context(), id)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to retrieve taxonomy children",
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(response.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Taxonomy children retrieved successfully",
		Data:            taxonomies,
	})
}

// GetTaxonomyDescendants handles GET /v1/taxonomies/:id/descendants requests
func (c *TaxonomyController) GetTaxonomyDescendants(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid taxonomy ID",
			Data:            map[string]interface{}{},
		})
	}

	taxonomies, err := c.taxonomyService.GetTaxonomyDescendants(ctx.Context(), id)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to retrieve taxonomy descendants",
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(response.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Taxonomy descendants retrieved successfully",
		Data:            taxonomies,
	})
}

// GetTaxonomyAncestors handles GET /v1/taxonomies/:id/ancestors requests
func (c *TaxonomyController) GetTaxonomyAncestors(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid taxonomy ID",
			Data:            map[string]interface{}{},
		})
	}

	taxonomies, err := c.taxonomyService.GetTaxonomyAncestors(ctx.Context(), id)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to retrieve taxonomy ancestors",
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(response.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Taxonomy ancestors retrieved successfully",
		Data:            taxonomies,
	})
}

// GetTaxonomySiblings handles GET /v1/taxonomies/:id/siblings requests
func (c *TaxonomyController) GetTaxonomySiblings(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid taxonomy ID",
			Data:            map[string]interface{}{},
		})
	}

	taxonomies, err := c.taxonomyService.GetTaxonomySiblings(ctx.Context(), id)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to retrieve taxonomy siblings",
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(response.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Taxonomy siblings retrieved successfully",
		Data:            taxonomies,
	})
}

// SearchTaxonomies handles GET /v1/taxonomies/search requests
func (c *TaxonomyController) SearchTaxonomies(ctx *fiber.Ctx) error {
	query := ctx.Query("q")
	if query == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Search query is required",
			Data:            map[string]interface{}{},
		})
	}

	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))
	offset, _ := strconv.Atoi(ctx.Query("offset", "0"))

	taxonomies, err := c.taxonomyService.SearchTaxonomies(ctx.Context(), query, limit, offset)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to search taxonomies",
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(response.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Taxonomy search completed successfully",
		Data:            taxonomies,
	})
}

// CreateTaxonomy handles POST /v1/taxonomies requests
func (c *TaxonomyController) CreateTaxonomy(ctx *fiber.Ctx) error {
	var request struct {
		Name        string     `json:"name"`
		Slug        string     `json:"slug"`
		Code        string     `json:"code"`
		Description string     `json:"description"`
		ParentID    *uuid.UUID `json:"parent_id"`
	}

	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid request body",
			Data:            map[string]interface{}{},
		})
	}

	taxonomy, err := c.taxonomyService.CreateTaxonomy(ctx.Context(), request.Name, request.Slug, request.Code, request.Description, request.ParentID)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: err.Error(),
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(response.CommonResponse{
		ResponseCode:    fiber.StatusCreated,
		ResponseMessage: "Taxonomy created successfully",
		Data:            taxonomy,
	})
}

// UpdateTaxonomy handles PUT /v1/taxonomies/:id requests
func (c *TaxonomyController) UpdateTaxonomy(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid taxonomy ID",
			Data:            map[string]interface{}{},
		})
	}

	var request struct {
		Name        string     `json:"name"`
		Slug        string     `json:"slug"`
		Code        string     `json:"code"`
		Description string     `json:"description"`
		ParentID    *uuid.UUID `json:"parent_id"`
	}

	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid request body",
			Data:            map[string]interface{}{},
		})
	}

	taxonomy, err := c.taxonomyService.UpdateTaxonomy(ctx.Context(), id, request.Name, request.Slug, request.Code, request.Description, request.ParentID)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: err.Error(),
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(response.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Taxonomy updated successfully",
		Data:            taxonomy,
	})
}

// DeleteTaxonomy handles DELETE /v1/taxonomies/:id requests
func (c *TaxonomyController) DeleteTaxonomy(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid taxonomy ID",
			Data:            map[string]interface{}{},
		})
	}

	err = c.taxonomyService.DeleteTaxonomy(ctx.Context(), id)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(response.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: err.Error(),
			Data:            map[string]interface{}{},
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(response.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Taxonomy deleted successfully",
		Data:            map[string]interface{}{},
	})
} 