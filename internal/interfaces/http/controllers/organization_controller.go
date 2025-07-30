package controllers

import (
	"net/http"
	"strconv"

	"github.com/turahe/go-restfull/internal/application/ports"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/interfaces/http/requests"
	"github.com/turahe/go-restfull/internal/interfaces/http/responses"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type OrganizationController struct {
	organizationService ports.OrganizationService
}

func NewOrganizationController(organizationService ports.OrganizationService) *OrganizationController {
	return &OrganizationController{
		organizationService: organizationService,
	}
}

// CreateOrganization godoc
// @Summary Create a new organization
// @Description Create a new organization, optionally with a parent
// @Tags organizations
// @Accept json
// @Produce json
// @Param organization body requests.CreateOrganizationRequest true "Organization to create"
// @Success 201 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /organizations [post]
func (c *OrganizationController) CreateOrganization(ctx *fiber.Ctx) error {
	var req requests.CreateOrganizationRequest
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

	organization, err := c.organizationService.CreateOrganization(
		ctx.Context(),
		req.Name,
		req.Description,
		req.Code,
		req.Email,
		req.Phone,
		req.Address,
		req.Website,
		req.LogoURL,
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
		Data:   organization,
	})
}

// GetOrganizationByID godoc
// @Summary Get organization by ID
// @Description Get a single organization by its ID
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Router /organizations/{id} [get]
func (c *OrganizationController) GetOrganizationByID(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid organization ID format",
		})
	}

	organization, err := c.organizationService.GetOrganizationByID(ctx.Context(), id)
	if err != nil {
		return ctx.Status(http.StatusNotFound).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Organization not found",
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   organization,
	})
}

// GetAllOrganizations godoc
// @Summary Get all organizations
// @Description Get a paginated list of all organizations
// @Tags organizations
// @Accept json
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /organizations [get]
func (c *OrganizationController) GetAllOrganizations(ctx *fiber.Ctx) error {
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

	organizations, err := c.organizationService.GetAllOrganizations(ctx.Context(), limit, offset)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   organizations,
	})
}

// UpdateOrganization godoc
// @Summary Update an organization
// @Description Update an organization's details
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Param organization body requests.UpdateOrganizationRequest true "Organization update data"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /organizations/{id} [put]
func (c *OrganizationController) UpdateOrganization(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid organization ID format",
		})
	}

	var req requests.UpdateOrganizationRequest
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

	organization, err := c.organizationService.UpdateOrganization(
		ctx.Context(),
		id,
		req.Name,
		req.Description,
		req.Code,
		req.Email,
		req.Phone,
		req.Address,
		req.Website,
		req.LogoURL,
	)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   organization,
	})
}

// DeleteOrganization godoc
// @Summary Delete an organization
// @Description Delete an organization by its ID
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /organizations/{id} [delete]
func (c *OrganizationController) DeleteOrganization(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid organization ID format",
		})
	}

	err = c.organizationService.DeleteOrganization(ctx.Context(), id)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status:  "success",
		Message: "Organization deleted successfully",
	})
}

// GetRootOrganizations godoc
// @Summary Get root organizations
// @Description Get all root organizations (organizations without a parent)
// @Tags organizations
// @Accept json
// @Produce json
// @Success 200 {object} responses.SuccessResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /organizations/root [get]
func (c *OrganizationController) GetRootOrganizations(ctx *fiber.Ctx) error {
	organizations, err := c.organizationService.GetRootOrganizations(ctx.Context())
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   organizations,
	})
}

// GetOrganizationChildren godoc
// @Summary Get children of an organization
// @Description Get direct children of an organization
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /organizations/{id}/children [get]
func (c *OrganizationController) GetOrganizationChildren(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid organization ID format",
		})
	}

	organizations, err := c.organizationService.GetChildOrganizations(ctx.Context(), id)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   organizations,
	})
}

// GetOrganizationDescendants godoc
// @Summary Get descendants of an organization
// @Description Get all descendant organizations of an organization
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /organizations/{id}/descendants [get]
func (c *OrganizationController) GetOrganizationDescendants(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid organization ID format",
		})
	}

	organizations, err := c.organizationService.GetDescendantOrganizations(ctx.Context(), id)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   organizations,
	})
}

// GetOrganizationAncestors godoc
// @Summary Get ancestors of an organization
// @Description Get all ancestor organizations of an organization
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /organizations/{id}/ancestors [get]
func (c *OrganizationController) GetOrganizationAncestors(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid organization ID format",
		})
	}

	organizations, err := c.organizationService.GetAncestorOrganizations(ctx.Context(), id)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   organizations,
	})
}

// GetOrganizationSiblings godoc
// @Summary Get siblings of an organization
// @Description Get all sibling organizations of an organization
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /organizations/{id}/siblings [get]
func (c *OrganizationController) GetOrganizationSiblings(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid organization ID format",
		})
	}

	organizations, err := c.organizationService.GetSiblingOrganizations(ctx.Context(), id)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   organizations,
	})
}

// GetOrganizationPath godoc
// @Summary Get path to an organization
// @Description Get the path from the root to the specified organization
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /organizations/{id}/path [get]
func (c *OrganizationController) GetOrganizationPath(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid organization ID format",
		})
	}

	organizations, err := c.organizationService.GetOrganizationPath(ctx.Context(), id)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   organizations,
	})
}

