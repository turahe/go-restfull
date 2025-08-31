package controllers

import (
	"errors"

	"github.com/turahe/go-restfull/internal/application/ports"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/services"
	"github.com/turahe/go-restfull/internal/helper/utils"
	"github.com/turahe/go-restfull/internal/interfaces/http/requests"
	"github.com/turahe/go-restfull/internal/interfaces/http/responses"
	"github.com/turahe/go-restfull/pkg/exception"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/turahe/go-restfull/pkg/logger"
	"go.uber.org/zap")

// CommentController handles comment-related endpoints
type CommentController struct {
	commentService              ports.CommentService
	notificationService         services.NotificationService
	notificationTemplateService services.NotificationTemplateService
}

func NewCommentController(commentService ports.CommentService, notificationService services.NotificationService, notificationTemplateService services.NotificationTemplateService) *CommentController {
	return &CommentController{
		commentService:              commentService,
		notificationService:         notificationService,
		notificationTemplateService: notificationTemplateService,
	}
}

// GetComments godoc
//
//	@Summary		List comments
//	@Description	Get all comments with optional filtering
//	@Tags			comments
//	@Accept			json
//	@Produce		json
//	@Param			post_id		query		string	false	"Filter by post ID"
//	@Param			user_id		query		string	false	"Filter by user ID"
//	@Param			parent_id	query		string	false	"Filter by parent comment ID"
//	@Param			status		query		string	false	"Filter by status (approved, pending, rejected)"
//	@Param			limit		query		int		false	"Number of comments to return (default: 10, max: 100)"
//	@Param			offset		query		int		false	"Number of comments to skip (default: 0)"
//	@Success		200			{object}	responses.CommentCollectionResponse
//	@Failure		400			{object}	responses.CommonResponse
//	@Failure		500			{object}	responses.CommonResponse
//	@Router			/v1/comments [get]
func (c *CommentController) GetComments(ctx *fiber.Ctx) error {
	// Parse query parameters
	var queryParams requests.CommentQueryParams
	if err := ctx.QueryParser(&queryParams); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid query parameters",
			Data:            nil,
		})
	}

	// Set default values
	queryParams.SetDefaults()

	// Get comments based on filters
	var comments []*entities.Comment
	var err error

	if queryParams.PostID != nil {
		// Get comments by post ID
		comments, err = c.commentService.GetCommentsByPostID(ctx.Context(), *queryParams.PostID, queryParams.Limit, queryParams.Offset)
	} else if queryParams.UserID != nil {
		// Get comments by user ID
		comments, err = c.commentService.GetCommentsByUserID(ctx.Context(), *queryParams.UserID, queryParams.Limit, queryParams.Offset)
	} else if queryParams.ParentID != nil {
		// Get comment replies
		comments, err = c.commentService.GetCommentReplies(ctx.Context(), *queryParams.ParentID, queryParams.Limit, queryParams.Offset)
	} else {
		// Get all comments based on status
		switch queryParams.Status {
		case "approved":
			comments, err = c.commentService.GetApprovedComments(ctx.Context(), queryParams.Limit, queryParams.Offset)
		case "pending":
			comments, err = c.commentService.GetPendingComments(ctx.Context(), queryParams.Limit, queryParams.Offset)
		default:
			comments, err = c.commentService.GetAllComments(ctx.Context(), queryParams.Limit, queryParams.Offset)
		}
	}

	if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to retrieve comments",
			Data:            nil,
		})
	}

	// Get total count for pagination
	var total int64
	switch {
	case queryParams.PostID != nil:
		total, err = c.commentService.GetCommentCountByPostID(ctx.Context(), *queryParams.PostID)
	case queryParams.UserID != nil:
		total, err = c.commentService.GetCommentCountByUserID(ctx.Context(), *queryParams.UserID)
	case queryParams.ParentID != nil:
		// For replies, we'll use the general comment count as approximation
		total, err = c.commentService.GetCommentCount(ctx.Context())
	default:
		switch queryParams.Status {
		case "approved":
			// For approved comments, we'll use the general comment count as approximation
			total, err = c.commentService.GetCommentCount(ctx.Context())
		case "pending":
			total, err = c.commentService.GetPendingCommentCount(ctx.Context())
		default:
			total, err = c.commentService.GetCommentCount(ctx.Context())
		}
	}

	if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to retrieve comment count",
			Data:            nil,
		})
	}

	// Calculate page from offset
	page := (queryParams.Offset / queryParams.Limit) + 1
	if queryParams.Offset == 0 {
		page = 1
	}

	// Build base URL for pagination links
	baseURL := ctx.BaseURL() + ctx.Path()

	// Return paginated comment collection response
	return ctx.Status(fiber.StatusOK).JSON(responses.NewPaginatedCommentCollectionResponse(
		comments, page, queryParams.Limit, int(total), baseURL, nil, nil,
	))
}

