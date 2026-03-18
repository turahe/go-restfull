package middleware

import (
	"net/http"

	"go-rest/internal/service"
	"go-rest/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RBAC enforces Casbin permissions based on (sub=userID, obj=route template, act=http method).
func RBAC(rbacSvc *service.RBACService, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth, ok := GetAuth(c)
		if !ok {
			response.Unauthorized(c, response.BuildResponseCode(http.StatusUnauthorized, response.ServiceCodeAuth, response.CaseCodeUnauthorized), "unauthorized", "missing auth")
			c.Abort()
			return
		}

		obj := c.FullPath()
		if obj == "" {
			obj = c.Request.URL.Path
		}
		act := c.Request.Method

		allowed, err := rbacSvc.Enforce(c.Request.Context(), auth.UserID, obj, act)
		if err != nil {
			if log != nil {
				log.Error("rbac enforce failed", zap.Error(err))
			}
			response.InternalServerError(c, response.BuildResponseCode(http.StatusInternalServerError, response.ServiceCodeCommon, response.CaseCodeInternalError), "internal error", "authorization failed")
			c.Abort()
			return
		}
		if !allowed {
			response.Forbidden(c, response.BuildResponseCode(http.StatusForbidden, response.ServiceCodeAuth, response.CaseCodePermissionDenied), "forbidden", "insufficient permissions")
			c.Abort()
			return
		}
		c.Next()
	}
}

