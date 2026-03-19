package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.GET("/", func(c *gin.Context) {
		id, ok := GetRequestID(c)
		require.True(t, ok)
		assert.NotEmpty(t, id)
		c.String(200, id)
	})

	t.Run("generates id when header missing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.NotEmpty(t, w.Header().Get(RequestIDHeader))
		assert.Equal(t, w.Header().Get(RequestIDHeader), w.Body.String())
	})

	t.Run("uses incoming X-Request-Id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(RequestIDHeader, "my-request-id-123")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "my-request-id-123", w.Header().Get(RequestIDHeader))
		assert.Equal(t, "my-request-id-123", w.Body.String())
	})
}

func TestGetRequestID(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	_, ok := GetRequestID(c)
	assert.False(t, ok)

	c.Set(ctxRequestIDKey, "abc")
	id, ok := GetRequestID(c)
	assert.True(t, ok)
	assert.Equal(t, "abc", id)

	c.Set(ctxRequestIDKey, "")
	_, ok = GetRequestID(c)
	assert.False(t, ok)
}