// GetCommentByID godoc
//
//	@Summary		Get comment by ID
//	@Description	Get a single comment by its ID
//	@Tags			comments
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Comment ID"
//	@Success		200	{object}	responses.CommentResourceResponse
//	@Failure		400	{object}	responses.CommonResponse
//	@Failure		404	{object}	responses.CommonResponse
//	@Failure		500	{object}	responses.CommonResponse
//	@Router			/v1/comments/{id} [get]
func (c *CommentController) GetCommentByID(ctx *fiber.Ctx) error {
	// Parse comment ID from path
	commentIDStr := ctx.Params("id")
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid comment ID format",
			Data:            nil,
		})
	}

	// Get comment by ID
	comment, err := c.commentService.GetCommentByID(ctx.Context(), commentID)
	if err != nil {
		if err == exception.DataNotFoundError {
			return ctx.Status(fiber.StatusNotFound).JSON(responses.CommonResponse{
				ResponseCode:    fiber.StatusNotFound,
				ResponseMessage: "Comment not found",
				Data:            nil,
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to retrieve comment",
			Data:            nil,
		})
	}

	// Return comment resource response
	return ctx.Status(fiber.StatusOK).JSON(responses.NewCommentResource(comment, nil, nil))
}

// CreateComment godoc
//
//	@Summary		Create comment
//	@Description	Create a new comment for a post
//	@Tags			comments
//	@Accept			json
//	@Produce		json
//	@Param			comment	body		requests.CreateCommentRequest	true	"Comment info"
//	@Success		201		{object}	responses.CommentResourceResponse
//	@Failure		400		{object}	responses.CommonResponse
//	@Failure		401		{object}	responses.CommonResponse
//	@Router			/v1/comments [post]
//	@Security		BearerAuth
func (c *CommentController) CreateComment(ctx *fiber.Ctx) error {
	// Get authenticated user ID
	userID, err := utils.GetUserID(ctx)
	if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(fiber.StatusUnauthorized).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusUnauthorized,
			ResponseMessage: "Authentication required",
			Data:            nil,
		})
	}

	// Parse request body
	var req requests.CreateCommentRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid request body",
			Data:            nil,
		})
	}

	// Transform request to entity
	comment := req.ToEntity(userID)

	// Create comment using the entity
	createdComment, err := c.commentService.CreateComment(ctx.Context(), comment)
	if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to create comment",
			Data:            nil,
		})
	}

	// Return comment resource response
	return ctx.Status(fiber.StatusCreated).JSON(responses.NewCommentResource(createdComment, nil, nil))
}

