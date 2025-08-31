package controllers

import (
	"strconv"

	"github.com/turahe/go-restfull/internal/application/ports"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/helper/utils"
	"github.com/turahe/go-restfull/internal/interfaces/http/requests"
	"github.com/turahe/go-restfull/internal/interfaces/http/responses"
	"github.com/turahe/go-restfull/pkg/exception"
	"github.com/turahe/go-restfull/pkg/logger"
	"github.com/turahe/go-restfull/pkg/storage"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// MediaController handles HTTP requests for media operations
//
//	@title						Media Management API
//	@version					1.0
//	@description				This is a media management API for uploading, managing, and serving media files
//	@termsOfService				http://swagger.io/terms/
//	@contact.name				API Support
//	@contact.email				support@example.com
//	@license.name				MIT
//	@license.url				https://opensource.org/licenses/MIT
//	@host						localhost:8000
//	@BasePath					/api/v1
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.
type MediaController struct {
	mediaService   ports.MediaService
	storageService *storage.StorageService
}

func NewMediaController(mediaService ports.MediaService, storageService *storage.StorageService) *MediaController {
	return &MediaController{
		mediaService:   mediaService,
		storageService: storageService,
	}
}

// GetMedia handles GET /v1/media requests
//
//	@Summary		Get all media
//	@Description	Retrieve a list of all media files with pagination
//	@Tags			media
//	@Accept			json
//	@Produce		json
//	@Param			limit	query		int												false	"Number of media items to return (default: 10, max: 100)"	default(10)	minimum(1)	maximum(100)
//	@Param			offset	query		int												false	"Number of media items to skip (default: 0)"				default(0)	minimum(0)
//	@Param			query	query		string											false	"Search query to filter media by name or filename"
//	@Success		200		{object}	responses.MediaCollectionResponse				"List of media files"
//	@Failure		400		{object}	responses.CommonResponse						"Bad request - Invalid parameters"
//	@Failure		500		{object}	responses.CommonResponse						"Internal server error"
//	@Security		BearerAuth
//	@Router			/media [get]
func (c *MediaController) GetMedia(ctx *fiber.Ctx) error {
	// Get query parameters
	limitStr := ctx.Query("limit", "10")
	offsetStr := ctx.Query("offset", "0")
	query := ctx.Query("query", "")

	// Parse pagination parameters
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if limit == 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get media based on whether search query is provided
	var media []*entities.Media
	if query != "" {
		media, err = c.mediaService.SearchMedia(ctx.Context(), query, limit, offset)
		if err != nil {
			logger.Log.Error("Failed to search media",
				zap.String("query", query),
				zap.Int("limit", limit),
				zap.Int("offset", offset),
				zap.Error(err),
			)
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to search media",
			})
		}
	} else {
		media, err = c.mediaService.GetAllMedia(ctx.Context(), limit, offset)
		if err != nil {
			logger.Log.Error("Failed to retrieve media",
				zap.Int("limit", limit),
				zap.Int("offset", offset),
				zap.Error(err),
			)
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to retrieve media",
			})
		}
	}

	// Get total count for pagination
	total, err := c.mediaService.GetMediaCount(ctx.Context())
	if err != nil {
		logger.Log.Error("Failed to retrieve media count",
			zap.Error(err),
		)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to retrieve media count",
		})
	}

	// Calculate page from offset
	page := (offset / limit) + 1
	if offset == 0 {
		page = 1
	}

	// Build base URL for pagination links
	baseURL := ctx.BaseURL() + ctx.Path()

	// Return paginated media collection response
	return ctx.Status(fiber.StatusOK).JSON(responses.NewPaginatedMediaCollection(
		media, page, limit, int(total), baseURL,
	))
}

// GetMediaByID handles GET /v1/media/:id requests
//
//	@Summary		Get media by ID
//	@Description	Retrieve a specific media file by its unique identifier
//	@Tags			media
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string													true	"Media ID"	format(uuid)
//	@Success		200	{object}	responses.MediaResourceResponse						"Media file details"
//	@Failure		400	{object}	responses.CommonResponse								"Bad request - Invalid media ID"
//	@Failure		404	{object}	responses.CommonResponse								"Not found - Media does not exist"
//	@Failure		500	{object}	responses.CommonResponse								"Internal server error"
//	@Security		BearerAuth
//	@Router			/media/{id} [get]
func (c *MediaController) GetMediaByID(ctx *fiber.Ctx) error {
	// Parse media ID
	mediaID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		logger.Log.Error("GetMediaByID: Invalid media ID format",
			zap.String("media_id", ctx.Params("id")),
			zap.Error(err),
		)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid media ID format",
		})
	}

	// Get media by ID
	media, err := c.mediaService.GetMediaByID(ctx.Context(), mediaID)
	if err != nil {
		if err == exception.DataNotFoundError {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "error",
				"message": "Media not found",
			})
		}
		logger.Log.Error("GetMediaByID: Failed to retrieve media",
			zap.String("media_id", mediaID.String()),
			zap.Error(err),
		)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to retrieve media",
		})
	}

	// Return media resource response
	return ctx.Status(fiber.StatusOK).JSON(responses.NewMediaResource(media))
}

