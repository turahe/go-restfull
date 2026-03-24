package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/turahe/go-restfull/pkg/response"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter is an in-memory per-client-IP token bucket limiter.
// Set rps=0 and burst=0 to disable.
func RateLimiter(rps float64, burst int) gin.HandlerFunc {
	if rps <= 0 || burst <= 0 {
		return func(c *gin.Context) { c.Next() }
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*clientLimiter)
	)

	// Best-effort cleanup.
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			cutoff := time.Now().Add(-10 * time.Minute)
			mu.Lock()
			for k, v := range clients {
				if v.lastSeen.Before(cutoff) {
					delete(clients, k)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		mu.Lock()
		cl, ok := clients[ip]
		if !ok {
			cl = &clientLimiter{
				limiter:  rate.NewLimiter(rate.Limit(rps), burst),
				lastSeen: time.Now(),
			}
			clients[ip] = cl
		}
		cl.lastSeen = time.Now()
		lim := cl.limiter
		mu.Unlock()

		if !lim.Allow() {
			response.JSON(
				c,
				http.StatusTooManyRequests,
				response.BuildResponseCode(http.StatusTooManyRequests, response.ServiceCodeCommon, response.CaseCodeRateLimitExceeded),
				"too many requests",
				nil,
				"Rate limit exceeded.",
			)
			c.Abort()
			return
		}
		c.Next()
	}
}

