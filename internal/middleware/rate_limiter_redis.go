package middleware

import (
	"context"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/turahe/go-restfull/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var tokenBucketLua = redis.NewScript(`
-- KEYS[1] = key
-- ARGV[1] = capacity (burst) as number
-- ARGV[2] = refill_rate (tokens per second) as number
-- ARGV[3] = ttl_seconds as number
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local rate = tonumber(ARGV[2])
local ttl = tonumber(ARGV[3])

local t = redis.call("TIME")
local now = tonumber(t[1]) + (tonumber(t[2]) / 1000000.0)

local data = redis.call("HMGET", key, "tokens", "ts")
local tokens = tonumber(data[1])
local ts = tonumber(data[2])

if tokens == nil then
  tokens = capacity
  ts = now
end

local delta = now - ts
if delta < 0 then delta = 0 end

local refill = delta * rate
tokens = math.min(capacity, tokens + refill)

local allowed = 0
if tokens >= 1 then
  allowed = 1
  tokens = tokens - 1
end

redis.call("HMSET", key, "tokens", tokens, "ts", now)
redis.call("EXPIRE", key, ttl)

return { allowed, tokens }
`)

// RateLimiterRedis is a Redis-backed per-IP token bucket limiter.
// Set rps=0 or burst=0 to disable.
func RateLimiterRedis(rdb *redis.Client, keyPrefix string, rps float64, burst int, log *zap.Logger) gin.HandlerFunc {
	if rdb == nil || rps <= 0 || burst <= 0 {
		return func(c *gin.Context) { c.Next() }
	}
	if keyPrefix == "" {
		keyPrefix = "rl:ip:"
	}

	ttlSeconds := int(math.Ceil((float64(burst) / rps) * 2))
	if ttlSeconds < 60 {
		ttlSeconds = 60
	}
	if ttlSeconds > 3600 {
		ttlSeconds = 3600
	}

	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := keyPrefix + ip

		ctx, cancel := context.WithTimeout(c.Request.Context(), 50*time.Millisecond)
		defer cancel()

		res, err := tokenBucketLua.Run(ctx, rdb, []string{key}, burst, rps, ttlSeconds).Result()
		if err != nil {
			// Fail-open to avoid taking the API down if Redis is unavailable.
			if log != nil {
				log.Warn("redis rate limiter failed (fail-open)", zap.Error(err))
			}
			c.Next()
			return
		}

		arr, ok := res.([]any)
		if !ok || len(arr) < 1 {
			c.Next()
			return
		}

		allowed := toInt64(arr[0]) == 1
		if !allowed {
			// Optional: expose remaining as header if present.
			if len(arr) >= 2 {
				c.Header("X-RateLimit-Remaining", strconv.FormatInt(int64(toFloat64(arr[1])), 10))
			}
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

func toInt64(v any) int64 {
	switch t := v.(type) {
	case int64:
		return t
	case int:
		return int64(t)
	case float64:
		return int64(t)
	case string:
		n, _ := strconv.ParseInt(t, 10, 64)
		return n
	default:
		return 0
	}
}

func toFloat64(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case int64:
		return float64(t)
	case int:
		return float64(t)
	case string:
		n, _ := strconv.ParseFloat(t, 64)
		return n
	default:
		return 0
	}
}

