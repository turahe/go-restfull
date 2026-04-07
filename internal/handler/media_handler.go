package handler

import (
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/turahe/go-restfull/internal/handler/request"
	"github.com/turahe/go-restfull/internal/middleware"
	"github.com/turahe/go-restfull/internal/model"
	"github.com/turahe/go-restfull/internal/repository"
	"github.com/turahe/go-restfull/internal/usecase"
	"github.com/turahe/go-restfull/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type MediaHandler struct {
	BaseHandler
	mediaSvc mediaUsecase
}

// mediaUsecase is satisfied by *usecase.MediaUsecase (tests may use a mock).
type mediaUsecase interface {
	Upload(ctx context.Context, actorUserID uint, fh *multipart.FileHeader, parentID *uint) (*model.Media, error)
	CreateFolderRoot(ctx context.Context, actorUserID uint, name string) (*model.Media, error)
	CreateFolderChild(ctx context.Context, actorUserID uint, parentID uint, name string) (*model.Media, error)
	GetTree(ctx context.Context, actorUserID uint) ([]usecase.MediaTreeNode, error)
	GetSubtree(ctx context.Context, actorUserID uint, mediaID uint) ([]usecase.MediaTreeNode, error)
	Update(ctx context.Context, actorUserID uint, id uint, name string) (*model.Media, error)
	List(ctx context.Context, actorUserID uint, req request.MediaListRequest) (repository.CursorPage, error)
	GetByID(ctx context.Context, actorUserID, id uint) (*model.Media, error)
	Delete(ctx context.Context, actorUserID, id uint) error
	PresignGet(ctx context.Context, objectKey string, expiry time.Duration) (string, error)
}

func NewMediaHandler(mediaSvc mediaUsecase, log *zap.Logger) *MediaHandler {
	return &MediaHandler{BaseHandler: BaseHandler{Log: log}, mediaSvc: mediaSvc}
}

// UploadMedia godoc
// @Summary      Upload media (image/file)
// @Tags         Media
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file      formData  file  true  "Upload file"
// @Param        parentId  formData  int   false  "Optional parent folder media id"
// @Success      201   {object}  response.Envelope
// @Failure      400   {object}  response.Envelope
// @Failure      401   {object}  response.Envelope
// @Failure      500   {object}  response.Envelope
// @Router       /api/v1/media [post]
func (h *MediaHandler) UploadMedia(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok {
		response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeMedia, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}

	fh, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeMedia, response.CaseCodeInvalidValue), "invalid request", "missing file field")
		return
	}

	var parentID *uint
	if s := strings.TrimSpace(c.PostForm("parentId")); s != "" {
		v, perr := strconv.ParseUint(s, 10, 32)
		if perr != nil || v == 0 {
			response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeMedia, response.CaseCodeInvalidValue), "invalid request", "parentId must be a positive uint")
			return
		}
		tmp := uint(v)
		parentID = &tmp
	}

	m, err := h.mediaSvc.Upload(c.Request.Context(), auth.UserID, fh, parentID)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrMediaTooLarge):
			response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeMedia, response.CaseCodeInvalidValue), "invalid request", "file too large")
		case errors.Is(err, usecase.ErrMediaParentNotFound):
			response.NotFound(c, response.BuildResponseCode(http.StatusNotFound, response.ServiceCodeMedia, response.CaseCodeNotFound), "not found", "parent media not found")
		default:
			h.internalError(c, response.ServiceCodeMedia, err, "upload failed")
		}
		return
	}

	response.Created(c, response.BuildResponseCode(http.StatusCreated, response.ServiceCodeMedia, response.CaseCodeCreated), "Successfully uploaded media", m)
}

// CreateFolderRoot godoc
// @Summary      Create root folder
// @Tags         Media
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      request.CreateMediaFolderRootBody  true  "Folder name"
// @Success      201   {object}  response.Envelope
// @Failure      400   {object}  response.Envelope
// @Failure      401   {object}  response.Envelope
// @Failure      409   {object}  response.Envelope
// @Failure      500   {object}  response.Envelope
// @Router       /api/v1/media/root [post]
func (h *MediaHandler) CreateFolderRoot(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok {
		response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeMedia, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}
	var req request.CreateMediaFolderRootBody
	if !h.bindJSON(c, response.ServiceCodeMedia, &req) {
		return
	}
	if !h.validate(c, response.ServiceCodeMedia, req) {
		return
	}
	m, err := h.mediaSvc.CreateFolderRoot(c.Request.Context(), auth.UserID, req.Name)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrMediaDuplicateName):
			response.Conflict(c, response.BuildResponseCode(http.StatusConflict, response.ServiceCodeMedia, response.CaseCodeConflict), "duplicate name", err.Error())
		default:
			h.internalError(c, response.ServiceCodeMedia, err, "create folder failed")
		}
		return
	}
	response.Created(c, response.BuildResponseCode(http.StatusCreated, response.ServiceCodeMedia, response.CaseCodeCreated), "created", m)
}

