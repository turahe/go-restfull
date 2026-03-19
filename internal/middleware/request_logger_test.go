package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestRequestLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)
	core, obs := observer.New(zap.InfoLevel)
	log := zap.New(core)
	r := gin.New()
	r.Use(RequestLogger(log))
	r.GET("/path", func(c *gin.Context) { c.String(200, "ok") })

	req := httptest.NewRequest(http.MethodGet, "/path", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	logs := obs.All()
	requireLen := 1
	assert.GreaterOrEqual(t, len(logs), requireLen, "expected at least one log entry")
	var found bool
	for _, e := range logs {
		if e.Message == "request" {
			found = true
			assert.Equal(t, "GET", e.ContextMap()["method"])
			assert.Equal(t, "/path", e.ContextMap()["path"])
			assert.Equal(t, int64(200), e.ContextMap()["status"])
			assert.Equal(t, "192.168.1.1", e.ContextMap()["client_ip"])
			break
		}
	}
	assert.True(t, found, "expected 'request' log entry")
}
