package job

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/turahe/go-restfull/pkg/logger"
	"time"

	"github.com/turahe/go-restfull/internal/domain/services"
	"go.uber.org/zap"
)

type BackupJobPayload struct {
	BackupDir     string `json:"backup_dir"`
	RetentionDays int    `json:"retention_days"`
	CleanupOld    bool   `json:"cleanup_old"`
}

type BackupJobHandler struct {
	backupService *services.BackupService
}

func NewBackupJobHandler(backupDir string) *BackupJobHandler {
	return &BackupJobHandler{
		backupService: services.NewBackupService(backupDir),
	}
}

func (h *BackupJobHandler) Handle() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	if logger.Log != nil {
		logger.Log.Info("Starting database backup job")
	}

	// Create backup
	result := h.backupService.CreateBackup(ctx)

	if !result.Success {
		return fmt.Errorf("backup failed: %s", result.Error)
	}

	if logger.Log != nil {
		logger.Log.Info("Database backup completed",
			zap.String("file_path", result.FilePath),
			zap.Int64("size_bytes", result.Size),
			zap.Float64("duration_seconds", result.Duration),
		)
	}

	return nil
}

// HandleWithPayload handles backup job with custom payload
func (h *BackupJobHandler) HandleWithPayload(payload json.RawMessage) error {
	var backupPayload BackupJobPayload
	if err := json.Unmarshal(payload, &backupPayload); err != nil {
		return fmt.Errorf("failed to unmarshal backup payload: %w", err)
	}

	// Use payload backup directory if provided, otherwise use default
	backupDir := "backups" // Default backup directory
	if backupPayload.BackupDir != "" {
		backupDir = backupPayload.BackupDir
	}

	// Create temporary backup service with custom directory
	tempBackupService := services.NewBackupService(backupDir)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	if logger.Log != nil {
		logger.Log.Info("Starting database backup job with custom payload",
			zap.String("backup_dir", backupDir),
			zap.Int("retention_days", backupPayload.RetentionDays),
			zap.Bool("cleanup_old", backupPayload.CleanupOld),
		)
	}

	// Create backup
	result := tempBackupService.CreateBackup(ctx)

	if !result.Success {
		return fmt.Errorf("backup failed: %s", result.Error)
	}

	if logger.Log != nil {
		logger.Log.Info("Database backup completed",
			zap.String("file_path", result.FilePath),
			zap.Int64("size_bytes", result.Size),
			zap.Float64("duration_seconds", result.Duration),
		)
	}

	// Cleanup old backups if requested
	if backupPayload.CleanupOld && backupPayload.RetentionDays > 0 {
		if err := tempBackupService.CleanupOldBackups(backupPayload.RetentionDays); err != nil {
			if logger.Log != nil {
				logger.Log.Error("Failed to cleanup old backups", zap.Error(err))
			}
			// Don't return error here as the main backup was successful
		}
	}

	return nil
}
