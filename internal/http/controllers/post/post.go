package post

import (
	"webapi/internal/app/post"
	"webapi/internal/db/model"
	"webapi/internal/helper/utils"
	"webapi/internal/http/requests"
	"webapi/internal/http/response"

	"github.com/gofiber/fiber/v2"
	"github.com/gomarkdown/markdown"
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
// @Failure 400 {object} response.CommonResponse
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

	slug := req.Slug
	if slug == "" {
		slug = utils.Slugify(req.Title)
	}

	contentHTML := string(markdown.ToHTML([]byte(req.Content), nil, nil))
	content := model.Content{
		ID:          uuid.New().String(),
		ModelType:   "post",
		ModelID:     "", // will be set in repo to post.ID
		ContentRaw:  req.Content,
		ContentHTML: contentHTML,
		CreatedBy:   userID.String(),
		UpdatedBy:   userID.String(),
	}

	recordOrdering := req.RecordOrdering
	if recordOrdering == 0 {
		// Fetch max record_ordering from repository
		maxOrdering, err := h.app.GetMaxPostRecordOrdering(c.Context())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get max record ordering"})
		}
		recordOrdering = maxOrdering + 1
	}

	post := &model.Post{
		ID:             uuid.New(),
		Slug:           slug,
		Title:          req.Title,
		Subtitle:       req.Subtitle,
		Description:    req.Description,
		Type:           req.Type,
		IsSticky:       req.IsSticky,
		PublishedAt:    req.PublishedAt,
		Language:       req.Language,
		Layout:         req.Layout,
		RecordOrdering: recordOrdering,
		CreatedBy:      &userID,
		UpdatedBy:      &userID,
		Contents:       []model.Content{content},
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
// @Success 200 {object} response.CommonResponse
// @Failure 400 {object} response.CommonResponse
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

	slug := req.Slug
	if slug == "" {
		slug = utils.Slugify(req.Title)
	}

	contentHTML := string(markdown.ToHTML([]byte(req.Content), nil, nil))
	content := model.Content{
		ID:          uuid.New().String(),
		ModelType:   "post",
		ModelID:     id.String(),
		ContentRaw:  req.Content,
		ContentHTML: contentHTML,
		CreatedBy:   userID.String(),
		UpdatedBy:   userID.String(),
	}

	recordOrdering := req.RecordOrdering
	if recordOrdering == 0 {
		maxOrdering, err := h.app.GetMaxPostRecordOrdering(c.Context())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get max record ordering"})
		}
		recordOrdering = maxOrdering + 1
	}

	post := &model.Post{
		ID:             id,
		Slug:           slug,
		Title:          req.Title,
		Subtitle:       req.Subtitle,
		Description:    req.Description,
		Type:           req.Type,
		IsSticky:       req.IsSticky,
		PublishedAt:    req.PublishedAt,
		Language:       req.Language,
		Layout:         req.Layout,
		RecordOrdering: recordOrdering,
		UpdatedBy:      &userID,
		Contents:       []model.Content{content},
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
// @Success 200 {object} response.CommonResponse
// @Failure 400 {object} response.CommonResponse
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
// @Failure 404 {object} response.CommonResponse
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
	limit := c.QueryInt("limit", 10)
	page := c.QueryInt("page", 1)
	query := c.Query("query", "")

	offset := (page - 1) * limit
	req := requests.DataWithPaginationRequest{
		Query: query,
		Limit: limit,
		Page:  offset,
	}
	posts, total, err := h.app.GetPostsWithPagination(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	totalPage := 1
	if limit > 0 {
		totalPage = (total + limit - 1) / limit
	}
	return c.JSON(response.PaginationResponse{
		TotalCount:   total,
		TotalPage:    totalPage,
		CurrentPage:  page,
		LastPage:     totalPage,
		PerPage:      limit,
		NextPage:     page + 1,
		PreviousPage: page - 1,
		Data:         posts,
		Path:         c.Path(),
	})
}
