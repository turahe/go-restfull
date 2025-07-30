package services

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/turahe/go-restfull/config"
	"github.com/turahe/go-restfull/internal/logger"

	"go.uber.org/zap"
)

type BackupService struct {
	backupDir string
}

type BackupResult struct {
	Success   bool      `json:"success"`
	FilePath  string    `json:"file_path"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
	Error     string    `json:"error,omitempty"`
	Duration  float64   `json:"duration_seconds"`
}

func NewBackupService(backupDir string) *BackupService {
	// Create backup directory if it doesn't exist
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		if logger.Log != nil {
			logger.Log.Error("Failed to create backup directory", zap.Error(err))
		}
	}

	return &BackupService{
		backupDir: backupDir,
	}
}

// CreateBackup creates a database backup using pg_dump
func (bs *BackupService) CreateBackup(ctx context.Context) *BackupResult {
	startTime := time.Now()
	cfg := config.GetConfig()

	// If config is nil, return an error result
	if cfg == nil {
		return &BackupResult{
			Success:   false,
			FilePath:  "",
			CreatedAt: time.Now(),
			Error:     "Configuration not available",
			Duration:  time.Since(startTime).Seconds(),
		}
	}

	// Generate backup filename with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("backup_%s_%s.sql", cfg.Postgres.Database, timestamp)
	filePath := filepath.Join(bs.backupDir, filename)

	// Build pg_dump command
	cmd := exec.CommandContext(ctx, "pg_dump",
		"-h", cfg.Postgres.Host,
		"-p", fmt.Sprintf("%d", cfg.Postgres.Port),
		"-U", cfg.Postgres.Username,
		"-d", cfg.Postgres.Database,
		"-f", filePath,
		"--verbose",
		"--no-password",
	)

	// Set environment variables
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", cfg.Postgres.Password))

	// Execute the command
	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime).Seconds()

	if err != nil {
		if logger.Log != nil {
			logger.Log.Error("Database backup failed",
				zap.Error(err),
				zap.String("output", string(output)),
				zap.String("file_path", filePath),
			)
		}

		return &BackupResult{
			Success:   false,
			FilePath:  filePath,
			CreatedAt: time.Now(),
			Error:     fmt.Sprintf("Backup failed: %v. Output: %s", err, string(output)),
			Duration:  duration,
		}
	}

	// Get file size
	fileInfo, statErr := os.Stat(filePath)
	var size int64
	if statErr == nil {
		size = fileInfo.Size()
	}

	if logger.Log != nil {
		logger.Log.Info("Database backup completed successfully",
			zap.String("file_path", filePath),
			zap.Int64("size_bytes", size),
			zap.Float64("duration_seconds", duration),
		)
	}

	return &BackupResult{
		Success:   true,
		FilePath:  filePath,
		Size:      size,
		CreatedAt: time.Now(),
		Duration:  duration,
	}
}

// CleanupOldBackups removes backup files older than the specified retention period
func (bs *BackupService) CleanupOldBackups(retentionDays int) error {
	if retentionDays <= 0 {
		return nil
	}

	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)

	files, err := os.ReadDir(bs.backupDir)
	if err != nil {
		return fmt.Errorf("failed to read backup directory: %w", err)
	}

	var deletedCount int
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Check if file is a backup file
		if !isBackupFile(file.Name()) {
			continue
		}

		filePath := filepath.Join(bs.backupDir, file.Name())
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			if logger.Log != nil {
				logger.Log.Warn("Failed to get file info", zap.String("file", file.Name()), zap.Error(err))
			}
			continue
		}

		if fileInfo.ModTime().Before(cutoffTime) {
			if err := os.Remove(filePath); err != nil {
				if logger.Log != nil {
					logger.Log.Error("Failed to delete old backup file",
						zap.String("file", file.Name()),
						zap.Error(err),
					)
				}
			} else {
				deletedCount++
				if logger.Log != nil {
					logger.Log.Info("Deleted old backup file", zap.String("file", file.Name()))
				}
			}
		}
	}

	if deletedCount > 0 && logger.Log != nil {
		logger.Log.Info("Cleanup completed", zap.Int("deleted_files", deletedCount))
	}

	return nil
}

// isBackupFile checks if a file is a backup file
func isBackupFile(filename string) bool {
	return len(filename) > 4 && filename[len(filename)-4:] == ".sql"
}

// GetBackupStats returns statistics about existing backups
func (bs *BackupService) GetBackupStats() (map[string]interface{}, error) {
	files, err := os.ReadDir(bs.backupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var totalSize int64
	var backupCount int
	var oldestBackup, newestBackup time.Time

	for _, file := range files {
		if file.IsDir() || !isBackupFile(file.Name()) {
			continue
		}

		filePath := filepath.Join(bs.backupDir, file.Name())
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			continue
		}

		backupCount++
		totalSize += fileInfo.Size()

		if oldestBackup.IsZero() || fileInfo.ModTime().Before(oldestBackup) {
			oldestBackup = fileInfo.ModTime()
		}

		if newestBackup.IsZero() || fileInfo.ModTime().After(newestBackup) {
			newestBackup = fileInfo.ModTime()
		}
	}

	return map[string]interface{}{
		"total_backups":    backupCount,
		"total_size_bytes": totalSize,
		"oldest_backup":    oldestBackup,
		"newest_backup":    newestBackup,
		"backup_directory": bs.backupDir,
	}, nil
}
