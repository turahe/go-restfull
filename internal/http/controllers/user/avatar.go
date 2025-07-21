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

// UploadAvatar godoc
// @Summary Upload avatar
// @Description Upload avatar for a user
// @Tags users
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "User UUID"
// @Param avatar formData file true "Avatar file"
// @Success 200 {object} response.CommonResponse
// @Failure 400 {object} response.CommonResponse
// @Failure 401 {object} response.CommonResponse
// @Failure 404 {object} response.CommonResponse
// @Failure 500 {object} response.CommonResponse
func (h *UserHTTPHandler) UploadAvatar(c *fiber.Ctx) error {
	id := c.Params("id")
	userID, err := uuid.Parse(id)
	if err != nil {
		return exception.InvalidIDError
	}
	var req requests.ChangeAvatarRequest
	if err := c.BodyParser(&req); err != nil {

	}
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

	file, err := c.FormFile("avatar")
	if err != nil {
		return exception.InvalidRequestQueryParamError
	}
	userMedia, err := h.app.UploadAvatar(c.Context(), user.ID, file)

	return c.JSON(response.CommonResponse{
		ResponseCode:    http.StatusOK,
		ResponseMessage: "OK",
		Data:            userMedia,
	})

}
