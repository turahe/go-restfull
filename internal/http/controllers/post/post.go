package post

import (
	"webapi/internal/app/post"
	"webapi/internal/db/model"
	"webapi/internal/http/requests"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type PostHttpHandler struct {
	app post.PostApp
}

func NewPostHttpHandler(app post.PostApp) *PostHttpHandler {
	return &PostHttpHandler{app: app}
}

// CreatePost godoc
// @Summary Create a new post
// @Tags posts
// @Accept json
// @Produce json
// @Param post body model.Post true "Post info"
// @Success 201 {object} model.Post
// @Failure 400 {object} fiber.Map
// @Router /v1/posts [post]
func (h *PostHttpHandler) CreatePost(c *fiber.Ctx) error {
	var req requests.CreatePostRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}
	validator := requests.XValidator{}
	errs := validator.Validate(&req)
	if len(errs) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"validation_errors": errs})
	}

	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	tagIDs := make([]uuid.UUID, 0, len(req.Tags))
	for _, tagStr := range req.Tags {
		tagID, err := uuid.Parse(tagStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid tag ID: " + tagStr})
		}
		tagIDs = append(tagIDs, tagID)
	}

	post := &model.Post{
		ID:             uuid.New(),
		Slug:           req.Slug,
		Title:          req.Title,
		Subtitle:       req.Subtitle,
		Description:    req.Description,
		Type:           req.Type,
		IsSticky:       req.IsSticky,
		PublishedAt:    req.PublishedAt,
		Language:       req.Language,
		Layout:         req.Layout,
		RecordOrdering: req.RecordOrdering,
		CreatedBy:      &userID,
		UpdatedBy:      &userID,
	}
	if err := h.app.CreatePostWithTags(c.Context(), post, tagIDs); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(post)
}

// UpdatePost godoc
// @Summary Update a post
// @Tags posts
// @Accept json
// @Produce json
// @Param id path string true "Post UUID"
// @Param post body model.Post true "Post info"
// @Success 200 {object} model.Post
// @Failure 400 {object} fiber.Map
// @Router /v1/posts/{id} [put]
func (h *PostHttpHandler) UpdatePost(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}
	var req requests.UpdatePostRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}
	validator := requests.XValidator{}
	errs := validator.Validate(&req)
	if len(errs) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"validation_errors": errs})
	}
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}
	tagIDs := make([]uuid.UUID, 0, len(req.Tags))
	for _, tagStr := range req.Tags {
		tagID, err := uuid.Parse(tagStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid tag ID: " + tagStr})
		}
		tagIDs = append(tagIDs, tagID)
	}
	post := &model.Post{
		ID:             id,
		Title:          req.Title,
		Subtitle:       req.Subtitle,
		Description:    req.Description,
		Type:           req.Type,
		IsSticky:       req.IsSticky,
		PublishedAt:    req.PublishedAt,
		Language:       req.Language,
		Layout:         req.Layout,
		RecordOrdering: req.RecordOrdering,
		UpdatedBy:      &userID,
	}
	if err := h.app.UpdatePostWithTags(c.Context(), post, tagIDs); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(post)
}

// DeletePost godoc
// @Summary Delete a post
// @Tags posts
// @Accept json
// @Produce json
// @Param id path string true "Post UUID"
// @Success 200 {object} fiber.Map
// @Failure 400 {object} fiber.Map
// @Router /v1/posts/{id} [delete]
func (h *PostHttpHandler) DeletePost(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}
	if err := h.app.DeletePost(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Post deleted"})
}

// GetPostByID godoc
// @Summary Get post by ID
// @Tags posts
// @Accept json
// @Produce json
// @Param id path string true "Post UUID"
// @Success 200 {object} model.Post
// @Failure 404 {object} fiber.Map
// @Router /v1/posts/{id} [get]
func (h *PostHttpHandler) GetPostByID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}
	post, err := h.app.GetPostByIDWithContents(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(post)
}

// GetAllPosts godoc
// @Summary Get all posts
// @Tags posts
// @Accept json
// @Produce json
// @Success 200 {array} model.Post
// @Router /v1/posts [get]
func (h *PostHttpHandler) GetAllPosts(c *fiber.Ctx) error {
	posts, err := h.app.GetAllPostsWithContents(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(posts)
}
