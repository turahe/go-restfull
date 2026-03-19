package handler

import (
	"context"
	"net/http"

	"go-rest/internal/handler/request"
	"go-rest/internal/middleware"
	"go-rest/internal/model"
	"go-rest/internal/repository"
	"go-rest/internal/service"
	"go-rest/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type CategoryHandler struct {
	BaseHandler
	categories CategoryService
}

type CategoryService interface {
	List(ctx context.Context, req request.CategoryListRequest) (repository.CursorPage, error)
	GetBySlug(ctx context.Context, slug string) (*model.Category, error)
	Create(ctx context.Context, actorUserID uint, req request.CreateCategoryRequest) (*model.Category, error)
	Update(ctx context.Context, id uint, actorUserID uint, req request.UpdateCategoryRequest) (*model.Category, error)
	Delete(ctx context.Context, id uint, actorUserID uint) error
}

func NewCategoryHandler(categories CategoryService, log *zap.Logger) *CategoryHandler {
	return &CategoryHandler{BaseHandler: BaseHandler{Log: log}, categories: categories}
}

// ListCategories godoc
// @Summary      List categories
// @Tags         Categories
// @Produce      json
// @Param        limit  query     int  false  "Max items (max 200)"
// @Success      200    {object}  response.OKPaginated
// @Failure      500    {object}  response.Envelope
// @Router       /api/v1/categories [get]
func (h *CategoryHandler) List(c *gin.Context) {
	var req request.CategoryListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c,
			response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeCategories, response.CaseCodeInvalidFormat),
			"invalid limit",
			err.Error(),
		)
		return
	}
	if !h.validate(c, response.ServiceCodeCategories, req) {
		return
	}

	page, err := h.categories.List(c.Request.Context(), req)
	if err != nil {
		h.internalError(c, response.ServiceCodeCategories, err, "list failed")
		return
	}
	next := page.NextCursor != nil
	prev := page.PrevCursor != nil
	response.OKPaginated(
		c,
		response.BuildResponseCode(http.StatusOK, response.ServiceCodeCategories, response.CaseCodeListRetrieved),
		"ok",
		page.Items,
		next,
		prev,
	)
}

// GetCategoryBySlug godoc
// @Summary      Get category by slug
// @Tags         Categories
// @Produce      json
// @Param        slug  path      string  true  "Category slug"
// @Success      200   {object}  response.OK
// @Failure      400   {object}  response.BadRequest
// @Failure      404   {object}  response.NotFound
// @Router       /api/v1/categories/{slug} [get]
func (h *CategoryHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	cat, err := h.categories.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		switch err {
		case service.ErrCategoryNotFound:
			response.NotFound(c, response.BuildResponseCode(http.StatusNotFound, response.ServiceCodeCategories, response.CaseCodeNotFound), "not found", "category not found")
		default:
			response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeCategories, response.CaseCodeInvalidFormat), "invalid request", err.Error())
		}
		return
	}
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeCategories, response.CaseCodeRetrieved), "ok", cat)
}

// CreateCategory godoc
// @Summary      Create a category
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      request.CreateCategoryRequest  true  "Create category payload"
// @Success      201   {object}  response.Created
// @Failure      400   {object}  response.BadRequest
// @Failure      401   {object}  response.Unauthorized
// @Failure      500   {object}  response.InternalServerError
// @Router       /api/v1/categories [post]
func (h *CategoryHandler) Create(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok {
		response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeCategories, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}
	var req request.CreateCategoryRequest
	if !h.bindJSON(c, response.ServiceCodeCategories, &req) {
		return
	}
	if !h.validate(c, response.ServiceCodeCategories, req) {
		return
	}

	cat, err := h.categories.Create(c.Request.Context(), auth.UserID, req)
	if err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeCategories, response.CaseCodeInvalidValue), "invalid request", err.Error())
		return
	}
	response.Created(c, response.BuildResponseCode(http.StatusCreated, response.ServiceCodeCategories, response.CaseCodeCreated), "Successfully created category", cat)
}

// UpdateCategory godoc
// @Summary      Update a category
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int                         true  "Category ID"
// @Param        body  body      request.UpdateCategoryRequest true  "Update category payload"
// @Success      200   {object}  response.OK
// @Failure      400   {object}  response.BadRequest
// @Failure      401   {object}  response.Unauthorized
// @Failure      404   {object}  response.NotFound
// @Failure      500   {object}  response.InternalServerError
// @Router       /api/v1/categories/{id} [put]
func (h *CategoryHandler) Update(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok {
		response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeCategories, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}
	id, err := h.ParseUintParam(c, "id")
	if err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeCategories, response.CaseCodeInvalidValue), "invalid id", "id must be uint")
		return
	}
	var req request.UpdateCategoryRequest
	if !h.bindJSON(c, response.ServiceCodeCategories, &req) {
		return
	}
	if !h.validate(c, response.ServiceCodeCategories, req) {
		return
	}

	cat, err := h.categories.Update(c.Request.Context(), id, auth.UserID, req)
	if err != nil {
		switch err {
		case service.ErrCategoryNotFound:
			response.NotFound(c, response.BuildResponseCode(http.StatusNotFound, response.ServiceCodeCategories, response.CaseCodeNotFound), "not found", "category not found")
		default:
			h.internalError(c, response.ServiceCodeCategories, err, "update failed")
		}
		return
	}
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeCategories, response.CaseCodeUpdated), "updated", cat)
}

// DeleteCategory godoc
// @Summary      Delete a category
// @Tags         Categories
// @Produce      json
// @Security     BearerAuth
// @Param        id  path      int  true  "Category ID"
// @Success      200 {object}  response.OK
// @Failure      400 {object}  response.BadRequest
// @Failure      401 {object}  response.Unauthorized
// @Failure      404 {object}  response.NotFound
// @Failure      500 {object}  response.InternalServerError
// @Router       /api/v1/categories/{id} [delete]
func (h *CategoryHandler) Delete(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok {
		response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeCategories, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}
	id, err := h.ParseUintParam(c, "id")
	if err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeCategories, response.CaseCodeInvalidValue), "invalid id", "id must be uint")
		return
	}

	err = h.categories.Delete(c.Request.Context(), id, auth.UserID)
	if err != nil {
		switch err {
		case service.ErrCategoryNotFound:
			response.NotFound(c, response.BuildResponseCode(http.StatusNotFound, response.ServiceCodeCategories, response.CaseCodeNotFound), "not found", "category not found")
		default:
			h.internalError(c, response.ServiceCodeCategories, err, "delete failed")
		}
		return
	}
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeCategories, response.CaseCodeDeleted), "deleted", gin.H{"id": uint(id)})
}
