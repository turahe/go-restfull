package handler

import (
	"context"
	"errors"
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

type CommentHandler struct {
	BaseHandler
	comments commentService
}

// commentService is satisfied by *service.CommentService (tests may use a mock).
type commentService interface {
	CreateRoot(ctx context.Context, postID uint, userID uint, req request.CreateCommentRequest) (*model.Comment, error)
	CreateChild(ctx context.Context, postID uint, parentID uint, userID uint, req request.CreateCommentRequest) (*model.Comment, error)
	GetTree(ctx context.Context, postID uint) ([]service.CommentTreeNode, error)
	GetSubtree(ctx context.Context, postID uint, commentID uint) ([]service.CommentTreeNode, error)
	Update(ctx context.Context, postID uint, commentID uint, userID uint, req request.UpdateCommentBody) (*model.Comment, error)
	Delete(ctx context.Context, postID uint, commentID uint, userID uint) error
	List(ctx context.Context, req request.CommentListRequest) (repository.CursorPage, error)
}

func NewCommentHandler(comments commentService, log *zap.Logger) *CommentHandler {
	return &CommentHandler{BaseHandler: BaseHandler{Log: log}, comments: comments}
}

// CreateRoot godoc
// @Summary      Add a root comment to a post
// @Tags         Comments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int                      true  "Post ID"
// @Param        body  body      request.CreateCommentRequest  true  "Create comment payload"
// @Success      201   {object}  response.Envelope
// @Failure      400   {object}  response.Envelope
// @Failure      401   {object}  response.Envelope
// @Failure      404   {object}  response.Envelope
// @Failure      500   {object}  response.Envelope
// @Router       /api/v1/posts/{id}/comments/root [post]
func (h *CommentHandler) CreateRoot(c *gin.Context) {
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

	cmt, err := h.comments.CreateRoot(c.Request.Context(), postID, auth.UserID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCommentPostMissing):
			response.NotFound(c, response.BuildResponseCode(http.StatusNotFound, response.ServiceCodeComments, response.CaseCodeNotFound), "not found", "post not found")
		default:
			h.internalError(c, response.ServiceCodeComments, err, "create comment failed")
		}
		return
	}

	response.Created(c, response.BuildResponseCode(201, response.ServiceCodeComments, response.CaseCodeCreated), "created", cmt)
}

// CreateChild godoc
// @Summary      Reply to a comment
// @Tags         Comments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int                      true  "Post ID"
// @Param        cid   path      int                      true  "Parent comment ID"
// @Param        body  body      request.CreateCommentRequest  true  "Create comment payload"
// @Success      201   {object}  response.Envelope
// @Failure      400   {object}  response.Envelope
// @Failure      401   {object}  response.Envelope
// @Failure      404   {object}  response.Envelope
// @Failure      500   {object}  response.Envelope
// @Router       /api/v1/posts/{id}/comments/{cid}/child [post]
func (h *CommentHandler) CreateChild(c *gin.Context) {
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
	parentID, err := h.ParseUintParam(c, "cid")
	if err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeComments, response.CaseCodeInvalidValue), "invalid parent comment id", "cid must be uint")
		return
	}

	var req request.CreateCommentRequest
	if !h.bindJSON(c, response.ServiceCodeComments, &req) {
		return
	}
	if !h.validate(c, response.ServiceCodeComments, req) {
		return
	}

	cmt, err := h.comments.CreateChild(c.Request.Context(), postID, uint(parentID), auth.UserID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCommentPostMissing):
			response.NotFound(c, response.BuildResponseCode(http.StatusNotFound, response.ServiceCodeComments, response.CaseCodeNotFound), "not found", "post not found")
		case errors.Is(err, service.ErrCommentParentNotFound):
			response.NotFound(c, response.BuildResponseCode(http.StatusNotFound, response.ServiceCodeComments, response.CaseCodeNotFound), "not found", "parent comment not found")
		default:
			h.internalError(c, response.ServiceCodeComments, err, "create comment failed")
		}
		return
	}

	response.Created(c, response.BuildResponseCode(201, response.ServiceCodeComments, response.CaseCodeCreated), "created", cmt)
}

// GetTree godoc
// @Summary      Comment tree for a post
// @Tags         Comments
// @Produce      json
// @Param        id  path      int  true  "Post ID"
// @Success      200 {object}  response.Envelope
// @Failure      400 {object}  response.Envelope
// @Failure      500 {object}  response.Envelope
// @Router       /api/v1/posts/{id}/comments/tree [get]
func (h *CommentHandler) GetTree(c *gin.Context) {
	postID, err := h.ParseUintParam(c, "id")
	if err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeComments, response.CaseCodeInvalidValue), "invalid post id", "id must be uint")
		return
	}

	tree, err := h.comments.GetTree(c.Request.Context(), postID)
	if err != nil {
		h.internalError(c, response.ServiceCodeComments, err, "get tree failed")
		return
	}
	response.OK(c, response.BuildResponseCode(200, response.ServiceCodeComments, response.CaseCodeListRetrieved), "ok", tree)
}

