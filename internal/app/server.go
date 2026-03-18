package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "go-rest/docs"
	"go-rest/internal/config"
	"go-rest/internal/database"
	"go-rest/internal/handler"
	"go-rest/internal/middleware"
	"go-rest/internal/repository"
	"go-rest/internal/rbac"
	"go-rest/internal/service"
	"go-rest/pkg/logger"
	"go-rest/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

func Serve(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log, err := logger.New(cfg.Env)
	if err != nil {
		return err
	}
	defer func() { _ = log.Sync() }()

	db, err := database.ConnectMySQL(cfg)
	if err != nil {
		log.Fatal("db connect failed", zap.Error(err))
	}
	defer func() { _ = db.SQL.Close() }()

	var rdbCloser func()
	var rdb *redis.Client
	if cfg.RedisAddr != "" {
		rr, err := database.ConnectRedis(cfg)
		if err != nil {
			log.Warn("redis connect failed (rate limiter will be disabled)", zap.Error(err))
		} else {
			rdb = rr
			rdbCloser = func() { _ = rdb.Close() }
		}
	}
	if rdbCloser != nil {
		defer rdbCloser()
	}

	if err := database.AutoMigrate(db.Gorm); err != nil {
		log.Fatal("db migrate failed", zap.Error(err))
	}
	log.Info("db migrated")

	jwtm, err := service.NewJWTService(cfg.JWTPrivateKeyPath, cfg.JWTPublicKeyPath, cfg.JWTIssuer, cfg.JWTAudience, cfg.JWTKeyID)
	if err != nil {
		log.Fatal("jwt keys load failed", zap.Error(err))
	}

	enf, err := rbac.NewEnforcer(db.Gorm, cfg.CasbinModelPath)
	if err != nil {
		log.Fatal("casbin init failed", zap.Error(err))
	}
	rbacSvc := service.NewRBACService(enf, db.Gorm)

	// Repositories
	userRepo := repository.NewUserRepository(db.Gorm)
	authRepo := repository.NewAuthRepository(db.Gorm)
	auditRepo := repository.NewAuditRepository(db.Gorm)
	categoryRepo := repository.NewCategoryRepository(db.Gorm)
	tagRepo := repository.NewTagRepository(db.Gorm)
	postRepo := repository.NewPostRepository(db.Gorm)
	commentRepo := repository.NewCommentRepository(db.Gorm)

	// Services
	authSvc := service.NewAuthService(userRepo, authRepo, auditRepo, rbacSvc, jwtm, cfg.AccessTokenTTLMinutes, cfg.RefreshTokenTTLDays, cfg.ImpersonationTTLMinutes, cfg.RefreshTokenPepper)
	categorySvc := service.NewCategoryService(categoryRepo)
	postSvc := service.NewPostService(postRepo, categoryRepo, tagRepo)
	commentSvc := service.NewCommentService(commentRepo, tagRepo)

	// Handlers
	authH := handler.NewAuthHandler(authSvc, log)
	categoryH := handler.NewCategoryHandler(categorySvc, log)
	postH := handler.NewPostHandler(postSvc, log)
	commentH := handler.NewCommentHandler(commentSvc, log)
	rbacH := handler.NewRBACHandler(rbacSvc, log)

	if cfg.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middleware.Recovery(log))
	r.Use(middleware.RequestLogger(log))
	if rdb != nil {
		r.Use(middleware.RateLimiterRedis(rdb, cfg.RateLimitKeyPrefix, cfg.RateLimitRPS, cfg.RateLimitBurst, log))
	} else {
		r.Use(middleware.RateLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst))
	}

	r.GET("/swagger/*any", ginswagger.WrapHandler(swaggerfiles.Handler))
	r.GET("/healthz", func(c *gin.Context) {
		response.OK(c, 2000001, "ok", gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")
	{
		api.POST("auth/register", authH.Register)
		api.POST("auth/login", authH.Login)
		api.POST("auth/refresh", authH.Refresh)

		api.GET("/posts", postH.List)
		// NOTE: Gin can't disambiguate /posts/:slug from /posts/:id/comments.
		api.GET("/posts/slug/:slug", postH.GetBySlug)
		api.GET("/categories", categoryH.List)
		api.GET("/categories/:slug", categoryH.GetBySlug)

		auth := api.Group("")
		auth.Use(middleware.JWTAuth(jwtm, authRepo, log))
		auth.Use(middleware.RBAC(rbacSvc, log))
		{
			auth.POST("/auth/impersonate", authH.Impersonate)
			auth.POST("/posts", postH.Create)
			auth.PUT("/posts/:id", postH.Update)
			auth.DELETE("/posts/:id", postH.Delete)

			auth.POST("/posts/:id/comments", commentH.Create)

			auth.POST("/categories", categoryH.Create)
			auth.PUT("/categories/:id", categoryH.Update)
			auth.DELETE("/categories/:id", categoryH.Delete)

			auth.POST("/rbac/assign-role", rbacH.AssignRole)
			auth.POST("/rbac/add-permission", rbacH.AddPermission)
		}

		api.GET("/posts/:id/comments", commentH.List)
	}

	srv := &http.Server{
		Addr:              ":" + cfg.ServerPort,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("server starting", zap.String("port", cfg.ServerPort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
	case <-quit:
	case err := <-errCh:
		if err != nil {
			log.Error("server failed", zap.Error(err))
			return err
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	log.Info("server stopped")
	return nil
}

