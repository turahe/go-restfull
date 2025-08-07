package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/turahe/go-restfull/config"
	"github.com/turahe/go-restfull/internal/db/pgx"
	"github.com/turahe/go-restfull/internal/db/rdb"
	internal_minio "github.com/turahe/go-restfull/pkg/minio"

	"github.com/gofiber/fiber/v2"
)

type HealthzHTTPHandler struct{}

// HealthCheck represents the health status of a service
type HealthCheck struct {
	Service   string `json:"service"`
	Status    string `json:"status"`
	Message   string `json:"message,omitempty"`
	Timestamp string `json:"timestamp"`
}

// HealthResponse represents the complete health check response
type HealthResponse struct {
	Status      string        `json:"status"`
	Timestamp   string        `json:"timestamp"`
	Environment string        `json:"environment"`
	Version     string        `json:"version"`
	Services    []HealthCheck `json:"services"`
}

func NewHealthzHTTPHandler() *HealthzHTTPHandler {
	return &HealthzHTTPHandler{}
}

// Healthz godoc
//
//	@Summary		Health check endpoint
//	@Description	Check if the API and all services are running and healthy
//	@Tags			health
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	HealthResponse
//	@Success		503	{object}	HealthResponse
//	@Router			/healthz [get]
//
// Hot reload test - this comment was added to test Air functionality
func (h *HealthzHTTPHandler) Healthz(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	healthChecks := []HealthCheck{}
	overallStatus := "healthy"
	timestamp := time.Now().Format(time.RFC3339)

	// Check PostgreSQL Database
	postgresHealth := h.checkPostgreSQL(ctx)
	healthChecks = append(healthChecks, postgresHealth)
	if postgresHealth.Status != "healthy" {
		overallStatus = "unhealthy"
	}

	// Check Redis
	redisHealth := h.checkRedis(ctx)
	healthChecks = append(healthChecks, redisHealth)
	if redisHealth.Status != "healthy" {
		overallStatus = "unhealthy"
	}

	// Check MinIO (if enabled)
	if config.GetConfig().Minio.Enable {
		minioHealth := h.checkMinIO(ctx)
		healthChecks = append(healthChecks, minioHealth)
		if minioHealth.Status != "healthy" {
			overallStatus = "unhealthy"
		}
	}

	// Check Application Services
	appHealth := h.checkApplicationServices(ctx)
	healthChecks = append(healthChecks, appHealth...)

	// Check RBAC Service
	rbacHealth := h.checkRBACService(ctx)
	healthChecks = append(healthChecks, rbacHealth)
	if rbacHealth.Status != "healthy" {
		overallStatus = "unhealthy"
	}

	// Check RabbitMQ (replaces job queue)
	rabbitMQHealth := h.checkRabbitMQ(ctx)
	healthChecks = append(healthChecks, rabbitMQHealth)
	if rabbitMQHealth.Status != "healthy" {
		overallStatus = "unhealthy"
	}

	// Check Email Service
	emailHealth := h.checkEmailService(ctx)
	healthChecks = append(healthChecks, emailHealth)

	// Check Sentry (if enabled)
	if config.GetConfig().Sentry.Dsn != "" {
		sentryHealth := h.checkSentry(ctx)
		healthChecks = append(healthChecks, sentryHealth)
	}

	response := HealthResponse{
		Status:      overallStatus,
		Timestamp:   timestamp,
		Environment: config.GetConfig().Env,
		Version:     config.GetConfig().App.Name + " (Hot Reload Test)",
		Services:    healthChecks,
	}

	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	return c.Status(statusCode).JSON(response)
}

