package http

import (
	"go-rest/internal/config"
	"go-rest/internal/handler"
	"go-rest/internal/middleware"
	"go-rest/internal/repository"
	"go-rest/internal/service"
	"go-rest/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

type Deps struct {
	Cfg     config.Config
	Log     *zap.Logger
	Redis   *redis.Client

	JWT     *service.JWTService
	RBAC    *service.RBACService
	AuthRepo *repository.AuthRepository

	Handlers Handlers
}

type Handlers struct {
	Auth     *handler.AuthHandler
	User     *handler.UserHandler
	Role     *handler.RoleHandler
	Category *handler.CategoryHandler
	Tag      *handler.TagHandler
	Post     *handler.PostHandler
	Comment  *handler.CommentHandler
	Media    *handler.MediaHandler
	RBAC     *handler.RBACHandler
}

func NewRouter(d Deps) *gin.Engine {
	if d.Cfg.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middleware.Recovery(d.Log))
	r.Use(middleware.RequestID())
	r.Use(middleware.RequestLogger(d.Log))
	if d.Redis != nil {
		r.Use(middleware.RateLimiterRedis(d.Redis, d.Cfg.RateLimitKeyPrefix, d.Cfg.RateLimitRPS, d.Cfg.RateLimitBurst, d.Log))
	} else {
		r.Use(middleware.RateLimiter(d.Cfg.RateLimitRPS, d.Cfg.RateLimitBurst))
	}

	if d.Cfg.Env == "local" {
		r.GET("/swagger/*any", ginswagger.WrapHandler(swaggerfiles.Handler))
	}
	r.GET("/healthz", func(c *gin.Context) {
		response.OK(c, 2000001, "ok", gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")
	{
		api.POST("auth/register", d.Handlers.Auth.Register)
		api.POST("auth/login", d.Handlers.Auth.Login)
		api.POST("auth/refresh", d.Handlers.Auth.Refresh)

		api.GET("/posts", d.Handlers.Post.List)
		api.GET("/posts/slug/:slug", d.Handlers.Post.GetBySlug)
		api.GET("/categories", d.Handlers.Category.List)
		api.GET("/categories/:slug", d.Handlers.Category.GetBySlug)
		api.GET("/tags", d.Handlers.Tag.List)
		api.GET("/tags/:slug", d.Handlers.Tag.GetBySlug)

		auth := api.Group("")
		auth.Use(middleware.JWTAuth(d.JWT, d.AuthRepo, d.Log))
		auth.Use(middleware.RBAC(d.RBAC, d.Log))
		{
			auth.GET("/auth/profile", d.Handlers.Auth.Profile)
			auth.POST("/auth/password/change", d.Handlers.Auth.ChangePassword)
			auth.POST("/auth/email/change", d.Handlers.Auth.ChangeEmail)
			auth.POST("/auth/2fa/setup", d.Handlers.Auth.TwoFASetup)
			auth.POST("/auth/2fa/enable", d.Handlers.Auth.TwoFAEnable)
			auth.POST("/auth/impersonate", d.Handlers.Auth.Impersonate)

			auth.POST("/posts", d.Handlers.Post.Create)
			auth.PUT("/posts/:id", d.Handlers.Post.Update)
			auth.DELETE("/posts/:id", d.Handlers.Post.Delete)
			auth.POST("/posts/:id/comments", d.Handlers.Comment.Create)

			auth.POST("/categories", d.Handlers.Category.Create)
			auth.PUT("/categories/:id", d.Handlers.Category.Update)
			auth.DELETE("/categories/:id", d.Handlers.Category.Delete)

			auth.POST("/tags", d.Handlers.Tag.Create)
			auth.PUT("/tags/:id", d.Handlers.Tag.Update)
			auth.DELETE("/tags/:id", d.Handlers.Tag.Delete)

			auth.POST("/media", d.Handlers.Media.UploadMedia)
			auth.GET("/media", d.Handlers.Media.ListMedia)
			auth.GET("/media/:id", d.Handlers.Media.GetMediaByID)
			auth.DELETE("/media/:id", d.Handlers.Media.DeleteMedia)

			auth.GET("/users", d.Handlers.User.List)
			auth.GET("/users/:id", d.Handlers.User.GetByID)

			auth.GET("/roles", d.Handlers.Role.List)
			auth.POST("/roles", d.Handlers.Role.Create)
			auth.DELETE("/roles/:id", d.Handlers.Role.Delete)

			auth.POST("/rbac/assign-role", d.Handlers.RBAC.AssignRole)
			auth.POST("/rbac/add-permission", d.Handlers.RBAC.AddPermission)
		}

		api.GET("/posts/:id/comments", d.Handlers.Comment.List)
		api.POST("/auth/2fa/verify", d.Handlers.Auth.TwoFAVerify)
	}

	return r
}