// CreateFolderChild godoc
// @Summary      Create child folder
// @Tags         Media
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int  true  "Parent media id"
// @Param        body  body      request.CreateMediaFolderChildBody  true  "Folder name"
// @Success      201   {object}  response.Envelope
// @Failure      400   {object}  response.Envelope
// @Failure      401   {object}  response.Envelope
// @Failure      404   {object}  response.Envelope
// @Failure      409   {object}  response.Envelope
// @Failure      500   {object}  response.Envelope
// @Router       /api/v1/media/{id}/child [post]
func (h *MediaHandler) CreateFolderChild(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok {
		response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeMedia, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}
	parentID, err := h.ParseUintParam(c, "id")
	if err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeMedia, response.CaseCodeInvalidValue), "invalid id", "id must be uint")
		return
	}
	var req request.CreateMediaFolderChildBody
	if !h.bindJSON(c, response.ServiceCodeMedia, &req) {
		return
	}
	if !h.validate(c, response.ServiceCodeMedia, req) {
		return
	}
	m, err := h.mediaSvc.CreateFolderChild(c.Request.Context(), auth.UserID, parentID, req.Name)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrMediaParentNotFound):
			response.NotFound(c, response.BuildResponseCode(http.StatusNotFound, response.ServiceCodeMedia, response.CaseCodeNotFound), "not found", "parent media not found")
		case errors.Is(err, usecase.ErrMediaDuplicateName):
			response.Conflict(c, response.BuildResponseCode(http.StatusConflict, response.ServiceCodeMedia, response.CaseCodeConflict), "duplicate name", err.Error())
		default:
			h.internalError(c, response.ServiceCodeMedia, err, "create folder failed")
		}
		return
	}
	response.Created(c, response.BuildResponseCode(http.StatusCreated, response.ServiceCodeMedia, response.CaseCodeCreated), "created", m)
}

// GetTree godoc
// @Summary      Media tree
// @Tags         Media
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Envelope
// @Failure      401  {object}  response.Envelope
// @Failure      500  {object}  response.Envelope
// @Router       /api/v1/media/tree [get]
func (h *MediaHandler) GetTree(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok {
		response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeMedia, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}
	tree, err := h.mediaSvc.GetTree(c.Request.Context(), auth.UserID)
	if err != nil {
		h.internalError(c, response.ServiceCodeMedia, err, "get tree failed")
		return
	}
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeMedia, response.CaseCodeListRetrieved), "ok", tree)
}

// GetSubtree godoc
// @Summary      Media subtree
// @Tags         Media
// @Produce      json
// @Security     BearerAuth
// @Param        id  path      int  true  "Media id (root of subtree)"
// @Success      200  {object}  response.Envelope
// @Failure      401  {object}  response.Envelope
// @Failure      404  {object}  response.Envelope
// @Failure      500  {object}  response.Envelope
// @Router       /api/v1/media/{id}/subtree [get]
func (h *MediaHandler) GetSubtree(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok {
		response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeMedia, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}
	id, err := h.ParseUintParam(c, "id")
	if err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeMedia, response.CaseCodeInvalidValue), "invalid id", "id must be uint")
		return
	}
	sub, err := h.mediaSvc.GetSubtree(c.Request.Context(), auth.UserID, id)
	if err != nil {
		if errors.Is(err, usecase.ErrMediaNotFound) {
			response.NotFound(c, response.BuildResponseCode(http.StatusNotFound, response.ServiceCodeMedia, response.CaseCodeNotFound), "not found", "media not found")
			return
		}
		h.internalError(c, response.ServiceCodeMedia, err, "get subtree failed")
		return
	}
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeMedia, response.CaseCodeRetrieved), "ok", sub)
}

