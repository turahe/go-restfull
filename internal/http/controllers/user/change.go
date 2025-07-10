package user

import (
	"errors"
	"net/http"
	"webapi/internal/http/requests"
	"webapi/internal/http/response"
	"webapi/internal/http/validation"
	"webapi/pkg/exception"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ChangePassword godoc
// @Summary Change user password
// @Description Change the password for a specific user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User UUID"
// @Param password body requests.ChangePasswordRequest true "Password change information"
// @Success 200 {object} response.CommonResponse{data=dto.GetUserDTO}
// @Failure 400 {object} response.CommonResponse
// @Failure 404 {object} response.CommonResponse
// @Failure 422 {object} response.CommonResponse
// @Router /v1/users/change-password [post]
func (h *UserHTTPHandler) ChangePassword(c *fiber.Ctx) error {
	id := c.Params("id")
	userID, err := uuid.Parse(id)
	if err != nil {
		return exception.InvalidIDError
	}
	var req requests.ChangePasswordRequest
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
	user, err := h.app.GetUserByID(c.Context(), userID)
	if err != nil {
		return exception.DataNotFoundError
	}
	dto, err := h.app.ChangePassword(c.Context(), user.ID, requests.ChangePasswordRequest{
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		return err
	}
	return c.JSON(response.CommonResponse{
		ResponseCode:    http.StatusOK,
		ResponseMessage: "OK",
		Data:            dto,
	})
}

// ChangeUserName godoc
// @Summary Change user username
// @Description Change the username for a specific user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User UUID"
// @Param username body requests.ChangeUserNameRequest true "Username change information"
// @Success 200 {object} response.CommonResponse{data=dto.GetUserDTO}
// @Failure 400 {object} response.CommonResponse
// @Failure 404 {object} response.CommonResponse
// @Failure 422 {object} response.CommonResponse
// @Router /v1/users/change-username [post]
func (h *UserHTTPHandler) ChangeUserName(c *fiber.Ctx) error {
	idParam := c.Params("id")
	userId, err := uuid.Parse(idParam)
	if err != nil {
		return exception.InvalidIDError
	}
	var req requests.ChangeUserNameRequest
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
	user, err := h.app.GetUserByID(c.Context(), userId)
	if err != nil {
		return exception.DataNotFoundError
	}
	dto, err := h.app.ChangeUserName(c.Context(), user.ID, req.UserName)
	if err != nil {
		return err
	}
	return c.JSON(response.CommonResponse{
		ResponseCode:    http.StatusOK,
		ResponseMessage: "OK",
		Data:            dto,
	})

}

// ChangePhone godoc
// @Summary Change user phone number
// @Description Change the phone number for a specific user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User UUID"
// @Param phone body requests.ChangePhoneRequest true "Phone change information"
// @Success 200 {object} response.CommonResponse{data=dto.GetUserDTO}
// @Failure 400 {object} response.CommonResponse
// @Failure 404 {object} response.CommonResponse
// @Failure 422 {object} response.CommonResponse
// @Router /v1/users/change-phone [post]
func (h *UserHTTPHandler) ChangePhone(c *fiber.Ctx) error {
	id := c.Params("id")
	userID, err := uuid.Parse(id)
	if err != nil {
		return exception.InvalidIDError
	}
	var req requests.ChangePhoneRequest
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
	user, err := h.app.GetUserByID(c.Context(), userID)
	if err != nil {
		return exception.DataNotFoundError
	}
	dto, err := h.app.ChangePhone(c.Context(), user.ID, req.Phone)
	if err != nil {
		return err
	}
	return c.JSON(response.CommonResponse{
		ResponseCode:    http.StatusOK,
		ResponseMessage: "OK",
		Data:            dto,
	})
}

// ChangeEmail godoc
// @Summary Change user email
// @Description Change the email address for a specific user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User UUID"
// @Param email body requests.ChangeEmailRequest true "Email change information"
// @Success 200 {object} response.CommonResponse{data=dto.GetUserDTO}
// @Failure 400 {object} response.CommonResponse
// @Failure 404 {object} response.CommonResponse
// @Failure 422 {object} response.CommonResponse
// @Router /v1/users/change-email [post]
func (h *UserHTTPHandler) ChangeEmail(c *fiber.Ctx) error {
	id := c.Params("id")
	userID, err := uuid.Parse(id)
	if err != nil {
		return exception.InvalidIDError
	}
	var req requests.ChangeEmailRequest
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
	user, err := h.app.GetUserByID(c.Context(), userID)
	if err != nil {
		return exception.DataNotFoundError
	}
	dto, err := h.app.ChangeEmail(c.Context(), user.ID, req.Email)
	if err != nil {
		return err
	}
	return c.JSON(response.CommonResponse{
		ResponseCode:    http.StatusOK,
		ResponseMessage: "OK",
		Data:            dto,
	})
}
