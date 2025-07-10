package tag

import (
	"webapi/internal/app/tag"
	"webapi/internal/db/model"
	"webapi/internal/helper/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TagHttpHandler struct {
	app tag.TagApp
}

func NewTagHttpHandler(app tag.TagApp) *TagHttpHandler {
	return &TagHttpHandler{app: app}
}

// CreateTag godoc
// @Summary Create a new tag
// @Tags tags
// @Accept json
// @Produce json
// @Param tag body model.Tag true "Tag info"
// @Success 200 {object} model.Tag
// @Failure 400 {object} fiber.Map
// @Router /v1/tags [post]
func (h *TagHttpHandler) CreateTag(c *fiber.Ctx) error {
	tag := new(model.Tag)
	if err := c.BodyParser(tag); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}
	userID, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}
	tag.CreatedBy = userID.String()
	tag.UpdatedBy = userID.String()
	if err := h.app.CreateTag(c.Context(), tag); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(tag)
}

// GetTagByID godoc
// @Summary Get tag by ID
// @Tags tags
// @Accept json
// @Produce json
// @Param id path string true "Tag UUID"
// @Success 200 {object} model.Tag
// @Failure 404 {object} fiber.Map
// @Router /v1/tags/{id} [get]
func (h *TagHttpHandler) GetTagByID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}
	tag, err := h.app.GetTagByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(tag)
}

// GetAllTags godoc
// @Summary Get all tags
// @Tags tags
// @Accept json
// @Produce json
// @Success 200 {array} model.Tag
// @Router /v1/tags [get]
func (h *TagHttpHandler) GetAllTags(c *fiber.Ctx) error {
	tags, err := h.app.GetAllTags(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(tags)
}

// UpdateTag godoc
// @Summary Update a tag
// @Tags tags
// @Accept json
// @Produce json
// @Param id path string true "Tag UUID"
// @Param tag body model.Tag true "Tag info"
// @Success 200 {object} model.Tag
// @Failure 400 {object} fiber.Map
// @Router /v1/tags/{id} [put]
func (h *TagHttpHandler) UpdateTag(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}
	tag := new(model.Tag)
	if err := c.BodyParser(tag); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}
	tag.ID = id.String()
	userID, err := utils.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}
	tag.UpdatedBy = userID.String()
	if err := h.app.UpdateTag(c.Context(), tag); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(tag)
}

// DeleteTag godoc
// @Summary Delete a tag
// @Tags tags
// @Accept json
// @Produce json
// @Param id path string true "Tag UUID"
// @Success 200 {object} fiber.Map
// @Failure 400 {object} fiber.Map
// @Router /v1/tags/{id} [delete]
func (h *TagHttpHandler) DeleteTag(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}
	if err := h.app.DeleteTag(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Tag deleted"})
}
