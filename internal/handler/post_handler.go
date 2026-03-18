package handler

import (
	"strconv"
	"strings"

	"go-rest/internal/handler/request"
	"go-rest/internal/middleware"
	"go-rest/internal/repository"
	"go-rest/internal/service"
	"go-rest/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PostHandler struct {
	BaseHandler
	posts *service.PostService
}

func NewPostHandler(posts *service.PostService, log *zap.Logger) *PostHandler {
	return &PostHandler{BaseHandler: BaseHandler{Log: log}, posts: posts}
}

// ListPosts godoc
// @Summary      List posts (cursor pagination)
// @Tags         Posts
// @Produce      json
// @Param        limit   query     int     false  "Page size (max 50)"  minimum(1)  maximum(50)
// @Param        cursor  query     string  false  "Cursor (post id)"
// @Param        dir     query     string  false  "Direction: next|prev"  Enums(next,prev)
// @Success      200     {object}  response.Envelope
// @Failure      400     {object}  response.Envelope
// @Failure      500     {object}  response.Envelope
// @Router       /api/v1/posts [get]
func (h *PostHandler) List(c *gin.Context) {
	limit := parseIntDefault(c.Query("limit"), 10)
	if limit > 50 {
		limit = 50
	}
	if limit <= 0 {
		limit = 10
	}

	var cursor *uint
	if s := strings.TrimSpace(c.Query("cursor")); s != "" {
		v, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			response.BadRequest(c, response.BuildResponseCode(400, response.ServiceCodePosts, response.CaseCodeInvalidValue), "invalid cursor", "cursor must be uint")
			return
		}
		tmp := uint(v)
		cursor = &tmp
	}

	dir := repository.CursorDirection(strings.TrimSpace(strings.ToLower(c.Query("dir"))))
	if dir == "" {
		dir = repository.CursorNext
	}

	page, err := h.posts.List(c.Request.Context(), cursor, limit, dir)
	if err != nil {
		h.internalError(c, response.ServiceCodePosts, err, "list failed")
		return
	}

	var next *string
	var prev *string
	if page.NextCursor != nil {
		s := strconv.FormatUint(uint64(*page.NextCursor), 10)
		next = &s
	}
	if page.PrevCursor != nil {
		s := strconv.FormatUint(uint64(*page.PrevCursor), 10)
		prev = &s
	}

	response.OKCursor(c, response.BuildResponseCode(200, response.ServiceCodePosts, response.CaseCodeListRetrieved), "ok", page.Items, next, prev)
}

// GetPostBySlug godoc
// @Summary      Get post by slug
// @Tags         Posts
// @Produce      json
// @Param        slug  path      string  true  "Post slug"
// @Success      200   {object}  response.Envelope
// @Failure      400   {object}  response.Envelope
// @Failure      404   {object}  response.Envelope
// @Router       /api/v1/posts/slug/{slug} [get]
func (h *PostHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	p, err := h.posts.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		if err == service.ErrPostNotFound {
			response.NotFound(c, response.BuildResponseCode(404, response.ServiceCodePosts, response.CaseCodeNotFound), "not found", "post not found")
			return
		}
		response.BadRequest(c, response.BuildResponseCode(400, response.ServiceCodePosts, response.CaseCodeInvalidFormat), "invalid request", err.Error())
		return
	}
	response.OK(c, response.BuildResponseCode(200, response.ServiceCodePosts, response.CaseCodeRetrieved), "ok", p)
}

