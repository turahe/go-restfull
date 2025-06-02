package media

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
	"webapi/config"
	"webapi/internal/app/media"
	"webapi/internal/db/model"
	"webapi/internal/http/requests"
	"webapi/internal/http/response"
	"webapi/internal/logger"
	"webapi/pkg/exception"
	internal_minio "webapi/pkg/minio"
)

type MediaHttpHandler struct {
	app media.MediaApp
}

func NewMediaHttpHandler(app media.MediaApp) *MediaHttpHandler {
	return &MediaHttpHandler{app: app}
}

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

func (h *MediaHttpHandler) CreateMedia(c *fiber.Ctx) error {

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Error getting file")
	}
	fileContent, err := file.Open()
	conf := config.GetConfig().Minio
	objectName := file.Filename
	bucketName := conf.BucketName
	contentType := file.Header.Get("Content-Type")

	minioClient := internal_minio.GetMinio()
	uploadInfo, err := minioClient.PutObject(c.Context(), bucketName, objectName, fileContent, file.Size, minio.PutObjectOptions{ContentType: contentType})
	//logger.Log.Error(uploadInfo, "File uploaded successfully")

	if err != nil {
		logger.Log.Error("File uploaded successfully", zap.Error(err))
		return err
	}

	mediaDto, err := h.app.CreateMedia(c.Context(), model.Media{
		FileName: objectName,
		Hash:     uploadInfo.ChecksumSHA256,
		//ParentID: req.ParentID,
		Size:     file.Size,
		MimeType: file.Header.Get("Content-Type"),
	})
	if err != nil {
		return err
	}

	return c.JSON(response.CommonResponse{
		Data: mediaDto,
	})
}

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
