package controllers

import (
	"net/http"

	"webapi/internal/application/ports"
	"webapi/internal/interfaces/http/requests"
	"webapi/internal/interfaces/http/responses"
	"webapi/internal/router/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// PostController handles HTTP requests for post operations
// @title Post Management API
// @version 1.0
// @description This is a post management API for creating, reading, updating, and deleting blog posts
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email support@example.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8000
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
type PostController struct {
	postService ports.PostService
}

// NewPostController creates a new post controller
func NewPostController(postService ports.PostService) *PostController {
	return &PostController{
		postService: postService,
	}
}

// CreatePost handles POST /posts
// @Summary Create a new post
// @Description Create a new blog post with the provided information
// @Tags posts
// @Accept json
// @Produce json
// @Param post body requests.CreatePostRequest true "Post creation request"
// @Success 201 {object} responses.SuccessResponse{data=responses.PostResponse} "Post created successfully"
// @Failure 400 {object} responses.ErrorResponse "Bad request - Invalid input data"
// @Failure 409 {object} responses.ErrorResponse "Conflict - Post with same slug already exists"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /posts [post]
func (c *PostController) CreatePost(ctx *fiber.Ctx) error {
	var req requests.CreatePostRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid request body",
		})
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	// Create post
	post, err := c.postService.CreatePost(ctx.Context(), req.Title, req.Content, req.Slug, req.Status, req.AuthorID)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusCreated).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   responses.NewPostResponse(post),
	})
}

// GetPostByID handles GET /posts/:id
// @Summary Get post by ID
// @Description Retrieve a post by its unique identifier
// @Tags posts
// @Accept json
// @Produce json
// @Param id path string true "Post ID" format(uuid)
// @Success 200 {object} responses.SuccessResponse{data=responses.PostResponse} "Post found"
// @Failure 400 {object} responses.ErrorResponse "Bad request - Invalid post ID"
// @Failure 404 {object} responses.ErrorResponse "Not found - Post does not exist"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /posts/{id} [get]
func (c *PostController) GetPostByID(ctx *fiber.Ctx) error {
	idParam := ctx.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid post ID",
		})
	}

	post, err := c.postService.GetPostByID(ctx.Context(), id)
	if err != nil {
		return ctx.Status(http.StatusNotFound).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Post not found",
		})
	}

	return ctx.JSON(responses.SuccessResponse{
		Status: "success",
		Data:   responses.NewPostResponse(post),
	})
}

// GetPostBySlug handles GET /posts/slug/:slug
// @Summary Get post by slug
// @Description Retrieve a post by its URL-friendly slug
// @Tags posts
// @Accept json
// @Produce json
// @Param slug path string true "Post slug"
// @Success 200 {object} responses.SuccessResponse{data=responses.PostResponse} "Post found"
// @Failure 400 {object} responses.ErrorResponse "Bad request - Slug is required"
// @Failure 404 {object} responses.ErrorResponse "Not found - Post does not exist"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /posts/slug/{slug} [get]
func (c *PostController) GetPostBySlug(ctx *fiber.Ctx) error {
	slug := ctx.Params("slug")
	if slug == "" {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Slug is required",
		})
	}

	post, err := c.postService.GetPostBySlug(ctx.Context(), slug)
	if err != nil {
		return ctx.Status(http.StatusNotFound).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Post not found",
		})
	}

	return ctx.JSON(responses.SuccessResponse{
		Status: "success",
		Data:   responses.NewPostResponse(post),
	})
}

// GetPosts handles GET /posts
// @Summary Get all posts
// @Description Retrieve a paginated list of posts with optional search and status filtering
// @Tags posts
// @Accept json
// @Produce json
// @Param limit query int false "Number of posts to return (default: 10, max: 100)" default(10) minimum(1) maximum(100)
// @Param offset query int false "Number of posts to skip (default: 0)" default(0) minimum(0)
// @Param query query string false "Search query to filter posts by title or content"
// @Param status query string false "Filter posts by status (published, draft, etc.)" Enums(published, draft, archived)
// @Success 200 {object} responses.SuccessResponse{data=[]responses.PostResponse} "List of posts"
// @Failure 400 {object} responses.ErrorResponse "Bad request - Invalid parameters"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /posts [get]
func (c *PostController) GetPosts(ctx *fiber.Ctx) error {
	// Get pagination parameters from middleware
	pagination := middleware.GetPaginationParams(ctx)

	// Get additional filters
	status := ctx.Query("status", "")

	// Use the service layer pagination method
	posts, total, err := c.postService.GetPostsWithPagination(ctx.Context(), pagination.Page, pagination.PerPage, pagination.Search, status)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	// Convert to response DTOs
	postResponses := make([]responses.PostResponse, len(posts))
	for i, post := range posts {
		postResponses[i] = *responses.NewPostResponse(post)
	}

	// Create paginated response using helper
	paginatedResult := responses.CreatePaginatedResult(postResponses, pagination.Page, pagination.PerPage, total)

	return ctx.JSON(responses.SuccessResponse{
		Status: "success",
		Data:   paginatedResult,
	})
}

