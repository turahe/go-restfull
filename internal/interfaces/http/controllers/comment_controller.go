package controllers

import (
	"webapi/internal/application/ports"
	"webapi/internal/domain/entities"
	"webapi/internal/helper/utils"
	"webapi/internal/interfaces/http/requests"
	"webapi/internal/interfaces/http/responses"
	"webapi/pkg/exception"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// CommentController handles comment-related endpoints
type CommentController struct {
	commentService ports.CommentService
}

func NewCommentController(commentService ports.CommentService) *CommentController {
	return &CommentController{
		commentService: commentService,
	}
}

// GetComments godoc
// @Summary      List comments
// @Description  Get all comments with optional filtering
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        post_id   query     string  false  "Filter by post ID"
// @Param        user_id   query     string  false  "Filter by user ID"
// @Param        parent_id query     string  false  "Filter by parent comment ID"
// @Param        status    query     string  false  "Filter by status (approved, pending, rejected)"
// @Param        limit     query     int     false  "Number of comments to return (default: 10, max: 100)"
// @Param        offset    query     int     false  "Number of comments to skip (default: 0)"
// @Success      200       {object}  responses.CommonResponse
// @Failure      400       {object}  responses.CommonResponse
// @Failure      500       {object}  responses.CommonResponse
// @Router       /v1/comments [get]
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
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to retrieve comments",
			Data:            nil,
		})
	}

	// Convert to interface slice for response
	var commentList []interface{}
	for _, comment := range comments {
		commentList = append(commentList, comment)
	}

	return ctx.Status(fiber.StatusOK).JSON(responses.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Comments retrieved successfully",
		Data:            commentList,
	})
}

// GetCommentByID godoc
// @Summary      Get comment by ID
// @Description  Get a single comment by its ID
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Comment ID"
// @Success      200  {object}  responses.CommonResponse
// @Failure      400  {object}  responses.CommonResponse
// @Failure      404  {object}  responses.CommonResponse
// @Failure      500  {object}  responses.CommonResponse
// @Router       /v1/comments/{id} [get]
func (c *CommentController) GetCommentByID(ctx *fiber.Ctx) error {
	// Parse comment ID from path
	commentIDStr := ctx.Params("id")
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
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

	return ctx.Status(fiber.StatusOK).JSON(responses.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Comment retrieved successfully",
		Data:            comment,
	})
}

// CreateComment godoc
// @Summary      Create comment
// @Description  Create a new comment
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        comment  body      requests.CreateCommentRequest  true  "Comment info"
// @Success      201      {object}  responses.CommonResponse
// @Failure      400      {object}  responses.CommonResponse
// @Failure      401      {object}  responses.CommonResponse
// @Failure      500      {object}  responses.CommonResponse
// @Router       /v1/comments [post]
func (c *CommentController) CreateComment(ctx *fiber.Ctx) error {
	// Get authenticated user ID
	userID, err := utils.GetUserID(ctx)
	if err != nil {
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

	// Create comment
	comment, err := c.commentService.CreateComment(ctx.Context(), req.Content, req.PostID, userID, req.ParentID, "pending")
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusInternalServerError,
			ResponseMessage: "Failed to create comment",
			Data:            nil,
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(responses.CommonResponse{
		ResponseCode:    fiber.StatusCreated,
		ResponseMessage: "Comment created successfully",
		Data:            comment,
	})
}

// UpdateComment godoc
// @Summary      Update comment
// @Description  Update an existing comment
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        id      path      string  true  "Comment ID"
// @Param        comment body      requests.UpdateCommentRequest  true  "Comment info"
// @Success      200     {object}  responses.CommonResponse
// @Failure      400     {object}  responses.CommonResponse
// @Failure      401     {object}  responses.CommonResponse
// @Failure      404     {object}  responses.CommonResponse
// @Failure      500     {object}  responses.CommonResponse
// @Router       /v1/comments/{id} [put]
func (c *CommentController) UpdateComment(ctx *fiber.Ctx) error {
	// Get authenticated user ID
	_, err := utils.GetUserID(ctx)
	if err != nil {
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

	// Update comment
	comment, err := c.commentService.UpdateComment(ctx.Context(), commentID, req.Content, "")
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

	return ctx.Status(fiber.StatusOK).JSON(responses.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Comment updated successfully",
		Data:            comment,
	})
}

// DeleteComment godoc
// @Summary      Delete comment
// @Description  Delete a comment by its ID
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Comment ID"
// @Success      200  {object}  responses.CommonResponse
// @Failure      400  {object}  responses.CommonResponse
// @Failure      401  {object}  responses.CommonResponse
// @Failure      404  {object}  responses.CommonResponse
// @Failure      500  {object}  responses.CommonResponse
// @Router       /v1/comments/{id} [delete]
func (c *CommentController) DeleteComment(ctx *fiber.Ctx) error {
	// Get authenticated user ID
	_, err := utils.GetUserID(ctx)
	if err != nil {
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
// @Summary      Approve comment
// @Description  Approve a comment by its ID
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Comment ID"
// @Success      200  {object}  responses.CommonResponse
// @Failure      400  {object}  responses.CommonResponse
// @Failure      401  {object}  responses.CommonResponse
// @Failure      404  {object}  responses.CommonResponse
// @Failure      500  {object}  responses.CommonResponse
// @Router       /v1/comments/{id}/approve [put]
func (c *CommentController) ApproveComment(ctx *fiber.Ctx) error {
	// Get authenticated user ID
	_, err := utils.GetUserID(ctx)
	if err != nil {
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

	return ctx.Status(fiber.StatusOK).JSON(responses.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Comment approved successfully",
		Data:            nil,
	})
}

// RejectComment godoc
// @Summary      Reject comment
// @Description  Reject a comment by its ID
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Comment ID"
// @Success      200  {object}  responses.CommonResponse
// @Failure      400  {object}  responses.CommonResponse
// @Failure      401  {object}  responses.CommonResponse
// @Failure      404  {object}  responses.CommonResponse
// @Failure      500  {object}  responses.CommonResponse
// @Router       /v1/comments/{id}/reject [put]
func (c *CommentController) RejectComment(ctx *fiber.Ctx) error {
	// Get authenticated user ID
	_, err := utils.GetUserID(ctx)
	if err != nil {
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
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    fiber.StatusBadRequest,
			ResponseMessage: "Invalid comment ID format",
			Data:            nil,
		})
	}

	// Reject comment
	err = c.commentService.RejectComment(ctx.Context(), commentID)
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
			ResponseMessage: "Failed to reject comment",
			Data:            nil,
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(responses.CommonResponse{
		ResponseCode:    fiber.StatusOK,
		ResponseMessage: "Comment rejected successfully",
		Data:            nil,
	})
}
