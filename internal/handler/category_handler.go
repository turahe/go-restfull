package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/turahe/go-restfull/internal/handler/request"
	"github.com/turahe/go-restfull/internal/middleware"
	"github.com/turahe/go-restfull/internal/model"
	"github.com/turahe/go-restfull/internal/usecase"
	"github.com/turahe/go-restfull/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// categoryUsecase is satisfied by *usecase.CategoryUsecase (tests may use a mock).
type categoryUsecase interface {
	CreateRoot(ctx context.Context, name string, actorUserID uint) (*model.CategoryModel, error)
	CreateChild(ctx context.Context, parentID uint, name string, actorUserID uint) (*model.CategoryModel, error)
	Update(ctx context.Context, id uint, name string, actorUserID uint) (*model.CategoryModel, error)
	Delete(ctx context.Context, id uint, actorUserID uint) error
	GetTree(ctx context.Context) ([]usecase.CategoryTreeNode, error)
	GetSubtree(ctx context.Context, categoryID uint) ([]usecase.CategoryTreeNode, error)
}

type CategoryHandler struct {
	BaseHandler
	uc categoryUsecase
}

func NewCategoryHandler(uc categoryUsecase, log *zap.Logger) *CategoryHandler {
	return &CategoryHandler{BaseHandler: BaseHandler{Log: log}, uc: uc}
}

// CreateRoot godoc
// @Summary      Create root category
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      request.CreateCategoryRootBody  true  "Root category name"
// @Success      201   {object}  response.Envelope
// @Failure      400   {object}  response.Envelope
// @Failure      401   {object}  response.Envelope
// @Failure      500   {object}  response.Envelope
// @Router       /api/v1/categories/root [post]
func (h *CategoryHandler) CreateRoot(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok {
		response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeCategories, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}
	var req request.CreateCategoryRootBody
	if !h.bindJSON(c, response.ServiceCodeCategories, &req) {
		return
	}
	if !h.validate(c, response.ServiceCodeCategories, req) {
		return
	}
	cat, err := h.uc.CreateRoot(c.Request.Context(), req.Name, auth.UserID)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidName):
			response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeCategories, response.CaseCodeInvalidValue), "invalid name", err.Error())
		case errors.Is(err, usecase.ErrCategoryDuplicateName):
			response.Conflict(c, response.BuildResponseCode(http.StatusConflict, response.ServiceCodeCategories, response.CaseCodeConflict), "duplicate name", err.Error())
		default:
			h.internalError(c, response.ServiceCodeCategories, err, "create root failed")
		}
		return
	}
	response.Created(c, response.BuildResponseCode(http.StatusCreated, response.ServiceCodeCategories, response.CaseCodeCreated), "created", cat)
}

// CreateChild godoc
// @Summary      Create child category
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int  true  "Parent category id"
// @Param        body  body      request.CreateCategoryChildBody  true  "Child category name"
// @Success      201   {object}  response.Envelope
// @Failure      400   {object}  response.Envelope
// @Failure      401   {object}  response.Envelope
// @Failure      404   {object}  response.Envelope
// @Failure      500   {object}  response.Envelope
// @Router       /api/v1/categories/{id}/child [post]
func (h *CategoryHandler) CreateChild(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok {
		response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeCategories, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}
	parentID, err := h.ParseUintParam(c, "id")
	if err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeCategories, response.CaseCodeInvalidValue), "invalid id", "id must be uint")
		return
	}
	var req request.CreateCategoryChildBody
	if !h.bindJSON(c, response.ServiceCodeCategories, &req) {
		return
	}
	if !h.validate(c, response.ServiceCodeCategories, req) {
		return
	}
	cat, err := h.uc.CreateChild(c.Request.Context(), uint(parentID), req.Name, auth.UserID)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidName):
			response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeCategories, response.CaseCodeInvalidValue), "invalid name", err.Error())
		case errors.Is(err, usecase.ErrCategoryNotFound):
			response.NotFound(c, response.BuildResponseCode(http.StatusNotFound, response.ServiceCodeCategories, response.CaseCodeNotFound), "not found", "parent category not found")
		case errors.Is(err, usecase.ErrCategoryDuplicateName):
			response.Conflict(c, response.BuildResponseCode(http.StatusConflict, response.ServiceCodeCategories, response.CaseCodeConflict), "duplicate name", err.Error())
		default:
			h.internalError(c, response.ServiceCodeCategories, err, "create child failed")
		}
		return
	}
	response.Created(c, response.BuildResponseCode(http.StatusCreated, response.ServiceCodeCategories, response.CaseCodeCreated), "created", cat)
}

