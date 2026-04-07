package handler

import (
	"context"
	"net/http"

	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/handler/request"
	"github.com/turahe/go-restfull/internal/middleware"
	"github.com/turahe/go-restfull/internal/model"
	"github.com/turahe/go-restfull/internal/repository"
	"github.com/turahe/go-restfull/internal/service"
	"github.com/turahe/go-restfull/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserHandler struct {
	BaseHandler
	users UserService
}

type UserService interface {
	List(ctx context.Context, req request.UserListRequest) (repository.CursorPage, error)
	GetByID(ctx context.Context, id uint) (*model.User, error)
	Create(ctx context.Context, req request.CreateUserRequest) (*service.UserCreateOutcome, error)
}

func NewUserHandler(users UserService, log *zap.Logger) *UserHandler {
	return &UserHandler{BaseHandler: BaseHandler{Log: log}, users: users}
}

// ListUsers godoc
// @Summary      List users
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Param        limit  query     int  false  "Max items (max 500)"
// @Success      200    {object}  response.Envelope
// @Failure      401    {object}  response.Envelope
// @Failure      403    {object}  response.Envelope
// @Failure      500    {object}  response.Envelope
// @Router       /api/v1/users [get]
func (h *UserHandler) List(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok {
		response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeUsers, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}
	if auth.Role != entities.RoleAdmin {
		response.Forbidden(c, response.BuildResponseCode(http.StatusForbidden, response.ServiceCodeUsers, response.CaseCodePermissionDenied), "forbidden", "admin only")
		return
	}

	var req request.UserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeUsers, response.CaseCodeInvalidFormat), "invalid request", err.Error())
		return
	}
	if !h.validate(c, response.ServiceCodeUsers, req) {
		return
	}

	page, err := h.users.List(c.Request.Context(), req)
	if err != nil {
		h.internalError(c, response.ServiceCodeUsers, err, "list failed")
		return
	}
	next := page.NextCursor != nil
	prev := page.PrevCursor != nil
	response.OKPaginated(c,
		response.BuildResponseCode(http.StatusOK, response.ServiceCodeUsers, response.CaseCodeListRetrieved),
		"Successfully retrieved users",
		page.Items,
		next,
		prev,
	)
}

// CreateUser godoc
// @Summary      Create user (admin)
// @Description  Creates a user account. Same shape as public register; role defaults to entities.RoleUser. Response data includes id, name, email, roleId (roles.id).
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      request.CreateUserRequest  true  "Create user payload"
// @Success      201   {object}  response.Envelope
// @Failure      400   {object}  response.Envelope
// @Failure      401   {object}  response.Envelope
// @Failure      403   {object}  response.Envelope
// @Failure      500   {object}  response.Envelope
// @Router       /api/v1/users [post]
func (h *UserHandler) Create(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok {
		response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeUsers, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}
	if auth.Role != entities.RoleAdmin {
		response.Forbidden(c, response.BuildResponseCode(http.StatusForbidden, response.ServiceCodeUsers, response.CaseCodePermissionDenied), "forbidden", "admin only")
		return
	}

	var req request.CreateUserRequest
	if !h.bindJSON(c, response.ServiceCodeUsers, &req) {
		return
	}
	if !h.validate(c, response.ServiceCodeUsers, req) {
		return
	}

	out, err := h.users.Create(c.Request.Context(), req)
	if err != nil {
		switch err {
		case service.ErrEmailTaken:
			response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeUsers, response.CaseCodeDuplicateEntry), "email already registered", "email taken")
			return
		case service.ErrRoleNotFound:
			response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeUsers, response.CaseCodeInvalidValue), "invalid role", "role not found")
			return
		}
		h.internalError(c, response.ServiceCodeUsers, err, "create user failed")
		return
	}

	data := gin.H{
		"id":     out.User.ID,
		"name":   out.User.Name,
		"email":  out.User.Email,
		"roleId": out.RoleID,
	}

	response.Created(c,
		response.BuildResponseCode(http.StatusCreated, response.ServiceCodeUsers, response.CaseCodeCreated),
		"Successfully created user",
		data)
}

// GetUserByID godoc
// @Summary      Get user by id
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Param        id  path      int  true  "User ID"
// @Success      200 {object}  response.Envelope
// @Failure      400 {object}  response.Envelope
// @Failure      401 {object}  response.Envelope
// @Failure      403 {object}  response.Envelope
// @Failure      404 {object}  response.Envelope
// @Failure      500 {object}  response.Envelope
// @Router       /api/v1/users/{id} [get]
func (h *UserHandler) GetByID(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok {
		response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeUsers, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}
	if auth.Role != entities.RoleAdmin {
		response.Forbidden(c, response.BuildResponseCode(http.StatusForbidden, response.ServiceCodeUsers, response.CaseCodePermissionDenied), "forbidden", "admin only")
		return
	}

	id, err := h.ParseUintParam(c, "id")
	if err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeUsers, response.CaseCodeInvalidValue), "invalid id", "id must be uint")
		return
	}

	u, err := h.users.GetByID(c.Request.Context(), id)
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			response.NotFound(c, response.BuildResponseCode(http.StatusNotFound, response.ServiceCodeUsers, response.CaseCodeNotFound), "not found", "user not found")
		default:
			h.internalError(c, response.ServiceCodeUsers, err, "get failed")
		}
		return
	}
	response.OK(c,
		response.BuildResponseCode(http.StatusOK, response.ServiceCodeUsers, response.CaseCodeRetrieved),
		"Successfully retrieved user",
		u,
	)
}
