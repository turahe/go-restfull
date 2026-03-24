package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/turahe/go-restfull/pkg/ids"
)

const (
	RequestIDHeader = "X-Request-Id"
	ctxRequestIDKey = "request_id"
)

// RequestID ensures every request has a request-id for tracing.
// It accepts an incoming X-Request-Id or generates a new one.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := strings.TrimSpace(c.GetHeader(RequestIDHeader))
		if rid == "" {
			id, err := ids.New()
			if err == nil {
				rid = id
			}
		}
		if rid != "" {
			c.Set(ctxRequestIDKey, rid)
			// Make request-id available to any logger using Go's request context
			// (e.g. GORM logger's Trace(ctx, ...)).
			c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), ctxRequestIDKey, rid))
			c.Header(RequestIDHeader, rid)
		}
		c.Next()
	}
}

func GetRequestID(c *gin.Context) (string, bool) {
	v, ok := c.Get(ctxRequestIDKey)
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok && s != ""
}

