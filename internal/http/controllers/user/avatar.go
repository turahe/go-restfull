package user

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"net/http"
	"webapi/internal/http/requests"
	"webapi/internal/http/response"
	"webapi/internal/http/validation"
	"webapi/pkg/exception"
)

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

	user, err := h.app.GetUserByID(c.Context(), requests.GetUserIdRequest{ID: userID})
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
