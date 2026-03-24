package http

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/turahe/go-restfull/docs"
	"github.com/turahe/go-restfull/internal/config"
	"github.com/turahe/go-restfull/internal/database"
	"github.com/turahe/go-restfull/internal/handler"
	"github.com/turahe/go-restfull/internal/rbac"
	"github.com/turahe/go-restfull/internal/repository"
	"github.com/turahe/go-restfull/internal/service"
	"github.com/turahe/go-restfull/pkg/logger"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Serve is the HTTP entrypoint for the application.
// It owns process-level wiring: config, infra connections, DI, router, and graceful shutdown.
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

	db, err := database.ConnectMySQL(cfg, log)
	if err != nil {
		log.Fatal("db connect failed", zap.Error(err))
	}
	defer func() { _ = db.SQL.Close() }()

	var rdb *redis.Client
	if cfg.RedisAddr != "" {
		rr, err := database.ConnectRedis(cfg)
		if err != nil {
			log.Warn("redis connect failed (rate limiter will be disabled)", zap.Error(err))
		} else {
			rdb = rr
			defer func() { _ = rdb.Close() }()
		}
	}

	if err := database.AutoMigrate(db.Gorm); err != nil {
		log.Fatal("db migrate failed", zap.Error(err))
	}
	log.Info("db migrated")

	jwtm, err := service.NewJWTService(cfg.JWTPrivateKey, cfg.JWTPublicKey, cfg.JWTIssuer, cfg.JWTAudience, cfg.JWTKeyID, log)
	if err != nil {
		log.Fatal("jwt keys load failed", zap.Error(err))
	}

	enf, err := rbac.NewEnforcer(db.Gorm, cfg.CasbinModelPath)
	if err != nil {
		log.Fatal("casbin init failed", zap.Error(err))
	}
	rbacSvc := service.NewRBACService(enf, db.Gorm, log)

	// Repositories
	userRepo := repository.NewUserRepository(db.Gorm, log)
	authRepo := repository.NewAuthRepository(db.Gorm, log)
	auditRepo := repository.NewAuditRepository(db.Gorm, log)
	categoryRepo := repository.NewCategoryRepository(db.Gorm, log)
	tagRepo := repository.NewTagRepository(db.Gorm, log)
	roleRepo := repository.NewRoleRepository(db.Gorm, log)
	postRepo := repository.NewPostRepository(db.Gorm, log)
	commentRepo := repository.NewCommentRepository(db.Gorm, log)
	twoFARepo := repository.NewTwoFactorRepository(db.Gorm, log)
	mediaRepo := repository.NewMediaRepository(db.Gorm, log)
	settingRepo := repository.NewSettingRepository(db.Gorm, log)

	// Services
	twoFASvc := service.NewTwoFactorService(twoFARepo, []byte(cfg.TwoFactorEncKey), cfg.TwoFactorIssuer, log)
	mediaSvc, err := service.NewMediaService(mediaRepo, cfg, log)
	if err != nil {
		return err
	}
	authSvc := service.NewAuthService(userRepo,
		authRepo,
		auditRepo,
		rbacSvc,
		jwtm,
		twoFASvc,
		mediaSvc,
		cfg.AccessTokenTTLMinutes,
		cfg.RefreshTokenTTLDays,
		cfg.ImpersonationTTLMinutes,
		cfg.RefreshTokenPepper,
		log,
	)
	userSvc := service.NewUserService(userRepo, mediaSvc, log)
	roleSvc := service.NewRoleService(roleRepo, log)
	categorySvc := service.NewCategoryService(categoryRepo, log)
	tagSvc := service.NewTagService(tagRepo, log)
	postSvc := service.NewPostService(postRepo, categoryRepo, tagRepo, log)
	commentSvc := service.NewCommentService(commentRepo, tagRepo, log)
	settingsSvc := service.NewSettingsService(settingRepo)

	// Handlers
	authH := handler.NewAuthHandler(authSvc, log)
	userH := handler.NewUserHandler(userSvc, log)
	roleH := handler.NewRoleHandler(roleSvc, log)
	categoryH := handler.NewCategoryHandler(categorySvc, log)
	tagH := handler.NewTagHandler(tagSvc, log)
	postH := handler.NewPostHandler(postSvc, log)
	commentH := handler.NewCommentHandler(commentSvc, log)
	mediaH := handler.NewMediaHandler(mediaSvc, log)
	rbacH := handler.NewRBACHandler(rbacSvc, log)
	settingsH := handler.NewSettingsHandler(settingsSvc, log)

	r := NewRouter(Deps{
		Cfg:      cfg,
		Log:      log,
		Redis:    rdb,
		JWT:      jwtm,
		RBAC:     rbacSvc,
		AuthRepo: authRepo,
		Handlers: Handlers{
			Auth:     authH,
			User:     userH,
			Role:     roleH,
			Category: categoryH,
			Tag:      tagH,
			Post:     postH,
			Comment:  commentH,
			Media:    mediaH,
			RBAC:     rbacH,
			Settings: settingsH,
		},
	})

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
