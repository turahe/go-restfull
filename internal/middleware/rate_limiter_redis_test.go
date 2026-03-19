package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func TestRateLimiterRedis_Integration(t *testing.T) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		t.Skip("REDIS_ADDR not set, skipping Redis rate limiter integration test")
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	defer rdb.Close()
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Skipf("Redis not reachable: %v", err)
	}

	log := zap.NewNop()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RateLimiterRedis(rdb, "rl:test:", 2, 2, log))
	r.GET("/ping", func(c *gin.Context) { c.String(200, "ok") })

	// Burst=2, same IP: first 2 requests 200, 3rd and 4th should be 429
	req := func() *http.Request {
		rq := httptest.NewRequest(http.MethodGet, "/ping", nil)
		rq.RemoteAddr = "192.168.1.1:12345"
		return rq
	}
	for i, wantCode := range []int{http.StatusOK, http.StatusOK, http.StatusTooManyRequests, http.StatusTooManyRequests} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req())
		if w.Code != wantCode {
			t.Errorf("request %d: got status %d, want %d", i+1, w.Code, wantCode)
		}
	}
}