// CreateMedia handles POST /v1/media requests
//
//	@Summary		Upload media
//	@Description	Upload a new media file to the system
//	@Tags			media
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file		formData	file													true	"Media file to upload"
//	@Param			name		formData	string													false	"Custom name for the media file"
//	@Param			description	formData	string													false	"Description of the media file"
//	@Success		201			{object}	responses.MediaResourceResponse					"Media uploaded successfully"
//	@Failure		400			{object}	responses.CommonResponse								"Bad request - Invalid file or parameters"
//	@Failure		413			{object}	responses.CommonResponse								"Payload too large - File size exceeds limit"
//	@Failure		415			{object}	responses.CommonResponse								"Unsupported media type"
//	@Failure		500			{object}	responses.CommonResponse								"Internal server error"
//	@Security		BearerAuth
//	@Router			/media [post]
func (c *MediaController) CreateMedia(ctx *fiber.Ctx) error {
	logger.Log.Info("CreateMedia: Request received",
		zap.String("ip", ctx.IP()),
		zap.String("method", ctx.Method()),
		zap.String("path", ctx.Path()),
		zap.String("user_agent", ctx.Get("User-Agent")),
	)

	// Get user ID from context (assuming it's set by JWT middleware)
	userIDInterface := ctx.Locals("user_id")
	if userIDInterface == nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "User not authenticated",
		})
	}

	// Type assert directly to uuid.UUID
	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		logger.Log.Error("CreateMedia: Invalid user ID format",
			zap.String("ip", ctx.IP()),
			zap.Any("user_id_interface", userIDInterface),
		)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid user ID format",
		})
	}

	// Get uploaded file
	file, err := ctx.FormFile("file")
	if err != nil {
		logger.Log.Error("CreateMedia: No file uploaded",
			zap.String("user_id", userID.String()),
			zap.String("ip", ctx.IP()),
			zap.Error(err),
		)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "No file uploaded",
		})
	}

	// Validate file size and type
	if file.Size == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "File is empty",
		})
	}

	// Validate file size (max 50MB)
	const maxFileSize = 50 * 1024 * 1024 // 50MB
	if file.Size > maxFileSize {
		return ctx.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
			"status":  "error",
			"message": "File size exceeds maximum limit of 50MB",
		})
	}

	// Validate file type
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "File type not detected",
		})
	}

	// Check if it's a supported media type
	if !utils.IsSupportedMediaType(contentType) {
		return ctx.Status(fiber.StatusUnsupportedMediaType).JSON(fiber.Map{
			"status":  "error",
			"message": "Unsupported file type: " + contentType,
		})
	}

	// Check if storage service is available
	if c.storageService == nil {
		logger.Log.Error("CreateMedia: Storage service not available",
			zap.String("user_id", userID.String()),
		)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Storage service not available",
		})
	}

	// Test storage connection before attempting upload
	if err := c.storageService.TestConnection(); err != nil {
		logger.Log.Error("CreateMedia: Storage connection test failed",
			zap.String("user_id", userID.String()),
			zap.String("filename", file.Filename),
			zap.Error(err),
		)
		return ctx.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":  "error",
			"message": "Storage service is not accessible. Please try again later.",
		})
	}

	// Upload file and save metadata using media service (which handles both storage and database)
	media, err := c.mediaService.UploadMedia(ctx.Context(), file, userID)
	if err != nil {
		logger.Log.Error("CreateMedia: Failed to upload media",
			zap.String("user_id", userID.String()),
			zap.String("filename", file.Filename),
			zap.Int64("file_size", file.Size),
			zap.String("content_type", file.Header.Get("Content-Type")),
			zap.Error(err),
		)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to upload media: " + err.Error(),
		})
	}

	logger.Log.Info("CreateMedia: Media uploaded successfully",
		zap.String("user_id", userID.String()),
		zap.String("media_id", media.ID.String()),
		zap.String("filename", file.Filename),
		zap.Int64("file_size", file.Size),
	)

	// Return media resource response
	return ctx.Status(fiber.StatusCreated).JSON(responses.NewMediaResource(media))
}

