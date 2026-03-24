package handler

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/turahe/go-restfull/internal/config"
	"github.com/turahe/go-restfull/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type HealthHandler struct {
	db  *sql.DB
	rdb *redis.Client
	cfg config.Config
}

func NewHealthHandler(db *sql.DB, rdb *redis.Client, cfg config.Config) *HealthHandler {
	return &HealthHandler{db: db, rdb: rdb, cfg: cfg}
}

func (h *HealthHandler) Health(c *gin.Context) {
	response.OK(c, 2000001, "ok", gin.H{"status": "ok"})
}

func (h *HealthHandler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	dbOK := h.db != nil && h.db.PingContext(ctx) == nil
	redisConfigured := strings.TrimSpace(h.cfg.RedisAddr) != ""
	redisOK := true
	redisStatus := "disabled"
	if redisConfigured {
		redisStatus = "ok"
		if h.rdb == nil {
			redisOK = false
			redisStatus = "unavailable"
		} else if err := h.rdb.Ping(ctx).Err(); err != nil {
			redisOK = false
			redisStatus = "unhealthy"
		}
	}

	configOK := h.cfg.JWTIssuer != "" &&
		h.cfg.JWTAudience != "" &&
		h.cfg.JWTKeyID != "" &&
		h.cfg.TwoFactorEncKey != "" &&
		(h.cfg.MediaStorage == "s3" || h.cfg.MediaStorage == "gcs")

	ready := dbOK && redisOK && configOK
	data := gin.H{
		"status":      "ready",
		"checks": gin.H{
			"database": dbOK,
			"redis":    redisStatus,
			"config":   configOK,
		},
		"config": gin.H{
			"dbDriver":     h.cfg.DBDriver,
			"mediaStorage": h.cfg.MediaStorage,
			"swagger":      h.cfg.SwaggerEnabled,
		},
	}
	if ready {
		response.OK(c, 2000002, "ready", data)
		return
	}

	data["status"] = "not_ready"
	response.JSON(c, http.StatusServiceUnavailable, 5000003, "not ready", data, "dependency/config check failed")
}