// GetSubtree godoc
// @Summary      Comment subtree from a node
// @Tags         Comments
// @Produce      json
// @Param        id   path      int  true  "Post ID"
// @Param        cid  path      int  true  "Comment ID (root of subtree)"
// @Success      200  {object}  response.Envelope
// @Failure      400  {object}  response.Envelope
// @Failure      404  {object}  response.Envelope
// @Failure      500  {object}  response.Envelope
// @Router       /api/v1/posts/{id}/comments/{cid}/subtree [get]
func (h *CommentHandler) GetSubtree(c *gin.Context) {
	postID, err := h.ParseUintParam(c, "id")
	if err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeComments, response.CaseCodeInvalidValue), "invalid post id", "id must be uint")
		return
	}
	cid, err := h.ParseUintParam(c, "cid")
	if err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeComments, response.CaseCodeInvalidValue), "invalid comment id", "cid must be uint")
		return
	}

	sub, err := h.comments.GetSubtree(c.Request.Context(), postID, uint(cid))
	if err != nil {
		if errors.Is(err, service.ErrCommentNotFound) {
			response.NotFound(c, response.BuildResponseCode(http.StatusNotFound, response.ServiceCodeComments, response.CaseCodeNotFound), "not found", "comment not found")
			return
		}
		h.internalError(c, response.ServiceCodeComments, err, "get subtree failed")
		return
	}
	response.OK(c, response.BuildResponseCode(200, response.ServiceCodeComments, response.CaseCodeRetrieved), "ok", sub)
}

// Update godoc
// @Summary      Update comment text
// @Tags         Comments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int                      true  "Post ID"
// @Param        cid   path      int                      true  "Comment ID"
// @Param        body  body      request.UpdateCommentBody  true  "New content"
// @Success      200   {object}  response.Envelope
// @Failure      400   {object}  response.Envelope
// @Failure      401   {object}  response.Envelope
// @Failure      404   {object}  response.Envelope
// @Failure      500   {object}  response.Envelope
// @Router       /api/v1/posts/{id}/comments/{cid} [put]
func (h *CommentHandler) Update(c *gin.Context) {
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
	cid, err := h.ParseUintParam(c, "cid")
	if err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeComments, response.CaseCodeInvalidValue), "invalid comment id", "cid must be uint")
		return
	}

	var req request.UpdateCommentBody
	if !h.bindJSON(c, response.ServiceCodeComments, &req) {
		return
	}
	if !h.validate(c, response.ServiceCodeComments, req) {
		return
	}

	cmt, err := h.comments.Update(c.Request.Context(), postID, uint(cid), auth.UserID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCommentNotFound):
			response.NotFound(c, response.BuildResponseCode(http.StatusNotFound, response.ServiceCodeComments, response.CaseCodeNotFound), "not found", "comment not found")
		case errors.Is(err, service.ErrCommentInvalidContent):
			response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeComments, response.CaseCodeInvalidValue), "invalid content", err.Error())
		default:
			h.internalError(c, response.ServiceCodeComments, err, "update failed")
		}
		return
	}
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeComments, response.CaseCodeUpdated), "updated", cmt)
}

// Delete godoc
// @Summary      Delete a comment subtree
// @Tags         Comments
// @Produce      json
// @Security     BearerAuth
// @Param        id  path      int  true  "Post ID"
// @Param        cid  path      int  true  "Comment ID (root of subtree to remove)"
// @Success      200  {object}  response.Envelope
// @Failure      400  {object}  response.Envelope
// @Failure      401  {object}  response.Envelope
// @Failure      404  {object}  response.Envelope
// @Failure      409  {object}  response.Envelope
// @Failure      500  {object}  response.Envelope
// @Router       /api/v1/posts/{id}/comments/{cid} [delete]
func (h *CommentHandler) Delete(c *gin.Context) {
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
	cid, err := h.ParseUintParam(c, "cid")
	if err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeComments, response.CaseCodeInvalidValue), "invalid comment id", "cid must be uint")
		return
	}

	err = h.comments.Delete(c.Request.Context(), postID, uint(cid), auth.UserID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCommentNotFound):
			response.NotFound(c, response.BuildResponseCode(http.StatusNotFound, response.ServiceCodeComments, response.CaseCodeNotFound), "not found", "comment not found")
		case errors.Is(err, service.ErrCommentSubtreeHasMedia):
			response.Conflict(c, response.BuildResponseCode(http.StatusConflict, response.ServiceCodeComments, response.CaseCodeConflict), "conflict", err.Error())
		default:
			h.internalError(c, response.ServiceCodeComments, err, "delete failed")
		}
		return
	}
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeComments, response.CaseCodeDeleted), "deleted", gin.H{"id": uint(cid)})
}

// List godoc
// @Summary      List comments for a post
// @Tags         Comments
// @Produce      json
// @Param        id     path      int  true   "Post ID"
// @Param        limit  query     int  false  "Max comments (max 200)"
// @Success      200    {object}  response.Envelope
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
