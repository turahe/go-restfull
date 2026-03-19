package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestRecovery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := zap.NewNop()
	r := gin.New()
	r.Use(Recovery(log))
	r.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})
	r.GET("/ok", func(c *gin.Context) { c.String(200, "ok") })

	t.Run("panic is recovered and returns 500", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/panic", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("no panic passes through", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/ok", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "ok", w.Body.String())
	})
}
