package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRateLimiter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("disabled when rps and burst zero", func(t *testing.T) {
		r := gin.New()
		r.Use(RateLimiter(0, 0))
		r.GET("/", func(c *gin.Context) { c.String(200, "ok") })
		for i := 0; i < 5; i++ {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = "10.0.0.1:1"
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, 200, w.Code)
		}
	})

	t.Run("enforces burst limit", func(t *testing.T) {
		r := gin.New()
		r.Use(RateLimiter(10, 2))
		r.GET("/", func(c *gin.Context) { c.String(200, "ok") })
		req := func() *http.Request {
			rq := httptest.NewRequest(http.MethodGet, "/", nil)
			rq.RemoteAddr = "10.0.0.2:1"
			return rq
		}
		// burst=2: first 2 OK, 3rd and 4th 429
		for i, want := range []int{200, 200, 429, 429} {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req())
			assert.Equal(t, want, w.Code, "request %d", i+1)
		}
	})

	t.Run("different IPs get separate limits", func(t *testing.T) {
		r := gin.New()
		r.Use(RateLimiter(10, 1))
		r.GET("/", func(c *gin.Context) { c.String(200, "ok") })
		// IP1: 1st OK, 2nd 429
		req1 := httptest.NewRequest(http.MethodGet, "/", nil)
		req1.RemoteAddr = "10.0.0.10:1"
		w1 := httptest.NewRecorder()
		r.ServeHTTP(w1, req1)
		assert.Equal(t, 200, w1.Code)
		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, req1)
		assert.Equal(t, 429, w2.Code)
		// IP2: still gets OK
		req2 := httptest.NewRequest(http.MethodGet, "/", nil)
		req2.RemoteAddr = "10.0.0.11:1"
		w3 := httptest.NewRecorder()
		r.ServeHTTP(w3, req2)
		assert.Equal(t, 200, w3.Code)
	})
}
