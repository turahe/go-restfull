package handler

import (
	"context"
	"net/http"

	"github.com/turahe/go-restfull/internal/handler/request"
	"github.com/turahe/go-restfull/internal/middleware"
	"github.com/turahe/go-restfull/internal/model"
	"github.com/turahe/go-restfull/internal/repository"
	"github.com/turahe/go-restfull/internal/service"
	"github.com/turahe/go-restfull/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PostHandler struct {
	BaseHandler
	posts PostService
}

type PostService interface {
	List(ctx context.Context, req request.PostListRequest) (repository.CursorPage, error)
	GetBySlug(ctx context.Context, slug string) (*model.Post, error)
	Create(ctx context.Context, userID uint, req request.CreatePostRequest) (*model.Post, error)
	Update(ctx context.Context, id uint, actorUserID uint, req request.UpdatePostRequest) (*model.Post, error)
	Delete(ctx context.Context, id uint, actorUserID uint) error
}

func NewPostHandler(posts PostService, log *zap.Logger) *PostHandler {
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
	var req request.PostListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodePosts, response.CaseCodeInvalidFormat), "invalid request", err.Error())
		return
	}
	if !h.validate(c, response.ServiceCodePosts, req) {
		return
	}

	page, err := h.posts.List(c.Request.Context(), req)
	if err != nil {
		h.internalError(c, response.ServiceCodePosts, err, "list failed")
		return
	}

	next := page.NextCursor != nil
	prev := page.PrevCursor != nil

	response.OKPaginated(c,
		response.BuildResponseCode(http.StatusOK, response.ServiceCodePosts, response.CaseCodeListRetrieved),
		"Successfully retrieved posts",
		page.Items,
		next,
		prev,
	)
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
			response.NotFound(c, response.BuildResponseCode(http.StatusNotFound, response.ServiceCodePosts, response.CaseCodeNotFound), "not found", "post not found")
			return
		}
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodePosts, response.CaseCodeInvalidFormat), "invalid request", err.Error())
		return
	}
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodePosts, response.CaseCodeRetrieved), "Successfully retrieved post by slug", p)
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
		response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodePosts, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}

	var req request.CreatePostRequest
	if !h.bindJSON(c, response.ServiceCodePosts, &req) {
		return
	}
	if !h.validate(c, response.ServiceCodePosts, req) {
		return
	}

	p, err := h.posts.Create(c.Request.Context(), auth.UserID, req)
	if err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodePosts, response.CaseCodeInvalidValue), "invalid request", err.Error())
		return
	}
	response.Created(c, response.BuildResponseCode(http.StatusCreated, response.ServiceCodePosts, response.CaseCodeCreated), "Successfully created post", p)
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
		response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodePosts, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}

	id, err := h.ParseUintParam(c, "id")
	if err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodePosts, response.CaseCodeInvalidValue), "invalid id", "id must be uint")
		return
	}

	var req request.UpdatePostRequest
	if !h.bindJSON(c, response.ServiceCodePosts, &req) {
		return
	}
	if !h.validate(c, response.ServiceCodePosts, req) {
		return
	}

	p, err := h.posts.Update(c.Request.Context(), id, auth.UserID, req)
	if err != nil {
		switch err {
		case service.ErrPostNotFound:
			response.NotFound(c, response.BuildResponseCode(http.StatusNotFound, response.ServiceCodePosts, response.CaseCodeNotFound), "not found", "post not found")
		case service.ErrNotPostOwner:
			response.Forbidden(c, response.BuildResponseCode(http.StatusForbidden, response.ServiceCodePosts, response.CaseCodePermissionDenied), "forbidden", "owner only")
		default:
			h.internalError(c, response.ServiceCodePosts, err, "update failed")
		}
		return
	}

	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodePosts, response.CaseCodeUpdated), "Successfully updated post", p)
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
		response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodePosts, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}

	id, err := h.ParseUintParam(c, "id")
	if err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodePosts, response.CaseCodeInvalidValue), "invalid id", "id must be uint")
		return
	}

	err = h.posts.Delete(c.Request.Context(), id, auth.UserID)
	if err != nil {
		switch err {
		case service.ErrPostNotFound:
			response.NotFound(c, response.BuildResponseCode(http.StatusNotFound, response.ServiceCodePosts, response.CaseCodeNotFound), "not found", "post not found")
		case service.ErrNotPostOwner:
			response.Forbidden(c, response.BuildResponseCode(http.StatusForbidden, response.ServiceCodePosts, response.CaseCodePermissionDenied), "forbidden", "owner only")
		default:
			h.internalError(c, response.ServiceCodePosts, err, "delete failed")
		}
		return
	}

	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodePosts, response.CaseCodeDeleted), "Successfully deleted post", gin.H{"id": id})
}