// UpdateComment godoc
//
//	@Summary		Update comment
//	@Description	Update an existing comment
//	@Tags			comments
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Comment ID"
//	@Param			comment	body		requests.UpdateCommentRequest	true	"Comment info"
//	@Success		200		{object}	responses.CommentResourceResponse
//	@Failure		400		{object}	responses.CommonResponse
//	@Failure		401		{object}	responses.CommonResponse
//	@Failure		404		{object}	responses.CommonResponse
//	@Failure		500		{object}	responses.CommonResponse
//	@Router			/v1/comments/{id} [put]
//	@Security		BearerAuth
func (c *CommentController) UpdateComment(ctx *fiber.Ctx) error {
	// Get authenticated user ID
	_, err := utils.GetUserID(ctx)
	if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(fiber.StatusUnauthorized).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusUnauthorized,
			ResponseMessage: "Authentication required",
			Data:            nil,
		})
	}

	// Parse comment ID from path
	commentIDStr := ctx.Params("id")
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid comment ID format",
			Data:            nil,
		})
	}

	// Parse request body
	var req requests.UpdateCommentRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid request body",
			Data:            nil,
		})
	}

	// Get existing comment
	existingComment, err := c.commentService.GetCommentByID(ctx.Context(), commentID)
	if err != nil {
		if err == exception.DataNotFoundError {
			return ctx.Status(fiber.StatusNotFound).JSON(responses.CommonResponse{
				ResponseCode:    fiber.StatusNotFound,
				ResponseMessage: "Comment not found",
				Data:            nil,
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to retrieve comment",
			Data:            nil,
		})
	}

	// Transform request to entity
	updatedComment := req.ToEntity(existingComment)

	// Update comment using the entity
	comment, err := c.commentService.UpdateComment(ctx.Context(), updatedComment)
	if err != nil {
		if err == exception.DataNotFoundError {
			return ctx.Status(fiber.StatusNotFound).JSON(responses.CommonResponse{
				ResponseCode:    fiber.StatusNotFound,
				ResponseMessage: "Comment not found",
				Data:            nil,
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to update comment",
			Data:            nil,
		})
	}

	// Return comment resource response
	return ctx.Status(fiber.StatusOK).JSON(responses.NewCommentResource(comment, nil, nil))
}

// DeleteComment godoc
//
//	@Summary		Delete comment
//	@Description	Delete a comment by its ID
//	@Tags			comments
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Comment ID"
//	@Success		200	{object}	responses.CommonResponse
//	@Failure		400	{object}	responses.CommonResponse
//	@Failure		401	{object}	responses.CommonResponse
//	@Failure		404	{object}	responses.CommonResponse
//	@Failure		500	{object}	responses.CommonResponse
//	@Router			/v1/comments/{id} [delete]
func (c *CommentController) DeleteComment(ctx *fiber.Ctx) error {
	// Get authenticated user ID
	_, err := utils.GetUserID(ctx)
	if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(fiber.StatusUnauthorized).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusUnauthorized,
			ResponseMessage: "Authentication required",
			Data:            nil,
		})
	}

	// Parse comment ID from path
	commentIDStr := ctx.Params("id")
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid comment ID format",
			Data:            nil,
		})
	}

	// Delete comment
	err = c.commentService.DeleteComment(ctx.Context(), commentID)
	if err != nil {
		if err == exception.DataNotFoundError {
			return ctx.Status(fiber.StatusNotFound).JSON(responses.CommonResponse{
				ResponseCode:    fiber.StatusNotFound,
				ResponseMessage: "Comment not found",
				Data:            nil,
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to delete comment",
			Data:            nil,
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(responses.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Comment deleted successfully",
		Data:            nil,
	})
}

// ApproveComment godoc
//
//	@Summary		Approve comment
//	@Description	Approve a comment by its ID
//	@Tags			comments
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Comment ID"
//	@Success		200	{object}	responses.CommentResourceResponse
//	@Failure		400	{object}	responses.CommonResponse
//	@Failure		401	{object}	responses.CommonResponse
//	@Failure		404	{object}	responses.CommonResponse
//	@Failure		500	{object}	responses.CommonResponse
//	@Router			/v1/comments/{id}/approve [put]
func (c *CommentController) ApproveComment(ctx *fiber.Ctx) error {
	// Get authenticated user ID
	_, err := utils.GetUserID(ctx)
	if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(fiber.StatusUnauthorized).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusUnauthorized,
			ResponseMessage: "Authentication required",
			Data:            nil,
		})
	}

	// Parse comment ID from path
	commentIDStr := ctx.Params("id")
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid comment ID format",
			Data:            nil,
		})
	}

	// Approve comment
	err = c.commentService.ApproveComment(ctx.Context(), commentID)
	if err != nil {
		if err == exception.DataNotFoundError {
			return ctx.Status(fiber.StatusNotFound).JSON(responses.CommonResponse{
				ResponseCode:    fiber.StatusNotFound,
				ResponseMessage: "Comment not found",
				Data:            nil,
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to approve comment",
			Data:            nil,
		})
	}

	// Get the updated comment to return in response
	comment, err := c.commentService.GetCommentByID(ctx.Context(), commentID)
	if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to retrieve updated comment",
			Data:            nil,
		})
	}

	// Return comment resource response
	return ctx.Status(fiber.StatusOK).JSON(responses.NewCommentResource(comment, nil, nil))
}