// CreatePost godoc
// @Summary      Create a post
// @Tags         Posts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      request.CreatePostRequest  true  "Create post payload"
// @Success      201   {object}  response.Envelope
// @Failure      400   {object}  response.Envelope
// @Failure      401   {object}  response.Envelope
// @Router       /api/v1/posts [post]
func (h *PostHandler) Create(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok {
		response.Unauthorized(c, response.BuildResponseCode(401, response.ServiceCodePosts, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}

	var req request.CreatePostRequest
	if !h.bindJSON(c, response.ServiceCodePosts, &req) {
		return
	}
	if !h.validate(c, response.ServiceCodePosts, req) {
		return
	}

	p, err := h.posts.Create(c.Request.Context(), auth.UserID, req.Title, req.Content)
	if err != nil {
		response.BadRequest(c, response.BuildResponseCode(400, response.ServiceCodePosts, response.CaseCodeInvalidValue), "invalid request", err.Error())
		return
	}
	response.Created(c, response.BuildResponseCode(201, response.ServiceCodePosts, response.CaseCodeCreated), "created", p)
}

// UpdatePost godoc
// @Summary      Update a post (owner only)
// @Tags         Posts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int           true  "Post ID"
// @Param        body  body      request.UpdatePostRequest  true  "Update post payload"
// @Success      200   {object}  response.Envelope
// @Failure      400   {object}  response.Envelope
// @Failure      401   {object}  response.Envelope
// @Failure      403   {object}  response.Envelope
// @Failure      404   {object}  response.Envelope
// @Failure      500   {object}  response.Envelope
// @Router       /api/v1/posts/{id} [put]
func (h *PostHandler) Update(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok {
		response.Unauthorized(c, response.BuildResponseCode(401, response.ServiceCodePosts, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}

	id, err := parseUintParam(c, "id")
	if err != nil {
		response.BadRequest(c, response.BuildResponseCode(400, response.ServiceCodePosts, response.CaseCodeInvalidValue), "invalid id", "id must be uint")
		return
	}

	var req request.UpdatePostRequest
	if !h.bindJSON(c, response.ServiceCodePosts, &req) {
		return
	}
	if !h.validate(c, response.ServiceCodePosts, req) {
		return
	}

	p, err := h.posts.Update(c.Request.Context(), id, auth.UserID, req.Title, req.Content)
	if err != nil {
		switch err {
		case service.ErrPostNotFound:
			response.NotFound(c, response.BuildResponseCode(404, response.ServiceCodePosts, response.CaseCodeNotFound), "not found", "post not found")
		case service.ErrNotPostOwner:
			response.Forbidden(c, response.BuildResponseCode(403, response.ServiceCodePosts, response.CaseCodePermissionDenied), "forbidden", "owner only")
		default:
			h.internalError(c, response.ServiceCodePosts, err, "update failed")
		}
		return
	}

	response.OK(c, response.BuildResponseCode(200, response.ServiceCodePosts, response.CaseCodeUpdated), "updated", p)
}

// DeletePost godoc
// @Summary      Delete a post (owner only)
// @Tags         Posts
// @Produce      json
// @Security     BearerAuth
// @Param        id  path      int  true  "Post ID"
// @Success      200 {object}  response.Envelope
// @Failure      400 {object}  response.Envelope
// @Failure      401 {object}  response.Envelope
// @Failure      403 {object}  response.Envelope
// @Failure      404 {object}  response.Envelope
// @Failure      500 {object}  response.Envelope
// @Router       /api/v1/posts/{id} [delete]
func (h *PostHandler) Delete(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok {
		response.Unauthorized(c, response.BuildResponseCode(401, response.ServiceCodePosts, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}

	id, err := parseUintParam(c, "id")
	if err != nil {
		response.BadRequest(c, response.BuildResponseCode(400, response.ServiceCodePosts, response.CaseCodeInvalidValue), "invalid id", "id must be uint")
		return
	}

	err = h.posts.Delete(c.Request.Context(), id, auth.UserID)
	if err != nil {
		switch err {
		case service.ErrPostNotFound:
			response.NotFound(c, response.BuildResponseCode(404, response.ServiceCodePosts, response.CaseCodeNotFound), "not found", "post not found")
		case service.ErrNotPostOwner:
			response.Forbidden(c, response.BuildResponseCode(403, response.ServiceCodePosts, response.CaseCodePermissionDenied), "forbidden", "owner only")
		default:
			h.internalError(c, response.ServiceCodePosts, err, "delete failed")
		}
		return
	}

	response.OK(c, response.BuildResponseCode(200, response.ServiceCodePosts, response.CaseCodeDeleted), "deleted", gin.H{"id": id})
}

func parseIntDefault(s string, def int) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func parseUintParam(c *gin.Context, name string) (uint, error) {
	s := strings.TrimSpace(c.Param(name))
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(n), nil
}

