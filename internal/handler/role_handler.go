package handler

import (
	"net/http"
	"strconv"
	"strings"

	"go-rest/internal/handler/request"
	"go-rest/internal/middleware"
	"go-rest/internal/service"
	"go-rest/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RoleHandler struct {
	BaseHandler
	roles *service.RoleService
}

func NewRoleHandler(roles *service.RoleService, log *zap.Logger) *RoleHandler {
	return &RoleHandler{BaseHandler: BaseHandler{Log: log}, roles: roles}
}

// ListRoles godoc
// @Summary      List roles
// @Tags         Roles
// @Produce      json
// @Security     BearerAuth
// @Param        limit  query     int  false  "Max items (max 500)"
// @Success      200    {object}  response.Envelope
// @Failure      401    {object}  response.Envelope
// @Failure      403    {object}  response.Envelope
// @Failure      500    {object}  response.Envelope
// @Router       /api/v1/roles [get]
func (h *RoleHandler) List(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok {
		response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeRoles, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}
	if auth.Role != "admin" {
		response.Forbidden(c, response.BuildResponseCode(http.StatusForbidden, response.ServiceCodeRoles, response.CaseCodePermissionDenied), "forbidden", "admin only")
		return
	}

	limit := 200
	if s := strings.TrimSpace(c.Query("limit")); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil {
			response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeRoles, response.CaseCodeInvalidValue), "invalid limit", "limit must be int")
			return
		}
		limit = n
	}

	rows, err := h.roles.List(c.Request.Context(), limit)
	if err != nil {
		h.internalError(c, response.ServiceCodeRoles, err, "list failed")
		return
	}
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeRoles, response.CaseCodeListRetrieved), "ok", rows)
}

// CreateRole godoc
// @Summary      Create role
// @Tags         Roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      request.CreateRoleRequest  true  "Create role payload"
// @Success      201   {object}  response.Envelope
// @Failure      400   {object}  response.Envelope
// @Failure      401   {object}  response.Envelope
// @Failure      403   {object}  response.Envelope
// @Failure      500   {object}  response.Envelope
// @Router       /api/v1/roles [post]
func (h *RoleHandler) Create(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok {
		response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeRoles, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}
	if auth.Role != "admin" {
		response.Forbidden(c, response.BuildResponseCode(http.StatusForbidden, response.ServiceCodeRoles, response.CaseCodePermissionDenied), "forbidden", "admin only")
		return
	}

	var req request.CreateRoleRequest
	if !h.bindJSON(c, response.ServiceCodeRoles, &req) {
		return
	}
	if !h.validate(c, response.ServiceCodeRoles, req) {
		return
	}

	role, err := h.roles.Create(c.Request.Context(), req.Name)
	if err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeRoles, response.CaseCodeInvalidValue), "invalid request", err.Error())
		return
	}
	response.Created(c, response.BuildResponseCode(http.StatusCreated, response.ServiceCodeRoles, response.CaseCodeCreated), "created", role)
}

// DeleteRole godoc
// @Summary      Delete role
// @Tags         Roles
// @Produce      json
// @Security     BearerAuth
// @Param        id  path      int  true  "Role ID"
// @Success      200 {object}  response.Envelope
// @Failure      400 {object}  response.Envelope
// @Failure      401 {object}  response.Envelope
// @Failure      403 {object}  response.Envelope
// @Failure      404 {object}  response.Envelope
// @Failure      500 {object}  response.Envelope
// @Router       /api/v1/roles/{id} [delete]
func (h *RoleHandler) Delete(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok {
		response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeRoles, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}
	if auth.Role != "admin" {
		response.Forbidden(c, response.BuildResponseCode(http.StatusForbidden, response.ServiceCodeRoles, response.CaseCodePermissionDenied), "forbidden", "admin only")
		return
	}

	id, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeRoles, response.CaseCodeInvalidValue), "invalid id", "id must be uint")
		return
	}

	err = h.roles.Delete(c.Request.Context(), uint(id))
	if err != nil {
		switch err {
		case service.ErrRoleNotFound:
			response.NotFound(c, response.BuildResponseCode(http.StatusNotFound, response.ServiceCodeRoles, response.CaseCodeNotFound), "not found", "role not found")
		default:
			h.internalError(c, response.ServiceCodeRoles, err, "delete failed")
		}
		return
	}
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeRoles, response.CaseCodeDeleted), "deleted", gin.H{"id": uint(id)})
}

