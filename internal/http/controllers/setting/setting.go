package setting

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"webapi/internal/app/setting"
	"webapi/internal/http/requests"
	"webapi/internal/http/response"
	"webapi/internal/http/validation"
	"webapi/pkg/exception"
)

type SettingHTTPHandler struct {
	app setting.SettingApp
}

func NewSettingHTTPHandler(app setting.SettingApp) *SettingHTTPHandler {
	return &SettingHTTPHandler{app: app}
}

// CreateSetting godoc
// @Summary Create a new setting
// @Description Create a new setting with the provided information
// @Tags settings
// @Accept json
// @Produce json
// @Param setting body requests.CreateSettingRequest true "Setting information"
// @Success 201 {object} response.CommonResponse{data=dto.SettingDTO}
// @Failure 400 {object} response.CommonResponse
// @Failure 422 {object} response.CommonResponse
// @Router /v1/settings [post]
func (h *SettingHTTPHandler) CreateSetting(c *fiber.Ctx) error {
	var req requests.CreateSettingRequest

	// Parse the request body
	if err := c.BodyParser(&req); err != nil {
		return exception.InvalidRequestBodyError
	}

	// Validate the request body
	v, _ := validation.GetValidator()
	if err := v.Struct(req); err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			return exception.NewValidationFailedErrors(validationErrs)
		}
	}

	// Process the business logic
	dto, err := h.app.CreateSetting(c.Context(), req)
	if err != nil {
		return err
	}

	return c.Status(http.StatusCreated).JSON(response.CommonResponse{
		ResponseCode:    http.StatusCreated,
		ResponseMessage: "Setting created successfully",
		Data:            dto,
	})
}

// GetSettingByKey godoc
// @Summary Get setting by key
// @Description Retrieve a specific setting by its key
// @Tags settings
// @Accept json
// @Produce json
// @Param key path string true "Setting key"
// @Success 200 {object} response.CommonResponse{data=dto.SettingDTO}
// @Failure 400 {object} response.CommonResponse
// @Failure 404 {object} response.CommonResponse
// @Router /v1/settings/{key} [get]
func (h *SettingHTTPHandler) GetSettingByKey(c *fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return exception.InvalidIDError
	}

	req := requests.GetSettingByKeyRequest{Key: key}
	dto, err := h.app.GetSettingByKey(c.Context(), req)
	if err != nil {
		return err
	}

	return c.JSON(response.CommonResponse{
		ResponseCode:    http.StatusOK,
		ResponseMessage: "Setting retrieved successfully",
		Data:            dto,
	})
}

// GetAllSettings godoc
// @Summary Get all settings
// @Description Retrieve a list of all settings
// @Tags settings
// @Accept json
// @Produce json
// @Success 200 {object} response.CommonResponse{data=[]dto.SettingDTO}
// @Failure 500 {object} response.CommonResponse
// @Router /v1/settings [get]
func (h *SettingHTTPHandler) GetAllSettings(c *fiber.Ctx) error {
	settings, err := h.app.GetAllSettings(c.Context())
	if err != nil {
		return err
	}

	return c.JSON(response.CommonResponse{
		ResponseCode:    http.StatusOK,
		ResponseMessage: "Settings retrieved successfully",
		Data:            settings,
	})
}

// UpdateSetting godoc
// @Summary Update setting
// @Description Update an existing setting by key
// @Tags settings
// @Accept json
// @Produce json
// @Param key path string true "Setting key"
// @Param setting body requests.UpdateSettingRequest true "Updated setting information"
// @Success 200 {object} response.CommonResponse{data=dto.SettingDTO}
// @Failure 400 {object} response.CommonResponse
// @Failure 404 {object} response.CommonResponse
// @Failure 422 {object} response.CommonResponse
// @Router /v1/settings/{key} [put]
func (h *SettingHTTPHandler) UpdateSetting(c *fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return exception.InvalidIDError
	}

	var req requests.UpdateSettingRequest

	// Parse the request body
	if err := c.BodyParser(&req); err != nil {
		return exception.InvalidRequestBodyError
	}

	// Validate the request body
	v, _ := validation.GetValidator()
	if err := v.Struct(req); err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			return exception.NewValidationFailedErrors(validationErrs)
		}
	}

	// Process the business logic
	dto, err := h.app.UpdateSetting(c.Context(), req)
	if err != nil {
		return err
	}

	return c.JSON(response.CommonResponse{
		ResponseCode:    http.StatusOK,
		ResponseMessage: "Setting updated successfully",
		Data:            dto,
	})
}

// DeleteSetting godoc
// @Summary Delete setting
// @Description Delete a setting by its key
// @Tags settings
// @Accept json
// @Produce json
// @Param key path string true "Setting key"
// @Success 200 {object} response.CommonResponse
// @Failure 400 {object} response.CommonResponse
// @Failure 404 {object} response.CommonResponse
// @Router /v1/settings/{key} [delete]
func (h *SettingHTTPHandler) DeleteSetting(c *fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return exception.InvalidIDError
	}

	err := h.app.DeleteSetting(c.Context(), key)
	if err != nil {
		return err
	}

	return c.JSON(response.CommonResponse{
		ResponseCode:    http.StatusOK,
		ResponseMessage: "Setting deleted successfully",
	})
} 