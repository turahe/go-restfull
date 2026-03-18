package main

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
	"go-rest/internal/service"
	"go-rest/pkg/logger"
	"go-rest/pkg/response"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// @title           Blog REST API
// @version         1.0
// @description     Blog API with JWT auth, posts, and comments.
// @BasePath        /
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	log, err := logger.New(cfg.Env)
	if err != nil {
		panic(err)
	}
	defer func() { _ = log.Sync() }()

	db, err := database.ConnectMySQL(cfg)
	if err != nil {
		log.Fatal("db connect failed", zap.Error(err))
	}
	defer func() { _ = db.SQL.Close() }()

	if err := database.AutoMigrate(db.Gorm); err != nil {
		log.Fatal("db migrate failed", zap.Error(err))
	}
	log.Info("db migrated")

	jwtm, err := service.NewJWTManager(cfg.JWTPrivateKeyPath, cfg.JWTPublicKeyPath, cfg.JWTIssuer, cfg.JWTTTLMinutes)
	if err != nil {
		log.Fatal("jwt keys load failed", zap.Error(err))
	}

	// Repositories
	userRepo := repository.NewUserRepository(db.Gorm)
	postRepo := repository.NewPostRepository(db.Gorm)
	commentRepo := repository.NewCommentRepository(db.Gorm)

	// Services
	authSvc := service.NewAuthService(userRepo, jwtm)
	postSvc := service.NewPostService(postRepo)
	commentSvc := service.NewCommentService(commentRepo)

	// Handlers
	authH := handler.NewAuthHandler(authSvc, log)
	postH := handler.NewPostHandler(postSvc, log)
	commentH := handler.NewCommentHandler(commentSvc, log)

	if cfg.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middleware.Recovery(log))
	r.Use(middleware.RequestLogger(log))

	r.GET("/swagger/*any", ginswagger.WrapHandler(swaggerfiles.Handler))

	r.GET("/healthz", func(c *gin.Context) {
		response.OK(c, 2000001, "ok", gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")
	{
		api.POST("auth/register", authH.Register)
		api.POST("auth/login", authH.Login)

		api.GET("/posts", postH.List)
		// NOTE: Gin can't disambiguate /posts/:slug from /posts/:id/comments.
		api.GET("/posts/slug/:slug", postH.GetBySlug)

		auth := api.Group("")
		auth.Use(middleware.JWTAuth(jwtm.PublicKey(), cfg.JWTIssuer, log))
		{
			auth.POST("/posts", postH.Create)
			auth.PUT("/posts/:id", postH.Update)
			auth.DELETE("/posts/:id", postH.Delete)

			auth.POST("/posts/:id/comments", commentH.Create)
		}

		api.GET("/posts/:id/comments", commentH.List)
	}

	srv := &http.Server{
		Addr:              ":" + cfg.ServerPort,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Info("server starting", zap.String("port", cfg.ServerPort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	log.Info("server stopped")
}