// RejectComment godoc
//
//	@Summary		Reject comment
//	@Description	Reject a comment by its ID
//	@Tags			comments
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Comment ID"
//	@Success		200	{object}	responses.CommentResourceResponse
//	@Failure		400	{object}	responses.CommonResponse
//	@Failure		401	{object}	responses.CommonResponse
//	@Failure		404	{object}	responses.CommonResponse
//	@Failure		500	{object}	responses.CommonResponse
//	@Router			/v1/comments/{id}/reject [put]
func (c *CommentController) RejectComment(ctx *fiber.Ctx) error {
	// Get authenticated user ID
	_, err := utils.GetUserID(ctx)
	if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(fiber.StatusUnauthorized).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusUnauthorized,
			ResponseMessage: "Authentication required",
			Data:            nil,
		})
	}

	// Parse comment ID from path
	commentIDStr := ctx.Params("id")
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid comment ID format",
			Data:            nil,
		})
	}

	// Get the comment first to get the user ID for notification
	comment, err := c.commentService.GetCommentByID(ctx.Context(), commentID)
	if err != nil {
		if errors.Is(err, exception.DataNotFoundError) {
			return ctx.Status(fiber.StatusNotFound).JSON(responses.CommonResponse{
				ResponseCode:    fiber.StatusNotFound,
				ResponseMessage: "Comment not found",
				Data:            nil,
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to retrieve comment",
			Data:            nil,
		})
	}

	// Reject comment
	err = c.commentService.RejectComment(ctx.Context(), commentID)
	if err != nil {
		if errors.Is(err, exception.DataNotFoundError) {
			return ctx.Status(fiber.StatusNotFound).JSON(responses.CommonResponse{
				ResponseCode:    fiber.StatusNotFound,
				ResponseMessage: "Comment not found",
				Data:            nil,
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to reject comment",
			Data:            nil,
		})
	}

	// Send notification to the comment author about the rejection using template
	if c.notificationService != nil {
		notificationData := map[string]interface{}{
			"comment_id":      comment.ID.String(),
			"comment_content": comment.Content,
			"model_type":      comment.ModelType,
			"model_id":        comment.ModelID.String(),
			"rejected_at":     comment.UpdatedAt.Format("2006-01-02 15:04:05"),
		}

		err = c.notificationService.SendNotificationFromTemplate(
			ctx.Context(),
			comment.CreatedBy,  // The user who created the comment
			"comment_rejected", // Template name from the migration
			notificationData,
		)

		// Log notification error but don't fail the request
		if err != nil {
			// In production, you might want to log this error
			// For now, we'll just continue with the response
		}
	}

	// Get the updated comment to return in response
	updatedComment, err := c.commentService.GetCommentByID(ctx.Context(), commentID)
	if err != nil {
		logger.Log.Error("Error occurred", zap.Error(err))
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to retrieve updated comment",
			Data:            nil,
		})
	}

	// Return comment resource response
	return ctx.Status(fiber.StatusOK).JSON(responses.NewCommentResource(updatedComment, nil, nil))
}
