package taxonomy

import (
	"net/http"
	"time"

	"webapi/internal/db/model"
	"webapi/internal/http/response"
	"webapi/internal/http/validation"
	"webapi/internal/repository"
	"webapi/pkg/exception"

	"math"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TaxonomyHandler struct {
	Repo repository.TaxonomyRepository
}

func NewTaxonomyHandler(repo repository.TaxonomyRepository) *TaxonomyHandler {
	return &TaxonomyHandler{Repo: repo}
}

// @Summary Create taxonomy
// @Description Create a new taxonomy
// @Tags taxonomies
// @Accept json
// @Produce json
// @Param taxonomy body model.Taxonomy true "Taxonomy info"
// @Success 201 {object} response.CommonResponse
// @Failure 400 {object} response.CommonResponse
// @Router /v1/taxonomies [post]
func (h *TaxonomyHandler) CreateTaxonomy(c *fiber.Ctx) error {
	var req model.Taxonomy
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(response.CommonResponse{ResponseCode: http.StatusBadRequest, ResponseMessage: "Invalid request", Errors: nil})
	}
	v, _ := validation.GetValidator()
	if err := v.Struct(req); err != nil {
		if validationErrs, ok := err.(validator.ValidationErrors); ok {
			return c.Status(http.StatusUnprocessableEntity).JSON(response.CommonResponse{
				ResponseCode:    http.StatusUnprocessableEntity,
				ResponseMessage: "Validation failed",
				Errors:          exception.NewValidationFailedErrors(validationErrs),
			})
		}
		return c.Status(http.StatusBadRequest).JSON(response.CommonResponse{ResponseCode: http.StatusBadRequest, ResponseMessage: err.Error()})
	}
	if req.ID == uuid.Nil {
		req.ID = uuid.New()
	}
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()
	if err := h.Repo.Create(c.Context(), &req); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(response.CommonResponse{ResponseCode: http.StatusInternalServerError, ResponseMessage: err.Error()})
	}
	return c.Status(http.StatusCreated).JSON(response.CommonResponse{ResponseCode: http.StatusCreated, ResponseMessage: "Created", Data: req})
}

// @Summary Get taxonomy by ID
// @Description Get a taxonomy by its ID
// @Tags taxonomies
// @Produce json
// @Param id path string true "Taxonomy UUID"
// @Success 200 {object} response.CommonResponse
// @Failure 404 {object} response.CommonResponse
// @Router /v1/taxonomies/{id} [get]
func (h *TaxonomyHandler) GetTaxonomyByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(response.CommonResponse{ResponseCode: http.StatusBadRequest, ResponseMessage: "Invalid ID"})
	}
	tax, err := h.Repo.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(response.CommonResponse{ResponseCode: http.StatusNotFound, ResponseMessage: err.Error()})
	}
	return c.Status(http.StatusOK).JSON(response.CommonResponse{ResponseCode: http.StatusOK, ResponseMessage: "OK", Data: tax})
}

// @Summary Get all taxonomies
// @Description Get all taxonomies
// @Tags taxonomies
// @Produce json
// @Success 200 {object} response.CommonResponse
// @Router /v1/taxonomies [get]
func (h *TaxonomyHandler) GetAllTaxonomies(c *fiber.Ctx) error {
	query := c.Query("query", "")
	limit := c.QueryInt("limit", 10)
	page := c.QueryInt("page", 1)

	taxonomies, total, err := h.Repo.GetPaginated(c.Context(), query, limit, page)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(response.CommonResponse{ResponseCode: http.StatusInternalServerError, ResponseMessage: err.Error()})
	}

	totalPage := int(math.Ceil(float64(total) / float64(limit)))
	lastPage := totalPage
	nextPage := page + 1
	if nextPage > totalPage {
		nextPage = totalPage
	}
	prevPage := page - 1
	if prevPage < 1 {
		prevPage = 1
	}

	paginated := response.PaginationResponse{
		Data:         taxonomies,
		TotalCount:   total,
		TotalPage:    totalPage,
		CurrentPage:  page,
		LastPage:     lastPage,
		PerPage:      limit,
		NextPage:     nextPage,
		PreviousPage: prevPage,
		Path:         c.OriginalURL(),
	}

	return c.Status(http.StatusOK).JSON(response.CommonResponse{
		ResponseCode:    http.StatusOK,
		ResponseMessage: "OK",
		Data:            paginated,
	})
}

// @Summary Update taxonomy
// @Description Update a taxonomy by its ID
// @Tags taxonomies
// @Accept json
// @Produce json
// @Param id path string true "Taxonomy UUID"
// @Param taxonomy body model.Taxonomy true "Taxonomy info"
// @Success 200 {object} response.CommonResponse
// @Failure 400 {object} response.CommonResponse
// @Router /v1/taxonomies/{id} [put]
func (h *TaxonomyHandler) UpdateTaxonomy(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(response.CommonResponse{ResponseCode: http.StatusBadRequest, ResponseMessage: "Invalid ID"})
	}
	var req model.Taxonomy
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(response.CommonResponse{ResponseCode: http.StatusBadRequest, ResponseMessage: "Invalid request"})
	}
	v, _ := validation.GetValidator()
	if err := v.Struct(req); err != nil {
		if validationErrs, ok := err.(validator.ValidationErrors); ok {
			return c.Status(http.StatusUnprocessableEntity).JSON(response.CommonResponse{
				ResponseCode:    http.StatusUnprocessableEntity,
				ResponseMessage: "Validation failed",
				Errors:          exception.NewValidationFailedErrors(validationErrs),
			})
		}
		return c.Status(http.StatusBadRequest).JSON(response.CommonResponse{ResponseCode: http.StatusBadRequest, ResponseMessage: err.Error()})
	}
	req.ID = id
	req.UpdatedAt = time.Now()
	if err := h.Repo.Update(c.Context(), &req); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(response.CommonResponse{ResponseCode: http.StatusInternalServerError, ResponseMessage: err.Error()})
	}
	return c.Status(http.StatusOK).JSON(response.CommonResponse{ResponseCode: http.StatusOK, ResponseMessage: "Updated", Data: req})
}

// @Summary Delete taxonomy
// @Description Delete a taxonomy by its ID
// @Tags taxonomies
// @Produce json
// @Param id path string true "Taxonomy UUID"
// @Success 200 {object} response.CommonResponse
// @Failure 404 {object} response.CommonResponse
// @Router /v1/taxonomies/{id} [delete]
func (h *TaxonomyHandler) DeleteTaxonomy(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(response.CommonResponse{ResponseCode: http.StatusBadRequest, ResponseMessage: "Invalid ID"})
	}
	if err := h.Repo.Delete(c.Context(), id); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(response.CommonResponse{ResponseCode: http.StatusInternalServerError, ResponseMessage: err.Error()})
	}
	return c.Status(http.StatusOK).JSON(response.CommonResponse{ResponseCode: http.StatusOK, ResponseMessage: "Deleted"})
}
