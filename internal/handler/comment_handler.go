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

type CommentHandler struct {
	BaseHandler
	comments CommentService
}

type CommentService interface {
	Create(ctx context.Context, postID uint, userID uint, req request.CreateCommentRequest) (*model.Comment, error)
	List(ctx context.Context, req request.CommentListRequest) (repository.CursorPage, error)
}

func NewCommentHandler(comments CommentService, log *zap.Logger) *CommentHandler {
	return &CommentHandler{BaseHandler: BaseHandler{Log: log}, comments: comments}
}

// CreateComment godoc
// @Summary      Add a comment to a post
// @Tags         Comments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int              true  "Post ID"
// @Param        body  body      request.CreateCommentRequest  true  "Create comment payload"
// @Success      201   {object}  response.Envelope
// @Failure      400   {object}  response.Envelope
// @Failure      401   {object}  response.Envelope
// @Failure      404   {object}  response.Envelope
// @Failure      500   {object}  response.Envelope
// @Router       /api/v1/posts/{id}/comments [post]
func (h *CommentHandler) Create(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok {
		response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeComments, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}

	postID, err := h.ParseUintParam(c, "id")
	if err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeComments, response.CaseCodeInvalidValue), "invalid post id", "id must be uint")
		return
	}

	var req request.CreateCommentRequest
	if !h.bindJSON(c, response.ServiceCodeComments, &req) {
		return
	}
	if !h.validate(c, response.ServiceCodeComments, req) {
		return
	}

	cmt, err := h.comments.Create(c.Request.Context(), postID, auth.UserID, req)
	if err != nil {
		switch err {
		case service.ErrPostMissing:
			response.NotFound(c, response.BuildResponseCode(http.StatusNotFound, response.ServiceCodeComments, response.CaseCodeNotFound), "not found", "post not found")
		default:
			h.internalError(c, response.ServiceCodeComments, err, "create comment failed")
		}
		return
	}

	response.Created(c, response.BuildResponseCode(201, response.ServiceCodeComments, response.CaseCodeCreated), "created", cmt)
}

// ListComments godoc
// @Summary      List comments for a post
// @Tags         Comments
// @Produce      json
// @Param        id     path      int  true   "Post ID"
// @Param        limit  query     int  false  "Max comments (max 200)"
// @Success      200    {object}  response.OKPaginated
// @Failure      400    {object}  response.Envelope
// @Failure      500    {object}  response.Envelope
// @Router       /api/v1/posts/{id}/comments [get]
func (h *CommentHandler) List(c *gin.Context) {
	postID, err := h.ParseUintParam(c, "id")
	if err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeComments, response.CaseCodeInvalidValue), "invalid post id", "id must be uint")
		return
	}

	var req request.CommentListRequest
	req.PostID = postID
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeComments, response.CaseCodeInvalidFormat), "invalid request", err.Error())
		return
	}
	// Ensure path param always wins over any query value.
	req.PostID = postID
	if !h.validate(c, response.ServiceCodeComments, req) {
		return
	}

	page, err := h.comments.List(c.Request.Context(), req)
	if err != nil {
		h.internalError(c, response.ServiceCodeComments, err, "list failed")
		return
	}

	response.OKPaginated(
		c,
		response.BuildResponseCode(200, response.ServiceCodeComments, response.CaseCodeListRetrieved),
		"ok",
		page.Items,
		page.NextCursor != nil,
		page.PrevCursor != nil,
	)
}