// UpdateMedia handles PUT /v1/media/:id requests
//
//	@Summary		Update media
//	@Description	Update media file metadata (name, description, etc.)
//	@Tags			media
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string													true	"Media ID"	format(uuid)
//	@Param			media	body		object													true	"Media update request"
//	@Success		200		{object}	responses.MediaResourceResponse					"Media updated successfully"
//	@Failure		400		{object}	responses.CommonResponse								"Bad request - Invalid input data"
//	@Failure		404		{object}	responses.CommonResponse								"Not found - Media does not exist"
//	@Failure		500		{object}	responses.CommonResponse								"Internal server error"
//	@Security		BearerAuth
//	@Router			/media/{id} [put]
func (c *MediaController) UpdateMedia(ctx *fiber.Ctx) error {
	// Parse media ID
	mediaID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		logger.Log.Error("UpdateMedia: Invalid media ID format",
			zap.String("media_id", ctx.Params("id")),
			zap.Error(err),
		)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid media ID format",
		})
	}

	// Parse request body
	var request requests.UpdateMediaRequest

	if err := ctx.BodyParser(&request); err != nil {
		logger.Log.Error("UpdateMedia: Invalid request body",
			zap.String("media_id", mediaID.String()),
			zap.Error(err),
		)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Validate request
	if err := request.Validate(); err != nil {
		logger.Log.Error("UpdateMedia: Validation failed",
			zap.String("media_id", mediaID.String()),
			zap.Error(err),
		)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Validation failed: " + err.Error(),
		})
	}

	// Get existing media
	existingMedia, err := c.mediaService.GetMediaByID(ctx.Context(), mediaID)
	if err != nil {
		if err == exception.DataNotFoundError {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "error",
				"message": "Media not found",
			})
		}
		logger.Log.Error("UpdateMedia: Failed to retrieve existing media",
			zap.String("media_id", mediaID.String()),
			zap.Error(err),
		)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to retrieve media",
		})
	}

	// Transform request to entity
	updatedMedia := request.ToEntity(existingMedia)

	// Update media using the entity
	media, err := c.mediaService.UpdateMedia(ctx.Context(), mediaID, updatedMedia.Name, updatedMedia.FileName,
		updatedMedia.Hash, updatedMedia.Disk, updatedMedia.MimeType, updatedMedia.Size)
	if err != nil {
		if err == exception.DataNotFoundError {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "error",
				"message": "Media not found",
			})
		}
		logger.Log.Error("UpdateMedia: Failed to update media",
			zap.String("media_id", mediaID.String()),
			zap.Error(err),
		)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update media",
		})
	}

	// Return media resource response
	return ctx.Status(fiber.StatusOK).JSON(responses.NewMediaResource(media))
}

// DeleteMedia handles DELETE /v1/media/:id requests
//
//	@Summary		Delete media
//	@Description	Delete a media file from the system
//	@Tags			media
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string						true	"Media ID"	format(uuid)
//	@Success		200	{object}	responses.CommonResponse	"Media deleted successfully"
//	@Failure		400	{object}	responses.CommonResponse	"Bad request - Invalid media ID"
//	@Failure		404	{object}	responses.CommonResponse	"Not found - Media does not exist"
//	@Failure		500	{object}	responses.CommonResponse	"Internal server error"
//	@Security		BearerAuth
//	@Router			/media/{id} [delete]
func (c *MediaController) DeleteMedia(ctx *fiber.Ctx) error {
	// Parse media ID
	mediaID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		logger.Log.Error("DeleteMedia: Invalid media ID format",
			zap.String("media_id", ctx.Params("id")),
			zap.Error(err),
		)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid media ID format",
		})
	}

	// Delete media
	err = c.mediaService.DeleteMedia(ctx.Context(), mediaID)
	if err != nil {
		if err == exception.DataNotFoundError {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "error",
				"message": "Media not found",
			})
		}
		logger.Log.Error("DeleteMedia: Failed to delete media",
			zap.String("media_id", mediaID.String()),
			zap.Error(err),
		)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to delete media",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Media deleted successfully",
	})
}