// GetTree godoc
// @Summary      Get category tree
// @Tags         Categories
// @Produce      json
// @Success      200  {object}  response.Envelope
// @Failure      500  {object}  response.Envelope
// @Router       /api/v1/categories/tree [get]
func (h *CategoryHandler) GetTree(c *gin.Context) {
	tree, err := h.uc.GetTree(c.Request.Context())
	if err != nil {
		h.internalError(c, response.ServiceCodeCategories, err, "get tree failed")
		return
	}
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeCategories, response.CaseCodeListRetrieved), "ok", tree)
}

// GetSubtree godoc
// @Summary      Get category subtree
// @Tags         Categories
// @Produce      json
// @Param        id  path      int  true  "Category id (root of subtree)"
// @Success      200  {object}  response.Envelope
// @Failure      400  {object}  response.Envelope
// @Failure      404  {object}  response.Envelope
// @Failure      500  {object}  response.Envelope
// @Router       /api/v1/categories/{id}/subtree [get]
func (h *CategoryHandler) GetSubtree(c *gin.Context) {
	id, err := h.ParseUintParam(c, "id")
	if err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeCategories, response.CaseCodeInvalidValue), "invalid id", "id must be uint")
		return
	}
	sub, err := h.uc.GetSubtree(c.Request.Context(), uint(id))
	if err != nil {
		switch err {
		case usecase.ErrCategoryNotFound:
			response.NotFound(c, response.BuildResponseCode(http.StatusNotFound, response.ServiceCodeCategories, response.CaseCodeNotFound), "not found", "category not found")
		default:
			h.internalError(c, response.ServiceCodeCategories, err, "get subtree failed")
		}
		return
	}
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeCategories, response.CaseCodeRetrieved), "ok", sub)
}

// Update godoc
// @Summary      Update category name
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int                      true  "Category id"
// @Param        body  body      request.UpdateCategoryBody  true  "New name"
// @Success      200   {object}  response.Envelope
// @Failure      400   {object}  response.Envelope
// @Failure      401   {object}  response.Envelope
// @Failure      404   {object}  response.Envelope
// @Failure      500   {object}  response.Envelope
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
	var req request.UpdateCategoryBody
	if !h.bindJSON(c, response.ServiceCodeCategories, &req) {
		return
	}
	if !h.validate(c, response.ServiceCodeCategories, req) {
		return
	}
	cat, err := h.uc.Update(c.Request.Context(), uint(id), req.Name, auth.UserID)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidName):
			response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeCategories, response.CaseCodeInvalidValue), "invalid name", err.Error())
		case errors.Is(err, usecase.ErrCategoryNotFound):
			response.NotFound(c, response.BuildResponseCode(http.StatusNotFound, response.ServiceCodeCategories, response.CaseCodeNotFound), "not found", "category not found")
		case errors.Is(err, usecase.ErrCategoryDuplicateName):
			response.Conflict(c, response.BuildResponseCode(http.StatusConflict, response.ServiceCodeCategories, response.CaseCodeConflict), "duplicate name", err.Error())
		default:
			h.internalError(c, response.ServiceCodeCategories, err, "update failed")
		}
		return
	}
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeCategories, response.CaseCodeUpdated), "updated", cat)
}

// Delete godoc
// @Summary      Delete category subtree
// @Description  Soft-deletes the category and all descendants (sets deleted_at / deleted_by) and rebalances nested-set indices on active rows. Blocked if any post references a category in this subtree.
// @Tags         Categories
// @Produce      json
// @Security     BearerAuth
// @Param        id  path      int  true  "Category id (root of subtree to remove)"
// @Success      200  {object}  response.Envelope
// @Failure      400  {object}  response.Envelope
// @Failure      401  {object}  response.Envelope
// @Failure      404  {object}  response.Envelope
// @Failure      409  {object}  response.Envelope
// @Failure      500  {object}  response.Envelope
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
	err = h.uc.Delete(c.Request.Context(), uint(id), auth.UserID)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrCategoryNotFound):
			response.NotFound(c, response.BuildResponseCode(http.StatusNotFound, response.ServiceCodeCategories, response.CaseCodeNotFound), "not found", "category not found")
		case errors.Is(err, usecase.ErrCategoryDeleteHasPosts):
			response.Conflict(c, response.BuildResponseCode(http.StatusConflict, response.ServiceCodeCategories, response.CaseCodeConflict), "conflict", err.Error())
		default:
			h.internalError(c, response.ServiceCodeCategories, err, "delete failed")
		}
		return
	}
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeCategories, response.CaseCodeDeleted), "deleted", gin.H{"id": uint(id)})
}
