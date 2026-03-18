package handler

import (
	"net/http"
	"strconv"
	"strings"

	"go-rest/internal/handler/request"
	"go-rest/internal/middleware"
	"go-rest/internal/service"
	"go-rest/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type TagHandler struct {
	BaseHandler
	tags *service.TagService
}

func NewTagHandler(tags *service.TagService, log *zap.Logger) *TagHandler {
	return &TagHandler{BaseHandler: BaseHandler{Log: log}, tags: tags}
}

// ListTags godoc
// @Summary      List tags
// @Tags         Tags
// @Produce      json
// @Param        limit  query     int  false  "Max items (max 500)"
// @Success      200    {object}  response.Envelope
// @Failure      500    {object}  response.Envelope
// @Router       /api/v1/tags [get]
func (h *TagHandler) List(c *gin.Context) {
	limit := 200
	if s := strings.TrimSpace(c.Query("limit")); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil {
			response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeTags, response.CaseCodeInvalidValue), "invalid limit", "limit must be int")
			return
		}
		limit = n
	}

	rows, err := h.tags.List(c.Request.Context(), limit)
	if err != nil {
		h.internalError(c, response.ServiceCodeTags, err, "list failed")
		return
	}
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeTags, response.CaseCodeListRetrieved), "ok", rows)
}

// GetTagBySlug godoc
// @Summary      Get tag by slug
// @Tags         Tags
// @Produce      json
// @Param        slug  path      string  true  "Tag slug"
// @Success      200   {object}  response.Envelope
// @Failure      404   {object}  response.Envelope
// @Failure      400   {object}  response.Envelope
// @Router       /api/v1/tags/{slug} [get]
func (h *TagHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	t, err := h.tags.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		switch err {
		case service.ErrTagNotFound:
			response.NotFound(c, response.BuildResponseCode(http.StatusNotFound, response.ServiceCodeTags, response.CaseCodeNotFound), "not found", "tag not found")
		default:
			response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeTags, response.CaseCodeInvalidFormat), "invalid request", err.Error())
		}
		return
	}
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeTags, response.CaseCodeRetrieved), "ok", t)
}

// CreateTag godoc
// @Summary      Create a tag
// @Tags         Tags
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      request.CreateTagRequest  true  "Create tag payload"
// @Success      201   {object}  response.Envelope
// @Failure      400   {object}  response.Envelope
// @Failure      401   {object}  response.Envelope
// @Failure      500   {object}  response.Envelope
// @Router       /api/v1/tags [post]
func (h *TagHandler) Create(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok {
		response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeTags, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}
	var req request.CreateTagRequest
	if !h.bindJSON(c, response.ServiceCodeTags, &req) {
		return
	}
	if !h.validate(c, response.ServiceCodeTags, req) {
		return
	}

	t, err := h.tags.Create(c.Request.Context(), auth.UserID, req.Name)
	if err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeTags, response.CaseCodeInvalidValue), "invalid request", err.Error())
		return
	}
	response.Created(c, response.BuildResponseCode(http.StatusCreated, response.ServiceCodeTags, response.CaseCodeCreated), "created", t)
}

// UpdateTag godoc
// @Summary      Update a tag
// @Tags         Tags
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int                       true  "Tag ID"
// @Param        body  body      request.UpdateTagRequest   true  "Update tag payload"
// @Success      200   {object}  response.Envelope
// @Failure      400   {object}  response.Envelope
// @Failure      401   {object}  response.Envelope
// @Failure      404   {object}  response.Envelope
// @Failure      500   {object}  response.Envelope
// @Router       /api/v1/tags/{id} [put]
func (h *TagHandler) Update(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok {
		response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeTags, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}
	id, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeTags, response.CaseCodeInvalidValue), "invalid id", "id must be uint")
		return
	}
	var req request.UpdateTagRequest
	if !h.bindJSON(c, response.ServiceCodeTags, &req) {
		return
	}
	if !h.validate(c, response.ServiceCodeTags, req) {
		return
	}

	t, err := h.tags.Update(c.Request.Context(), uint(id), auth.UserID, req.Name)
	if err != nil {
		switch err {
		case service.ErrTagNotFound:
			response.NotFound(c, response.BuildResponseCode(http.StatusNotFound, response.ServiceCodeTags, response.CaseCodeNotFound), "not found", "tag not found")
		default:
			h.internalError(c, response.ServiceCodeTags, err, "update failed")
		}
		return
	}
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeTags, response.CaseCodeUpdated), "updated", t)
}

// DeleteTag godoc
// @Summary      Delete a tag
// @Tags         Tags
// @Produce      json
// @Security     BearerAuth
// @Param        id  path      int  true  "Tag ID"
// @Success      200 {object}  response.Envelope
// @Failure      400 {object}  response.Envelope
// @Failure      401 {object}  response.Envelope
// @Failure      404 {object}  response.Envelope
// @Failure      500 {object}  response.Envelope
// @Router       /api/v1/tags/{id} [delete]
func (h *TagHandler) Delete(c *gin.Context) {
	_, ok := middleware.GetAuth(c)
	if !ok {
		response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeTags, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}
	id, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeTags, response.CaseCodeInvalidValue), "invalid id", "id must be uint")
		return
	}

	err = h.tags.Delete(c.Request.Context(), uint(id), 0)
	if err != nil {
		switch err {
		case service.ErrTagNotFound:
			response.NotFound(c, response.BuildResponseCode(http.StatusNotFound, response.ServiceCodeTags, response.CaseCodeNotFound), "not found", "tag not found")
		default:
			h.internalError(c, response.ServiceCodeTags, err, "delete failed")
		}
		return
	}
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeTags, response.CaseCodeDeleted), "deleted", gin.H{"id": uint(id)})
}