// UpdateMedia godoc
// @Summary      Rename media (folder or file) in tree
// @Tags         Media
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int  true  "Media id"
// @Param        body  body      request.UpdateMediaBody  true  "New name"
// @Success      200   {object}  response.Envelope
// @Failure      400   {object}  response.Envelope
// @Failure      401   {object}  response.Envelope
// @Failure      404   {object}  response.Envelope
// @Failure      409   {object}  response.Envelope
// @Failure      500   {object}  response.Envelope
// @Router       /api/v1/media/{id} [put]
func (h *MediaHandler) UpdateMedia(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok {
		response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeMedia, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}
	id, err := h.ParseUintParam(c, "id")
	if err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeMedia, response.CaseCodeInvalidValue), "invalid id", "id must be uint")
		return
	}
	var req request.UpdateMediaBody
	if !h.bindJSON(c, response.ServiceCodeMedia, &req) {
		return
	}
	if !h.validate(c, response.ServiceCodeMedia, req) {
		return
	}
	m, err := h.mediaSvc.Update(c.Request.Context(), auth.UserID, id, req.Name)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrMediaNotFound):
			response.NotFound(c, response.BuildResponseCode(http.StatusNotFound, response.ServiceCodeMedia, response.CaseCodeNotFound), "not found", "media not found")
		case errors.Is(err, usecase.ErrMediaDuplicateName):
			response.Conflict(c, response.BuildResponseCode(http.StatusConflict, response.ServiceCodeMedia, response.CaseCodeConflict), "duplicate name", err.Error())
		default:
			h.internalError(c, response.ServiceCodeMedia, err, "update failed")
		}
		return
	}
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeMedia, response.CaseCodeUpdated), "updated", m)
}

// ListMedia godoc
// @Summary      List my media
// @Tags         Media
// @Produce      json
// @Security     BearerAuth
// @Param        limit  query     int  false  "Max items (max 500)"
// @Success      200    {object}  response.Envelope
// @Failure      401    {object}  response.Envelope
// @Failure      500    {object}  response.Envelope
// @Router       /api/v1/media [get]
func (h *MediaHandler) ListMedia(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok {
		response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeMedia, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}
	var req request.MediaListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeMedia, response.CaseCodeInvalidFormat), "invalid request", err.Error())
		return
	}

	page, err := h.mediaSvc.List(c.Request.Context(), auth.UserID, req)
	if err != nil {
		h.internalError(c, response.ServiceCodeMedia, err, "list failed")
		return
	}

	next := page.NextCursor != nil
	prev := page.PrevCursor != nil
	response.OKPaginated(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeMedia, response.CaseCodeListRetrieved), "Successfully retrieved media list", page.Items, next, prev)
}

// GetMediaByID godoc
// @Summary      Get my media by id
// @Tags         Media
// @Produce      json
// @Security     BearerAuth
// @Param        id  path      int  true  "Media ID"
// @Success      200 {object}  response.Envelope
// @Failure      400 {object}  response.Envelope
// @Failure      401 {object}  response.Envelope
// @Failure      404 {object}  response.Envelope
// @Failure      500 {object}  response.Envelope
// @Router       /api/v1/media/{id} [get]
func (h *MediaHandler) GetMediaByID(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok {
		response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeMedia, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}

	id, err := h.ParseUintParam(c, "id")
	if err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeMedia, response.CaseCodeInvalidValue), "invalid id", "id must be uint")
		return
	}

	m, err := h.mediaSvc.GetByID(c.Request.Context(), auth.UserID, id)
	if err != nil {
		if errors.Is(err, usecase.ErrMediaNotFound) {
			response.NotFound(c, response.BuildResponseCode(http.StatusNotFound, response.ServiceCodeMedia, response.CaseCodeNotFound), "not found", "media not found")
			return
		}
		h.internalError(c, response.ServiceCodeMedia, err, "get failed")
		return
	}

	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeMedia, response.CaseCodeRetrieved), "Successfully retrieved media", m)
}

// DeleteMedia godoc
// @Summary      Delete my media subtree
// @Tags         Media
// @Produce      json
// @Security     BearerAuth
// @Param        id  path      int  true  "Media ID"
// @Success      200 {object}  response.Envelope
// @Failure      400 {object}  response.Envelope
// @Failure      401 {object}  response.Envelope
// @Failure      404 {object}  response.Envelope
// @Failure      500 {object}  response.Envelope
// @Router       /api/v1/media/{id} [delete]
func (h *MediaHandler) DeleteMedia(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok {
		response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeMedia, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}

	id, err := h.ParseUintParam(c, "id")
	if err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeMedia, response.CaseCodeInvalidValue), "invalid id", err.Error())
		return
	}
	if err := h.mediaSvc.Delete(c.Request.Context(), auth.UserID, id); err != nil {
		if errors.Is(err, usecase.ErrMediaNotFound) {
			response.NotFound(c, response.BuildResponseCode(http.StatusNotFound, response.ServiceCodeMedia, response.CaseCodeNotFound), "not found", "media not found")
			return
		}
		h.internalError(c, response.ServiceCodeMedia, err, "delete failed")
		return
	}

	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeMedia, response.CaseCodeDeleted), "Successfully deleted media", gin.H{"id": uint(id)})
}