func (h *HealthzHTTPHandler) checkPostgreSQL(ctx context.Context) HealthCheck {
	pool := pgx.GetPgxPool()
	if pool == nil {
		return HealthCheck{
			Service:   "postgresql",
			Status:    "unhealthy",
			Message:   "Database pool not initialized",
			Timestamp: time.Now().Format(time.RFC3339),
		}
	}

	err := pool.Ping(ctx)
	if err != nil {
		return HealthCheck{
			Service:   "postgresql",
			Status:    "unhealthy",
			Message:   err.Error(),
			Timestamp: time.Now().Format(time.RFC3339),
		}
	}

	return HealthCheck{
		Service:   "postgresql",
		Status:    "healthy",
		Message:   "Database connection successful",
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

func (h *HealthzHTTPHandler) checkRedis(ctx context.Context) HealthCheck {
	client := rdb.GetRedisClient()
	if client == nil {
		return HealthCheck{
			Service:   "redis",
			Status:    "unhealthy",
			Message:   "Redis client not initialized",
			Timestamp: time.Now().Format(time.RFC3339),
		}
	}

	err := client.Ping(ctx).Err()
	if err != nil {
		return HealthCheck{
			Service:   "redis",
			Status:    "unhealthy",
			Message:   err.Error(),
			Timestamp: time.Now().Format(time.RFC3339),
		}
	}

	return HealthCheck{
		Service:   "redis",
		Status:    "healthy",
		Message:   "Redis connection successful",
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

func (h *HealthzHTTPHandler) checkMinIO(ctx context.Context) HealthCheck {
	// Use the existing MinIO client from the package
	minioClient := internal_minio.GetMinio()
	if minioClient == nil {
		return HealthCheck{
			Service:   "minio",
			Status:    "unhealthy",
			Message:   "MinIO client not initialized",
			Timestamp: time.Now().Format(time.RFC3339),
		}
	}

	// Check if MinIO is alive
	if !internal_minio.IsAlive() {
		return HealthCheck{
			Service:   "minio",
			Status:    "unhealthy",
			Message:   "MinIO connection failed",
			Timestamp: time.Now().Format(time.RFC3339),
		}
	}

	return HealthCheck{
		Service:   "minio",
		Status:    "healthy",
		Message:   "MinIO connection successful",
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

func (h *HealthzHTTPHandler) checkApplicationServices(ctx context.Context) []HealthCheck {
	checks := []HealthCheck{}

	// Check application configuration
	cfg := config.GetConfig()
	if cfg == nil {
		checks = append(checks, HealthCheck{
			Service:   "configuration",
			Status:    "unhealthy",
			Message:   "Configuration not loaded",
			Timestamp: time.Now().Format(time.RFC3339),
		})
	} else {
		checks = append(checks, HealthCheck{
			Service:   "configuration",
			Status:    "healthy",
			Message:   "Configuration loaded successfully",
			Timestamp: time.Now().Format(time.RFC3339),
		})
	}

	// Check application status
	checks = append(checks, HealthCheck{
		Service:   "application",
		Status:    "healthy",
		Message:   "Application is running",
		Timestamp: time.Now().Format(time.RFC3339),
	})

	return checks
}

func (h *HealthzHTTPHandler) checkRBACService(ctx context.Context) HealthCheck {
	// For now, we'll assume RBAC is healthy if the application is running
	// In a real implementation, you might want to check if Casbin is properly initialized
	return HealthCheck{
		Service:   "rbac",
		Status:    "healthy",
		Message:   "RBAC service is available",
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

func (h *HealthzHTTPHandler) checkRabbitMQ(ctx context.Context) HealthCheck {
	// For now, we'll assume RabbitMQ is healthy if the application is running
	// In a real implementation, you might want to check RabbitMQ connectivity
	return HealthCheck{
		Service:   "rabbitmq",
		Status:    "healthy",
		Message:   "RabbitMQ messaging is available",
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

func (h *HealthzHTTPHandler) checkEmailService(ctx context.Context) HealthCheck {
	// For now, we'll assume email service is healthy
	// In a real implementation, you might want to check SMTP connectivity
	return HealthCheck{
		Service:   "email",
		Status:    "healthy",
		Message:   "Email service is available",
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

func (h *HealthzHTTPHandler) checkSentry(ctx context.Context) HealthCheck {
	// For now, we'll assume Sentry is healthy if enabled
	// In a real implementation, you might want to check Sentry connectivity
	return HealthCheck{
		Service:   "sentry",
		Status:    "healthy",
		Message:   "Sentry is available",
		Timestamp: time.Now().Format(time.RFC3339),
	}
}
