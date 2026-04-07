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

type RBACHandler struct {
	BaseHandler
	rbac *service.RBACService
}

func NewRBACHandler(rbacSvc *service.RBACService, log *zap.Logger) *RBACHandler {
	return &RBACHandler{BaseHandler: BaseHandler{Log: log}, rbac: rbacSvc}
}

// AssignRole godoc
// @Summary      Assign role to user
// @Tags         RBAC
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      request.AssignRoleRequest  true  "Assign role"
// @Success      200   {object}  response.Envelope
// @Failure      400   {object}  response.Envelope
// @Failure      401   {object}  response.Envelope
// @Failure      403   {object}  response.Envelope
// @Failure      500   {object}  response.Envelope
// @Router       /api/v1/rbac/assign-role [post]
func (h *RBACHandler) AssignRole(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok {
		response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeAuth, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}
	if auth.Role != "admin" {
		response.Forbidden(c, response.BuildResponseCode(http.StatusForbidden, response.ServiceCodeAuth, response.CaseCodePermissionDenied), "forbidden", "admin only")
		return
	}

	var req request.AssignRoleRequest
	if !h.bindJSON(c, response.ServiceCodeCommon, &req) {
		return
	}
	if !h.validate(c, response.ServiceCodeCommon, req) {
		return
	}

	ok2, err := h.rbac.AssignRole(c.Request.Context(), req.UserID, req.Role)
	if err != nil {
		h.internalError(c, response.ServiceCodeCommon, err, "assign role failed")
		return
	}
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeCommon, response.CaseCodeSuccess), "ok", gin.H{"assigned": ok2})
}

// AddPermission godoc
// @Summary      Add permission to role
// @Tags         RBAC
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      request.AddPermissionRequest  true  "Add permission"
// @Success      200   {object}  response.Envelope
// @Failure      400   {object}  response.Envelope
// @Failure      401   {object}  response.Envelope
// @Failure      403   {object}  response.Envelope
// @Failure      500   {object}  response.Envelope
// @Router       /api/v1/rbac/add-permission [post]
func (h *RBACHandler) AddPermission(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok {
		response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeAuth, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
		return
	}
	if auth.Role != "admin" {
		response.Forbidden(c, response.BuildResponseCode(http.StatusForbidden, response.ServiceCodeAuth, response.CaseCodePermissionDenied), "forbidden", "admin only")
		return
	}

	var req request.AddPermissionRequest
	if !h.bindJSON(c, response.ServiceCodeCommon, &req) {
		return
	}
	if !h.validate(c, response.ServiceCodeCommon, req) {
		return
	}

	ok2, err := h.rbac.AddPermissionToRole(c.Request.Context(), req.Role, req.Obj, req.Act)
	if err != nil {
		h.internalError(c, response.ServiceCodeCommon, err, "add permission failed")
		return
	}
	response.OK(c, response.BuildResponseCode(http.StatusOK, response.ServiceCodeCommon, response.CaseCodeSuccess), "ok", gin.H{"added": ok2})
}
