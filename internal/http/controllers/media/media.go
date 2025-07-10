package media

import (
	"webapi/config"
	"webapi/internal/app/media"
	"webapi/internal/db/model"
	"webapi/internal/dto"
	"webapi/internal/helper/utils"
	"webapi/internal/http/requests"
	"webapi/internal/http/response"
	"webapi/internal/logger"
	"webapi/pkg/exception"
	internal_minio "webapi/pkg/minio"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
)

type MediaHttpHandler struct {
	app media.MediaApp
}

func NewMediaHttpHandler(app media.MediaApp) *MediaHttpHandler {
	return &MediaHttpHandler{app: app}
}

// GetMediaId godoc
// @Summary Get media by ID
// @Description Retrieve a specific media file by its UUID
// @Tags media
// @Accept json
// @Produce json
// @Param id path string true "Media UUID"
// @Success 200 {object} response.CommonResponse
// @Failure 400 {object} response.CommonResponse
// @Failure 404 {object} response.CommonResponse
// @Router /v1/media/{id} [get]
func (h *MediaHttpHandler) GetMediaId(c *fiber.Ctx) error {
	idParam := c.Params("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		return exception.InvalidIDError
	}
	mediaDto, err := h.app.GetMediaByID(c.Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(response.CommonResponse{
		Data: mediaDto,
	})
}

// CreateMedia godoc
// @Summary Upload media file
// @Description Upload a new media file to the system
// @Tags media
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Media file to upload"
// @Success 200 {object} response.CommonResponse
// @Failure 400 {object} response.CommonResponse
// @Failure 500 {object} response.CommonResponse
// @Router /v1/media [post]
func (h *MediaHttpHandler) CreateMedia(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Error getting file")
	}
	userID, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}
	fileContent, err := file.Open()
	conf := config.GetConfig().Minio
	objectName := file.Filename
	bucketName := conf.BucketName
	contentType := file.Header.Get("Content-Type")

	// Parse tags from form (as JSON array or comma-separated string)
	tags := []string{}
	tagsStr := c.FormValue("tags")
	if tagsStr != "" {
		if tagsStr[0] == '[' {
			// JSON array
			if err := c.BodyParser(&tags); err != nil {
				tags = []string{}
			}
		} else {
			// Comma-separated
			tags = append(tags, tagsStr)
		}
	}
	tagIDs := make([]uuid.UUID, 0, len(tags))
	for _, tagStr := range tags {
		tagID, err := uuid.Parse(tagStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid tag ID: " + tagStr})
		}
		tagIDs = append(tagIDs, tagID)
	}

	minioClient := internal_minio.GetMinio()
	uploadInfo, err := minioClient.PutObject(c.Context(), bucketName, objectName, fileContent, file.Size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		logger.Log.Error("File uploaded successfully", zap.Error(err))
		return err
	}

	mediaModel := model.Media{
		FileName:  objectName,
		Hash:      uploadInfo.ChecksumSHA256,
		Size:      file.Size,
		MimeType:  file.Header.Get("Content-Type"),
		CreatedBy: userID.String(),
		UpdatedBy: userID.String(),
	}
	var mediaDto *dto.GetMediaDTO
	if len(tagIDs) > 0 {
		mediaDto, err = h.app.CreateMediaWithTags(c.Context(), mediaModel, tagIDs)
	} else {
		mediaDto, err = h.app.CreateMedia(c.Context(), mediaModel)
	}
	if err != nil {
		return err
	}

	return c.JSON(response.CommonResponse{
		Data: mediaDto,
	})
}

// DeleteMedia godoc
// @Summary Delete media file
// @Description Delete a media file by its UUID
// @Tags media
// @Accept json
// @Produce json
// @Param id path string true "Media UUID"
// @Success 200 {object} response.CommonResponse
// @Failure 400 {object} response.CommonResponse
// @Failure 404 {object} response.CommonResponse
// @Router /v1/media/{id} [delete]
func (h *MediaHttpHandler) DeleteMedia(c *fiber.Ctx) error {
	idParam := c.Params("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		return exception.InvalidIDError
	}
	_, err = h.app.DeleteMedia(c.Context(), model.Media{
		ID: id,
	})
	if err != nil {
		return err
	}
	return c.JSON(response.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Media deleted successfully",
	})
}

// GetMediaList godoc
// @Summary Get all media files with pagination
// @Description Retrieve a paginated list of media files with optional search query
// @Tags media
// @Accept json
// @Produce json
// @Param limit query int false "Number of items per page (default: 10)"
// @Param page query int false "Page number (default: 1)"
// @Param query query string false "Search query for filtering media"
// @Success 200 {object} response.PaginationResponse
// @Failure 400 {object} response.CommonResponse
// @Failure 500 {object} response.CommonResponse
// @Router /v1/media [get]
func (h *MediaHttpHandler) GetMediaList(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 10) // Default to 10 if not provided
	page := c.QueryInt("page", 1)    // Default to 1 if not provided
	query := c.Query("query", "")    // Default to empty string if not provided

	offset := (page - 1) * limit
	req := requests.DataWithPaginationRequest{
		Query: query,
		Limit: limit,
		Page:  offset,
	}
	responseMedia, err := h.app.GetMediaWithPagination(c.Context(), req)
	if err != nil {
		return err
	}

	return c.JSON(response.PaginationResponse{
		TotalCount:   responseMedia.Total,
		TotalPage:    responseMedia.Total / limit,
		CurrentPage:  page,
		LastPage:     responseMedia.LastPage,
		PerPage:      limit,
		NextPage:     page + 1,
		PreviousPage: page - 1,
		Data:         responseMedia.Data,
		Path:         c.Path(),
	})
}