// GetPostsByAuthor handles GET /posts/author/:authorID
// @Summary Get posts by author
// @Description Retrieve all posts written by a specific author
// @Tags posts
// @Accept json
// @Produce json
// @Param authorID path string true "Author ID" format(uuid)
// @Param limit query int false "Number of posts to return (default: 10, max: 100)" default(10) minimum(1) maximum(100)
// @Param offset query int false "Number of posts to skip (default: 0)" default(0) minimum(0)
// @Success 200 {object} responses.SuccessResponse{data=[]responses.PostResponse} "List of posts by author"
// @Failure 400 {object} responses.ErrorResponse "Bad request - Invalid author ID"
// @Failure 404 {object} responses.ErrorResponse "Not found - Author does not exist"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /posts/author/{authorID} [get]
func (c *PostController) GetPostsByAuthor(ctx *fiber.Ctx) error {
	authorIDParam := ctx.Params("authorID")
	authorID, err := uuid.Parse(authorIDParam)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid author ID",
		})
	}

	// Get pagination parameters from middleware
	pagination := middleware.GetPaginationParams(ctx)
	offset := middleware.GetOffset(ctx)

	posts, err := c.postService.GetPostsByAuthor(ctx.Context(), authorID, pagination.PerPage, offset)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	// Convert to response DTOs
	postResponses := make([]responses.PostResponse, len(posts))
	for i, post := range posts {
		postResponses[i] = *responses.NewPostResponse(post)
	}

	// For now, use simple count. In real implementation, get total count
	total := int64(len(posts))

	// Create paginated response using helper
	paginatedResult := responses.CreatePaginatedResult(postResponses, pagination.Page, pagination.PerPage, total)

	return ctx.JSON(responses.SuccessResponse{
		Status: "success",
		Data:   paginatedResult,
	})
}

// UpdatePost handles PUT /posts/:id
// @Summary Update post
// @Description Update an existing post's information
// @Tags posts
// @Accept json
// @Produce json
// @Param id path string true "Post ID" format(uuid)
// @Param post body requests.UpdatePostRequest true "Post update request"
// @Success 200 {object} responses.SuccessResponse{data=responses.PostResponse} "Post updated successfully"
// @Failure 400 {object} responses.ErrorResponse "Bad request - Invalid input data"
// @Failure 404 {object} responses.ErrorResponse "Not found - Post does not exist"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /posts/{id} [put]
func (c *PostController) UpdatePost(ctx *fiber.Ctx) error {
	idParam := ctx.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid post ID",
		})
	}

	var req requests.UpdatePostRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid request body",
		})
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	// Update post
	post, err := c.postService.UpdatePost(ctx.Context(), id, req.Title, req.Content, req.Slug, req.Status)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.JSON(responses.SuccessResponse{
		Status: "success",
		Data:   responses.NewPostResponse(post),
	})
}

// DeletePost handles DELETE /posts/:id
// @Summary Delete post
// @Description Delete a post (soft delete)
// @Tags posts
// @Accept json
// @Produce json
// @Param id path string true "Post ID" format(uuid)
// @Success 200 {object} responses.SuccessResponse "Post deleted successfully"
// @Failure 400 {object} responses.ErrorResponse "Bad request - Invalid post ID"
// @Failure 404 {object} responses.ErrorResponse "Not found - Post does not exist"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /posts/{id} [delete]
func (c *PostController) DeletePost(ctx *fiber.Ctx) error {
	idParam := ctx.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid post ID",
		})
	}

	err = c.postService.DeletePost(ctx.Context(), id)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.JSON(responses.SuccessResponse{
		Status:  "success",
		Message: "Post deleted successfully",
	})
}

// PublishPost handles PUT /posts/:id/publish
// @Summary Publish post
// @Description Publish a draft post to make it publicly visible
// @Tags posts
// @Accept json
// @Produce json
// @Param id path string true "Post ID" format(uuid)
// @Success 200 {object} responses.SuccessResponse{data=responses.PostResponse} "Post published successfully"
// @Failure 400 {object} responses.ErrorResponse "Bad request - Invalid post ID"
// @Failure 404 {object} responses.ErrorResponse "Not found - Post does not exist"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /posts/{id}/publish [put]
func (c *PostController) PublishPost(ctx *fiber.Ctx) error {
	idParam := ctx.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid post ID",
		})
	}

	err = c.postService.PublishPost(ctx.Context(), id)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.JSON(responses.SuccessResponse{
		Status:  "success",
		Message: "Post published successfully",
	})
}

// UnpublishPost handles PUT /posts/:id/unpublish
// @Summary Unpublish post
// @Description Unpublish a post to make it a draft (not publicly visible)
// @Tags posts
// @Accept json
// @Produce json
// @Param id path string true "Post ID" format(uuid)
// @Success 200 {object} responses.SuccessResponse{data=responses.PostResponse} "Post unpublished successfully"
// @Failure 400 {object} responses.ErrorResponse "Bad request - Invalid post ID"
// @Failure 404 {object} responses.ErrorResponse "Not found - Post does not exist"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /posts/{id}/unpublish [put]
func (c *PostController) UnpublishPost(ctx *fiber.Ctx) error {
	idParam := ctx.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid post ID",
		})
	}

	err = c.postService.UnpublishPost(ctx.Context(), id)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return ctx.JSON(responses.SuccessResponse{
		Status:  "success",
		Message: "Post unpublished successfully",
	})
}
