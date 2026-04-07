package handler

import (
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/turahe/go-restfull/internal/handler/request"
	"github.com/turahe/go-restfull/internal/middleware"
	"github.com/turahe/go-restfull/internal/model"
	"github.com/turahe/go-restfull/internal/repository"
	"github.com/turahe/go-restfull/internal/service"
	"github.com/turahe/go-restfull/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type MediaHandler struct {
	BaseHandler
	mediaSvc MediaService
}

type MediaService interface {
	Upload(ctx context.Context, actorUserID uint, fh *multipart.FileHeader) (*model.Media, error)
	List(ctx context.Context, actorUserID uint, req request.MediaListRequest) (repository.CursorPage, error)
	GetByID(ctx context.Context, actorUserID, id uint) (*model.Media, error)
	Delete(ctx context.Context, actorUserID, id uint) error
	PresignGet(ctx context.Context, objectKey string, expiry time.Duration) (string, error)
}

func NewMediaHandler(mediaSvc MediaService, log *zap.Logger) *MediaHandler {
	return &MediaHandler{BaseHandler: BaseHandler{Log: log}, mediaSvc: mediaSvc}
}

// UploadMedia godoc
// @Summary      Upload media (image/file)
// @Tags         Media
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file  formData  file  true  "Upload file"
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

	m, err := h.mediaSvc.Upload(c.Request.Context(), auth.UserID, fh)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMediaTooLarge):
			response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeMedia, response.CaseCodeInvalidValue), "invalid request", "file too large")
		default:
			h.internalError(c, response.ServiceCodeMedia, err, "upload failed")
		}
		return
	}

	response.Created(c, response.BuildResponseCode(http.StatusCreated, response.ServiceCodeMedia, response.CaseCodeCreated), "Successfully uploaded media", m)
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
		if errors.Is(err, service.ErrMediaNotFound) {
			response.NotFound(c, response.BuildResponseCode(http.StatusNotFound, response.ServiceCodeMedia, response.CaseCodeNotFound), "not found", "media not found")
			return
		}
		h.internalError(c, response.ServiceCodeMedia, err, "get failed")
		return
	}

	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeMedia, response.CaseCodeRetrieved), "Successfully retrieved media", m)
}

// DeleteMedia godoc
// @Summary      Delete my media
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
		if errors.Is(err, service.ErrMediaNotFound) {
			response.NotFound(c, response.BuildResponseCode(http.StatusNotFound, response.ServiceCodeMedia, response.CaseCodeNotFound), "not found", "media not found")
			return
		}
		h.internalError(c, response.ServiceCodeMedia, err, "delete failed")
		return
	}

	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeMedia, response.CaseCodeDeleted), "Successfully deleted media", gin.H{"id": uint(id)})
}
