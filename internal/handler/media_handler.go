package handler

import (
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-rest/internal/middleware"
	"go-rest/internal/model"
	"go-rest/internal/service"
	"go-rest/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type MediaHandler struct {
	BaseHandler
	mediaSvc MediaService
}

type MediaService interface {
	Upload(ctx context.Context, actorUserID uint, mediaableType string, mediaableID *uint, fh *multipart.FileHeader) (*model.Media, error)
	List(ctx context.Context, actorUserID uint, limit int) ([]model.Media, error)
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

	mediaableTypeRaw := strings.ToLower(strings.TrimSpace(c.PostForm("mediaableType")))
	mediaableIDRaw := strings.TrimSpace(c.PostForm("mediaableId"))

	var mediaableType string
	var mediaableID *uint

	if mediaableTypeRaw != "" || mediaableIDRaw != "" {
		if mediaableTypeRaw == "" || mediaableIDRaw == "" {
			response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeMedia, response.CaseCodeInvalidValue), "invalid request", "mediaableType and mediaableId are required together")
			return
		}

		switch mediaableTypeRaw {
		case "user":
			mediaableType = "User"
		case "post":
			mediaableType = "Post"
		case "category":
			mediaableType = "Category"
		case "comment":
			mediaableType = "Comment"
		default:
			response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeMedia, response.CaseCodeInvalidValue), "invalid request", "invalid mediaableType")
			return
		}

		n, err := strconv.ParseUint(mediaableIDRaw, 10, 64)
		if err != nil || n == 0 {
			response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeMedia, response.CaseCodeInvalidValue), "invalid request", "mediaableId must be uint > 0")
			return
		}
		u := uint(n)
		mediaableID = &u
	}

	m, err := h.mediaSvc.Upload(c.Request.Context(), auth.UserID, mediaableType, mediaableID, fh)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMediaTooLarge):
			response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeMedia, response.CaseCodeInvalidValue), "invalid request", "file too large")
		default:
			h.internalError(c, response.ServiceCodeMedia, err, "upload failed")
		}
		return
	}

	response.Created(c, response.BuildResponseCode(http.StatusCreated, response.ServiceCodeMedia, response.CaseCodeCreated), "uploaded", m)
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

	limit := 50
	if s := strings.TrimSpace(c.Query("limit")); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil {
			response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeMedia, response.CaseCodeInvalidValue), "invalid limit", "limit must be int")
			return
		}
		limit = n
	}

	rows, err := h.mediaSvc.List(c.Request.Context(), auth.UserID, limit)
	if err != nil {
		h.internalError(c, response.ServiceCodeMedia, err, "list failed")
		return
	}

	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeMedia, response.CaseCodeListRetrieved), "ok", rows)
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

	id, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeMedia, response.CaseCodeInvalidValue), "invalid id", "id must be uint")
		return
	}

	m, err := h.mediaSvc.GetByID(c.Request.Context(), auth.UserID, uint(id))
	if err != nil {
		if errors.Is(err, service.ErrMediaNotFound) {
			response.NotFound(c, response.BuildResponseCode(http.StatusNotFound, response.ServiceCodeMedia, response.CaseCodeNotFound), "not found", "media not found")
			return
		}
		h.internalError(c, response.ServiceCodeMedia, err, "get failed")
		return
	}

	// Best-effort: return a temporary download URL when MinIO is enabled.
	if m != nil {
		if url, _ := h.mediaSvc.PresignGet(c.Request.Context(), m.StoragePath, 15*time.Minute); url != "" {
			m.DownloadURL = url
		}
	}

	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeMedia, response.CaseCodeRetrieved), "ok", m)
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

	id, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeMedia, response.CaseCodeInvalidValue), "invalid id", "id must be uint")
		return
	}

	if err := h.mediaSvc.Delete(c.Request.Context(), auth.UserID, uint(id)); err != nil {
		if errors.Is(err, service.ErrMediaNotFound) {
			response.NotFound(c, response.BuildResponseCode(http.StatusNotFound, response.ServiceCodeMedia, response.CaseCodeNotFound), "not found", "media not found")
			return
		}
		h.internalError(c, response.ServiceCodeMedia, err, "delete failed")
		return
	}

	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeMedia, response.CaseCodeDeleted), "deleted", gin.H{"id": uint(id)})
}

