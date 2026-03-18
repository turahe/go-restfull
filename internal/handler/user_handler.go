package handler

import (
	"net/http"
	"strconv"
	"strings"

	"go-rest/internal/middleware"
	"go-rest/internal/service"
	"go-rest/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserHandler struct {
	BaseHandler
	users *service.UserService
}

func NewUserHandler(users *service.UserService, log *zap.Logger) *UserHandler {
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
	if auth.Role != "admin" {
		response.Forbidden(c, response.BuildResponseCode(http.StatusForbidden, response.ServiceCodeUsers, response.CaseCodePermissionDenied), "forbidden", "admin only")
		return
	}

	limit := 100
	if s := strings.TrimSpace(c.Query("limit")); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil {
			response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeUsers, response.CaseCodeInvalidValue), "invalid limit", "limit must be int")
			return
		}
		limit = n
	}

	rows, err := h.users.List(c.Request.Context(), limit)
	if err != nil {
		h.internalError(c, response.ServiceCodeUsers, err, "list failed")
		return
	}
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeUsers, response.CaseCodeListRetrieved), "ok", rows)
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
	if auth.Role != "admin" {
		response.Forbidden(c, response.BuildResponseCode(http.StatusForbidden, response.ServiceCodeUsers, response.CaseCodePermissionDenied), "forbidden", "admin only")
		return
	}

	id, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeUsers, response.CaseCodeInvalidValue), "invalid id", "id must be uint")
		return
	}

	u, err := h.users.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			response.NotFound(c, response.BuildResponseCode(http.StatusNotFound, response.ServiceCodeUsers, response.CaseCodeNotFound), "not found", "user not found")
		default:
			h.internalError(c, response.ServiceCodeUsers, err, "get failed")
		}
		return
	}
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeUsers, response.CaseCodeRetrieved), "ok", u)
}