// GetOrganizationTree godoc
// @Summary Get the full organization tree
// @Description Get the entire organization tree structure
// @Tags organizations
// @Accept json
// @Produce json
// @Success 200 {object} responses.SuccessResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /organizations/tree [get]
func (c *OrganizationController) GetOrganizationTree(ctx *fiber.Ctx) error {
	organizations, err := c.organizationService.GetOrganizationTree(ctx.Context())
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   organizations,
	})
}

// GetOrganizationSubtree godoc
// @Summary Get a subtree of an organization
// @Description Get the subtree rooted at the specified organization
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /organizations/{id}/subtree [get]
func (c *OrganizationController) GetOrganizationSubtree(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid organization ID format",
		})
	}

	organizations, err := c.organizationService.GetOrganizationSubtree(ctx.Context(), id)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   organizations,
	})
}

// AddOrganizationChild godoc
// @Summary Add a child organization
// @Description Add a new child organization to a parent
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Parent Organization ID"
// @Param organization body requests.CreateOrganizationRequest true "Child organization to create"
// @Success 201 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /organizations/{id}/children [post]
func (c *OrganizationController) AddOrganizationChild(ctx *fiber.Ctx) error {
	parentIDStr := ctx.Params("id")
	parentID, err := uuid.Parse(parentIDStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid parent organization ID format",
		})
	}

	var req requests.CreateOrganizationRequest
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

	// Create the child organization first
	childOrg, err := c.organizationService.CreateOrganization(
		ctx.Context(),
		req.Name,
		req.Description,
		req.Code,
		req.Email,
		req.Phone,
		req.Address,
		req.Website,
		req.LogoURL,
		&parentID,
	)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusCreated).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   childOrg,
	})
}

// MoveOrganizationSubtree godoc
// @Summary Move an organization subtree
// @Description Move an organization and all its descendants to a new parent
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Param move body requests.MoveOrganizationRequest true "Move request"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /organizations/{id}/move [post]
func (c *OrganizationController) MoveOrganizationSubtree(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid organization ID format",
		})
	}

	var req requests.MoveOrganizationRequest
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

	newParentID, err := uuid.Parse(req.NewParentID)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid new parent ID format",
		})
	}

	err = c.organizationService.MoveOrganization(ctx.Context(), id, newParentID)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status:  "success",
		Message: "Organization subtree moved successfully",
	})
}

// DeleteOrganizationSubtree godoc
// @Summary Delete an organization subtree
// @Description Delete an organization and all its descendants
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /organizations/{id}/subtree [delete]
func (c *OrganizationController) DeleteOrganizationSubtree(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid organization ID format",
		})
	}

	err = c.organizationService.DeleteOrganizationSubtree(ctx.Context(), id)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status:  "success",
		Message: "Organization subtree deleted successfully",
	})
}

// SetOrganizationStatus godoc
// @Summary Set organization status
// @Description Set the status of an organization
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Param status body requests.SetOrganizationStatusRequest true "Status request"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /organizations/{id}/status [put]
func (c *OrganizationController) SetOrganizationStatus(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid organization ID format",
		})
	}

	var req requests.SetOrganizationStatusRequest
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

	status := entities.OrganizationStatus(req.Status)
	err = c.organizationService.SetOrganizationStatus(ctx.Context(), id, status)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status:  "success",
		Message: "Organization status updated successfully",
	})
}

// SearchOrganizations godoc
// @Summary Search organizations
// @Description Search organizations by query string
// @Tags organizations
// @Accept json
// @Produce json
// @Param query query string false "Search query"
// @Param page query int false "Page number"
// @Param per_page query int false "Results per page"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /organizations/search [get]
func (c *OrganizationController) SearchOrganizations(ctx *fiber.Ctx) error {
	var req requests.SearchOrganizationsRequest
	if err := ctx.QueryParser(&req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid query parameters",
		})
	}

	if err := req.Validate(); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	limit := req.PerPage
	offset := (req.Page - 1) * req.PerPage

	organizations, err := c.organizationService.SearchOrganizations(
		ctx.Context(),
		req.Query,
		limit,
		offset,
	)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   organizations,
	})
}

// GetOrganizationStats godoc
// @Summary Get organization statistics
// @Description Get statistics for an organization (children and descendants count)
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /organizations/{id}/stats [get]
func (c *OrganizationController) GetOrganizationStats(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid organization ID format",
		})
	}

	childrenCount, err := c.organizationService.GetChildrenCount(ctx.Context(), id)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	descendantsCount, err := c.organizationService.GetDescendantsCount(ctx.Context(), id)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	stats := map[string]interface{}{
		"children_count":    childrenCount,
		"descendants_count": descendantsCount,
	}

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   stats,
	})
}

// ValidateOrganizationHierarchy godoc
// @Summary Validate organization hierarchy
// @Description Validate if a parent-child relationship is valid
// @Tags organizations
// @Accept json
// @Produce json
// @Param parent_id query string true "Parent Organization ID"
// @Param child_id query string true "Child Organization ID"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Router /organizations/validate-hierarchy [get]
func (c *OrganizationController) ValidateOrganizationHierarchy(ctx *fiber.Ctx) error {
	parentIDStr := ctx.Query("parent_id")
	childIDStr := ctx.Query("child_id")

	if parentIDStr == "" || childIDStr == "" {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Both parent_id and child_id are required",
		})
	}

	parentID, err := uuid.Parse(parentIDStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid parent ID format",
		})
	}

	childID, err := uuid.Parse(childIDStr)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid child ID format",
		})
	}

	err = c.organizationService.ValidateOrganizationHierarchy(ctx.Context(), parentID, childID)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status:  "success",
		Message: "Organization hierarchy is valid",
	})
}
