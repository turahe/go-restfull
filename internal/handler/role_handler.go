package handler

import (
	"net/http"

	"github.com/turahe/go-restfull/internal/handler/request"
	"github.com/turahe/go-restfull/internal/middleware"
	"github.com/turahe/go-restfull/internal/service"
	"github.com/turahe/go-restfull/pkg/response"

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

	var req request.RoleListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeRoles, response.CaseCodeInvalidValue), "invalid limit", err.Error())
		return
	}
	if !h.validate(c, response.ServiceCodeRoles, req) {
		return
	}

	page, err := h.roles.List(c.Request.Context(), req)
	if err != nil {
		h.internalError(c, response.ServiceCodeRoles, err, "list failed")
		return
	}

	next := page.NextCursor != nil
	prev := page.PrevCursor != nil
	response.OKPaginated(
		c,
		response.BuildResponseCode(http.StatusOK, response.ServiceCodeRoles, response.CaseCodeListRetrieved),
		"ok",
		page.Items,
		next,
		prev,
	)
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

	role, err := h.roles.Create(c.Request.Context(), req)
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

	id, err := h.ParseUintParam(c, "id")
	if err != nil {
		response.BadRequest(c, response.BuildResponseCode(http.StatusBadRequest, response.ServiceCodeRoles, response.CaseCodeInvalidValue), "invalid id", "id must be uint")
		return
	}

	err = h.roles.Delete(c.Request.Context(), id)
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
