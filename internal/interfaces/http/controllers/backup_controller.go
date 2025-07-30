package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/turahe/go-restfull/config"
	"github.com/turahe/go-restfull/internal/domain/services"
	"github.com/turahe/go-restfull/internal/interfaces/http/responses"
	"github.com/turahe/go-restfull/internal/logger"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type BackupController struct {
	backupService *services.BackupService
}

func NewBackupController() *BackupController {
	cfg := config.GetConfig()
	backupDir := "backups" // Default directory
	if cfg.Backup.Directory != "" {
		backupDir = cfg.Backup.Directory
	}

	return &BackupController{
		backupService: services.NewBackupService(backupDir),
	}
}

// CreateBackup handles manual backup creation
func (bc *BackupController) CreateBackup(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	logger.Log.Info("Manual database backup requested")

	result := bc.backupService.CreateBackup(ctx)

	if !result.Success {
		logger.Log.Error("Manual backup failed",
			zap.String("error", result.Error),
			zap.Float64("duration", result.Duration),
		)

		return c.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Backup failed: " + result.Error,
		})
	}

	logger.Log.Info("Manual backup completed successfully",
		zap.String("file_path", result.FilePath),
		zap.Int64("size_bytes", result.Size),
		zap.Float64("duration_seconds", result.Duration),
	)

	return c.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status: "success",
		Data: map[string]interface{}{
			"file_path":        result.FilePath,
			"size_bytes":       result.Size,
			"duration_seconds": result.Duration,
			"created_at":       result.CreatedAt,
		},
	})
}

// GetBackupStats returns statistics about existing backups
func (bc *BackupController) GetBackupStats(c *fiber.Ctx) error {
	stats, err := bc.backupService.GetBackupStats()
	if err != nil {
		logger.Log.Error("Failed to get backup stats", zap.Error(err))

		return c.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Failed to get backup statistics: " + err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status: "success",
		Data:   stats,
	})
}

// CleanupOldBackups manually triggers cleanup of old backup files
func (bc *BackupController) CleanupOldBackups(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	retentionDays := cfg.Backup.RetentionDays
	if retentionDays <= 0 {
		retentionDays = 30 // Default to 30 days
	}

	logger.Log.Info("Manual cleanup of old backups requested", zap.Int("retention_days", retentionDays))

	if err := bc.backupService.CleanupOldBackups(retentionDays); err != nil {
		logger.Log.Error("Failed to cleanup old backups", zap.Error(err))

		return c.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Failed to cleanup old backups: " + err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status:  "success",
		Message: "Old backups cleaned up successfully",
	})
}

// CreateBackupWithPayload handles backup creation with custom parameters
func (bc *BackupController) CreateBackupWithPayload(c *fiber.Ctx) error {
	var payload struct {
		BackupDir     string `json:"backup_dir"`
		RetentionDays int    `json:"retention_days"`
		CleanupOld    bool   `json:"cleanup_old"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Invalid request body: " + err.Error(),
		})
	}

	// Use payload backup directory if provided, otherwise use default
	backupDir := "backups"
	if payload.BackupDir != "" {
		backupDir = payload.BackupDir
	}

	// Create temporary backup service with custom directory
	tempBackupService := services.NewBackupService(backupDir)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	logger.Log.Info("Custom backup requested",
		zap.String("backup_dir", backupDir),
		zap.Int("retention_days", payload.RetentionDays),
		zap.Bool("cleanup_old", payload.CleanupOld),
	)

	result := tempBackupService.CreateBackup(ctx)

	if !result.Success {
		logger.Log.Error("Custom backup failed",
			zap.String("error", result.Error),
			zap.Float64("duration", result.Duration),
		)

		return c.Status(http.StatusInternalServerError).JSON(responses.ErrorResponse{
			Status:  "error",
			Message: "Backup failed: " + result.Error,
		})
	}

	// Cleanup old backups if requested
	if payload.CleanupOld && payload.RetentionDays > 0 {
		if err := tempBackupService.CleanupOldBackups(payload.RetentionDays); err != nil {
			logger.Log.Error("Failed to cleanup old backups", zap.Error(err))
			// Don't return error here as the main backup was successful
		}
	}

	logger.Log.Info("Custom backup completed successfully",
		zap.String("file_path", result.FilePath),
		zap.Int64("size_bytes", result.Size),
		zap.Float64("duration_seconds", result.Duration),
	)

	return c.Status(http.StatusOK).JSON(responses.SuccessResponse{
		Status: "success",
		Data: map[string]interface{}{
			"file_path":        result.FilePath,
			"size_bytes":       result.Size,
			"duration_seconds": result.Duration,
			"created_at":       result.CreatedAt,
		},
	})
}
