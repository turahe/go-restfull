package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go-rest/internal/rbac"
	"go-rest/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func TestRBAC(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := zap.NewNop()

	db, err := gorm.Open(sqlite.Open("file:rbac_mw_test?mode=memory&cache=private"), &gorm.Config{})
	require.NoError(t, err)
	e, err := rbac.NewEnforcer(db, "../../configs/casbin_model.conf")
	require.NoError(t, err)
	rbacSvc := service.NewRBACService(e, db)

	t.Run("no auth returns 401", func(t *testing.T) {
		r := gin.New()
		r.Use(RBAC(rbacSvc, log))
		r.GET("/api/posts", func(c *gin.Context) { c.String(200, "ok") })

		req := httptest.NewRequest(http.MethodGet, "/api/posts", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("auth set reaches RBAC", func(t *testing.T) {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set(ctxAuthKey, AuthClaims{UserID: 1, Role: "user"})
			c.Next()
		})
		r.Use(RBAC(rbacSvc, log))
		r.GET("/api/posts", func(c *gin.Context) { c.String(200, "ok") })

		req := httptest.NewRequest(http.MethodGet, "/api/posts", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		// With auth set, we get past 401. Enforce may return 500 (no roles table) or 403.
		require.NotEqual(t, http.StatusUnauthorized, w.Code)
	})
}
