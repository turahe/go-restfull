package middleware

import (
	"net/http"

	"go-rest/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Recovery(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Error("panic", zap.Any("recover", rec))
				response.JSON(c, http.StatusInternalServerError, 5000001, "internal server error", nil, "panic")
				c.Abort()
			}
		}()
		c.Next()
	}
}

