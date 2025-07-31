package controllers

import (
	"strconv"
	"github.com/turahe/go-restfull/internal/application/ports"
	"github.com/turahe/go-restfull/internal/interfaces/http/responses"
	"github.com/turahe/go-restfull/pkg/exception"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// MediaController handles HTTP requests for media operations
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
	mediaService ports.MediaService
}

func NewMediaController(mediaService ports.MediaService) *MediaController {
	return &MediaController{
		mediaService: mediaService,
	}
}

// GetMedia handles GET /v1/media requests
//	@Summary		Get all media
//	@Description	Retrieve a list of all media files with pagination
//	@Tags			media
//	@Accept			json
//	@Produce		json
//	@Param			limit	query		int												false	"Number of media items to return (default: 10, max: 100)"	default(10)	minimum(1)	maximum(100)
//	@Param			offset	query		int												false	"Number of media items to skip (default: 0)"				default(0)	minimum(0)
//	@Param			query	query		string											false	"Search query to filter media by name or filename"
//	@Success		200		{object}	responses.CommonResponse{data=[]interface{}}	"List of media files"
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

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get media based on whether search query is provided
	var mediaList []interface{}
	if query != "" {
		media, err := c.mediaService.SearchMedia(ctx.Context(), query, limit, offset)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
				ResponseCode:    fiber.StatusInternalServerError,
				ResponseMessage: "Failed to search media",
				Data:            nil,
			})
		}
		// Convert to interface slice
		for _, m := range media {
			mediaList = append(mediaList, m)
		}
	} else {
		media, err := c.mediaService.GetAllMedia(ctx.Context(), limit, offset)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
				ResponseCode:    fiber.StatusInternalServerError,
				ResponseMessage: "Failed to retrieve media",
				Data:            nil,
			})
		}
		// Convert to interface slice
		for _, m := range media {
			mediaList = append(mediaList, m)
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(responses.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Media retrieved successfully",
		Data:            mediaList,
	})
}

// GetMediaByID handles GET /v1/media/:id requests
//	@Summary		Get media by ID
//	@Description	Retrieve a specific media file by its unique identifier
//	@Tags			media
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string													true	"Media ID"	format(uuid)
//	@Success		200	{object}	responses.CommonResponse{data=map[string]interface{}}	"Media file details"
//	@Failure		400	{object}	responses.CommonResponse								"Bad request - Invalid media ID"
//	@Failure		404	{object}	responses.CommonResponse								"Not found - Media does not exist"
//	@Failure		500	{object}	responses.CommonResponse								"Internal server error"
//	@Security		BearerAuth
//	@Router			/media/{id} [get]
func (c *MediaController) GetMediaByID(ctx *fiber.Ctx) error {
	// Parse media ID
	mediaID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid media ID format",
			Data:            nil,
		})
	}

	// Get media by ID
	media, err := c.mediaService.GetMediaByID(ctx.Context(), mediaID)
	if err != nil {
		if err == exception.DataNotFoundError {
			return ctx.Status(fiber.StatusNotFound).JSON(responses.CommonResponse{
				ResponseCode:    fiber.StatusNotFound,
				ResponseMessage: "Media not found",
				Data:            nil,
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to retrieve media",
			Data:            nil,
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(responses.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Media retrieved successfully",
		Data:            media,
	})
}

// CreateMedia handles POST /v1/media requests
//	@Summary		Upload media
//	@Description	Upload a new media file to the system
//	@Tags			media
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file		formData	file													true	"Media file to upload"
//	@Param			name		formData	string													false	"Custom name for the media file"
//	@Param			description	formData	string													false	"Description of the media file"
//	@Success		201			{object}	responses.CommonResponse{data=map[string]interface{}}	"Media uploaded successfully"
//	@Failure		400			{object}	responses.CommonResponse								"Bad request - Invalid file or parameters"
//	@Failure		413			{object}	responses.CommonResponse								"Payload too large - File size exceeds limit"
//	@Failure		415			{object}	responses.CommonResponse								"Unsupported media type"
//	@Failure		500			{object}	responses.CommonResponse								"Internal server error"
//	@Security		BearerAuth
//	@Router			/media [post]
func (c *MediaController) CreateMedia(ctx *fiber.Ctx) error {
	// Get user ID from context (assuming it's set by JWT middleware)
	userIDStr := ctx.Locals("user_id")
	if userIDStr == nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusUnauthorized,
			ResponseMessage: "User not authenticated",
			Data:            nil,
		})
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid user ID",
			Data:            nil,
		})
	}

	// Get uploaded file
	file, err := ctx.FormFile("file")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "No file uploaded",
			Data:            nil,
		})
	}

	// Upload media
	media, err := c.mediaService.UploadMedia(ctx.Context(), file, userID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to upload media",
			Data:            nil,
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(responses.CommonResponse{
		ResponseCode:    fiber.StatusCreated,
		ResponseMessage: "Media uploaded successfully",
		Data:            media,
	})
}

// UpdateMedia handles PUT /v1/media/:id requests
//	@Summary		Update media
//	@Description	Update media file metadata (name, description, etc.)
//	@Tags			media
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string													true	"Media ID"	format(uuid)
//	@Param			media	body		object													true	"Media update request"
//	@Success		200		{object}	responses.CommonResponse{data=map[string]interface{}}	"Media updated successfully"
//	@Failure		400		{object}	responses.CommonResponse								"Bad request - Invalid input data"
//	@Failure		404		{object}	responses.CommonResponse								"Not found - Media does not exist"
//	@Failure		500		{object}	responses.CommonResponse								"Internal server error"
//	@Security		BearerAuth
//	@Router			/media/{id} [put]
func (c *MediaController) UpdateMedia(ctx *fiber.Ctx) error {
	// Parse media ID
	mediaID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid media ID format",
			Data:            nil,
		})
	}

	// Parse request body
	var requestBody struct {
		FileName     string `json:"file_name"`
		OriginalName string `json:"original_name"`
		MimeType     string `json:"mime_type"`
		Path         string `json:"path"`
		URL          string `json:"url"`
		Size         int64  `json:"size"`
	}

	if err := ctx.BodyParser(&requestBody); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid request body",
			Data:            nil,
		})
	}

	// Update media
	media, err := c.mediaService.UpdateMedia(ctx.Context(), mediaID, requestBody.FileName, requestBody.OriginalName,
		requestBody.MimeType, requestBody.Path, requestBody.URL, requestBody.Size)
	if err != nil {
		if err == exception.DataNotFoundError {
			return ctx.Status(fiber.StatusNotFound).JSON(responses.CommonResponse{
				ResponseCode:    fiber.StatusNotFound,
				ResponseMessage: "Media not found",
				Data:            nil,
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to update media",
			Data:            nil,
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(responses.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Media updated successfully",
		Data:            media,
	})
}

// DeleteMedia handles DELETE /v1/media/:id requests
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
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid media ID format",
			Data:            nil,
		})
	}

	// Delete media
	err = c.mediaService.DeleteMedia(ctx.Context(), mediaID)
	if err != nil {
		if err == exception.DataNotFoundError {
			return ctx.Status(fiber.StatusNotFound).JSON(responses.CommonResponse{
				ResponseCode:    fiber.StatusNotFound,
				ResponseMessage: "Media not found",
				Data:            nil,
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to delete media",
			Data:            nil,
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(responses.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Media deleted successfully",
		Data:            nil,
	})
}
