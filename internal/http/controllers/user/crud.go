package user

import (
	"errors"
	"github.com/google/uuid"
	"net/http"
	"webapi/internal/http/requests"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"webapi/internal/http/response"
	"webapi/internal/http/validation"
	"webapi/pkg/exception"
)

// GetUsers godoc
// @Summary Get all users with pagination
// @Description Retrieve a paginated list of users with optional search query
// @Tags users
// @Accept json
// @Produce json
// @Param limit query int false "Number of items per page (default: 10)"
// @Param page query int false "Page number (default: 1)"
// @Param query query string false "Search query for filtering users"
// @Success 200 {object} response.PaginationResponse{data=[]dto.GetUserDTO}
// @Failure 400 {object} response.CommonResponse
// @Failure 500 {object} response.CommonResponse
// @Router /v1/users [get]
func (h *UserHTTPHandler) GetUsers(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 10) // Default to 10 if not provided
	page := c.QueryInt("page", 1)    // Default to 1 if not provided
	query := c.Query("query", "")    // Default to empty string if not provided

	offset := (page - 1) * limit
	req := requests.DataWithPaginationRequest{
		Query: query,
		Limit: limit,
		Page:  offset,
	}
	responseUser, err := h.app.GetUsersWithPagination(c.Context(), req)
	if err != nil {
		return err
	}

	return c.JSON(response.PaginationResponse{
		TotalCount:   responseUser.Total,
		TotalPage:    responseUser.Total / limit,
		CurrentPage:  page,
		LastPage:     responseUser.LastPage,
		PerPage:      limit,
		NextPage:     page + 1,
		PreviousPage: page - 1,
		Data:         responseUser.Data,
		Path:         c.Path(),
	})
}

// GetUserByID godoc
// @Summary Get user by ID
// @Description Retrieve a specific user by their UUID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User UUID"
// @Success 200 {object} response.CommonResponse{data=dto.GetUserDTO}
// @Failure 400 {object} response.CommonResponse
// @Failure 404 {object} response.CommonResponse
// @Router /v1/users/{id} [get]
func (h *UserHTTPHandler) GetUserByID(c *fiber.Ctx) error {

	idParam := c.Params("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		return exception.InvalidIDError
	}

	userDto, err := h.app.GetUserByID(c.Context(), requests.GetUserIdRequest{ID: id})
	if err != nil {
		return err
	}

	return c.JSON(response.CommonResponse{
		ResponseCode:    http.StatusOK,
		ResponseMessage: "OK",
		Data:            userDto,
	})
}

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user with the provided information
// @Tags users
// @Accept json
// @Produce json
// @Param user body requests.UpdateUserRequest true "User information"
// @Success 201 {object} response.CommonResponse{data=dto.GetUserDTO}
// @Failure 400 {object} response.CommonResponse
// @Failure 422 {object} response.CommonResponse
// @Router /v1/users [post]
func (h *UserHTTPHandler) CreateUser(c *fiber.Ctx) error {
	var req requests.UpdateUserRequest

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
	dto, err := h.app.CreateUser(c.Context(), requests.CreateUserRequest{
		UserName: req.UserName,
		Email:    req.Email,
		Phone:    req.Phone,
	})

	if err != nil {
		return err
	}

	return c.Status(http.StatusCreated).JSON(response.CommonResponse{
		ResponseCode:    http.StatusCreated,
		ResponseMessage: "OK",
		Data:            dto,
	})
}

// UpdateUser godoc
// @Summary Update user information
// @Description Update an existing user's information by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User UUID"
// @Param user body requests.UpdateUserRequest true "Updated user information"
// @Success 200 {object} response.CommonResponse{data=dto.GetUserDTO}
// @Failure 400 {object} response.CommonResponse
// @Failure 404 {object} response.CommonResponse
// @Failure 422 {object} response.CommonResponse
// @Router /v1/users/{id} [put]
func (h *UserHTTPHandler) UpdateUser(c *fiber.Ctx) error {
	idParam := c.Params("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		return exception.InvalidIDError
	}

	var req requests.UpdateUserRequest

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

	user, err := h.app.GetUserByID(c.Context(), requests.GetUserIdRequest{ID: id})
	if err != nil {
		return exception.DataNotFoundError
	}

	dto, err := h.app.UpdateUser(c.Context(), user.ID, requests.UpdateUserRequest{
		UserName: req.UserName,
		Email:    req.Email,
		Phone:    req.Phone,
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

// DeleteUser godoc
// @Summary Delete user
// @Description Delete a user by their UUID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User UUID"
// @Success 200 {object} response.CommonResponse
// @Failure 400 {object} response.CommonResponse
// @Failure 404 {object} response.CommonResponse
// @Router /v1/users/{id} [delete]
func (h *UserHTTPHandler) DeleteUser(c *fiber.Ctx) error {
	idParam := c.Params("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		return exception.InvalidIDError
	}

	_, err = h.app.DeleteUser(c.Context(), requests.GetUserIdRequest{ID: id})
	if err != nil {
		return err
	}

	return c.JSON(response.CommonResponse{
		ResponseCode:    http.StatusOK,
		ResponseMessage: "OK",
	})
}
