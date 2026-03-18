package handler

import (
	"strconv"
	"strings"

	"go-rest/internal/middleware"
	"go-rest/internal/service"
	"go-rest/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type CommentHandler struct {
	comments *service.CommentService
	log      *zap.Logger
}

func NewCommentHandler(comments *service.CommentService, log *zap.Logger) *CommentHandler {
	return &CommentHandler{comments: comments, log: log}
}

type createCommentReq struct {
	Content string `json:"content" binding:"required,min=1,max=2000"`
}

// CreateComment godoc
// @Summary      Add a comment to a post
// @Tags         Comments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int              true  "Post ID"
// @Param        body  body      createCommentReq  true  "Create comment payload"
// @Success      201   {object}  response.Envelope
// @Failure      400   {object}  response.Envelope
// @Failure      401   {object}  response.Envelope
// @Failure      404   {object}  response.Envelope
// @Failure      500   {object}  response.Envelope
// @Router       /api/posts/{id}/comments [post]
func (h *CommentHandler) Create(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok {
		response.Unauthorized(c, 4010401, "unauthorized", "missing auth")
		return
	}

	postID, err := parseUintParam(c, "id")
	if err != nil {
		response.BadRequest(c, 4000401, "invalid post id", "id must be uint")
		return
	}

	var req createCommentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, 4000402, "invalid request", err.Error())
		return
	}

	cmt, err := h.comments.Create(c.Request.Context(), postID, auth.UserID, req.Content)
	if err != nil {
		switch err {
		case service.ErrPostMissing:
			response.NotFound(c, 4040401, "not found", "post not found")
		default:
			h.log.Error("create comment failed", zap.Error(err))
			response.BadRequest(c, 4000403, "invalid request", err.Error())
		}
		return
	}

	response.Created(c, 2010401, "created", cmt)
}

// ListComments godoc
// @Summary      List comments for a post
// @Tags         Comments
// @Produce      json
// @Param        id     path      int  true   "Post ID"
// @Param        limit  query     int  false  "Max comments (max 200)"
// @Success      200    {object}  response.Envelope
// @Failure      400    {object}  response.Envelope
// @Failure      500    {object}  response.Envelope
// @Router       /api/posts/{id}/comments [get]
func (h *CommentHandler) List(c *gin.Context) {
	postID, err := parseUintParam(c, "id")
	if err != nil {
		response.BadRequest(c, 4000404, "invalid post id", "id must be uint")
		return
	}

	limit := 50
	if s := strings.TrimSpace(c.Query("limit")); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil {
			response.BadRequest(c, 4000405, "invalid limit", "limit must be int")
			return
		}
		limit = n
	}

	rows, err := h.comments.List(c.Request.Context(), postID, limit)
	if err != nil {
		h.log.Error("list comments failed", zap.Error(err))
		response.InternalServerError(c, 5000401, "internal error", "list failed")
		return
	}

	response.OK(c, 2000401, "ok", rows)
}

