package controllers

import (
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/turahe/go-restfull/internal/interfaces/http/responses"
	"github.com/turahe/go-restfull/internal/shared/errors"
)

// BaseController provides common functionality for all controllers
type BaseController struct{}

// NewBaseController creates a new base controller
func NewBaseController() *BaseController {
	return &BaseController{}
}

// handleError handles errors and returns appropriate HTTP responses
func (bc *BaseController) handleError(ctx *fiber.Ctx, err error) error {
	if domainErr := errors.GetDomainError(err); domainErr != nil {
		status := bc.getHTTPStatusFromDomainError(domainErr)
		return ctx.Status(status).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: domainErr.Message,
		})
	}

	// Generic error
	return ctx.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
		Status:  "error",
		Message: "An internal error occurred",
	})
}

// getHTTPStatusFromDomainError maps domain error codes to HTTP status codes
func (bc *BaseController) getHTTPStatusFromDomainError(err *errors.DomainError) int {
	switch err.Code {
	case errors.ValidationErrorCode, errors.InvalidEmailErrorCode,
		errors.InvalidPhoneErrorCode, errors.InvalidPasswordErrorCode,
		errors.RequiredFieldErrorCode:
		return http.StatusBadRequest
	case errors.EmailAlreadyExistsCode, errors.UsernameAlreadyExistsCode,
		errors.PhoneAlreadyExistsCode, errors.EmailAlreadyVerifiedCode,
		errors.PhoneAlreadyVerifiedCode, errors.RoleAlreadyAssignedCode:
		return http.StatusConflict
	case errors.UserNotFoundCode, errors.RoleNotFoundCode, errors.NotFoundErrorCode:
		return http.StatusNotFound
	case errors.UnauthorizedErrorCode, errors.InvalidCredentialsCode:
		return http.StatusUnauthorized
	case errors.ForbiddenErrorCode:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

// validateRequest validates a request struct and returns error if validation fails
func (bc *BaseController) validateRequest(ctx *fiber.Ctx, req interface{}) error {
	if err := ctx.BodyParser(req); err != nil {
		return errors.NewDomainError("INVALID_REQUEST", "Invalid request body")
	}

	// Basic validation - in practice, you'd use a proper validation library
	return nil
}

// parseUUID parses a UUID from a parameter and returns error if invalid
func (bc *BaseController) parseUUID(param string) (string, error) {
	if param == "" {
		return "", errors.NewDomainError("INVALID_ID", "ID is required")
	}

	// Basic UUID validation - in practice, you'd use a proper UUID library
	if len(param) != 36 {
		return "", errors.NewDomainError("INVALID_ID", "Invalid ID format")
	}

	return param, nil
}

// getPaginationParams extracts and validates pagination parameters
func (bc *BaseController) getPaginationParams(ctx *fiber.Ctx) (page, pageSize int) {
	pageStr := ctx.Query("page", "1")
	pageSizeStr := ctx.Query("page_size", "10")

	page, _ = strconv.Atoi(pageStr)
	pageSize, _ = strconv.Atoi(pageSizeStr)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	return page, pageSize
}

// getSortParams extracts and validates sorting parameters
func (bc *BaseController) getSortParams(ctx *fiber.Ctx) (sortBy, sortDir string) {
	sortBy = ctx.Query("sort_by", "created_at")
	sortDir = ctx.Query("sort_dir", "desc")

	// Validate sort direction
	if sortDir != "asc" && sortDir != "desc" {
		sortDir = "desc"
	}

	return sortBy, sortDir
}

// successResponse returns a success response
func (bc *BaseController) successResponse(ctx *fiber.Ctx, data interface{}) error {
	return ctx.JSON(responses.SuccessResponse{
		Data: data,
	})
}

// createdResponse returns a created response
func (bc *BaseController) createdResponse(ctx *fiber.Ctx, data interface{}) error {
	return ctx.Status(http.StatusCreated).JSON(responses.SuccessResponse{
		Data: data,
	})
}

// noContentResponse returns a no content response
func (bc *BaseController) noContentResponse(ctx *fiber.Ctx) error {
	return ctx.SendStatus(http.StatusNoContent)
}
