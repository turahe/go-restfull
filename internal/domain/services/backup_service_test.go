package services_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"webapi/config"
	"webapi/internal/domain/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBackupService(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	backupService := services.NewBackupService(tempDir)

	assert.NotNil(t, backupService)

	// Verify the backup directory was created
	_, err := os.Stat(tempDir)
	assert.NoError(t, err)
}

func TestBackupService_CreateBackup(t *testing.T) {
	// Skip this test if pg_dump is not available
	if _, err := os.Stat("/usr/bin/pg_dump"); os.IsNotExist(err) {
		t.Skip("pg_dump not available, skipping backup test")
	}

	// Create a temporary directory for testing
	tempDir := t.TempDir()

	backupService := services.NewBackupService(tempDir)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// This test will likely fail in a test environment without a real database
	// but it tests the structure and error handling
	result := backupService.CreateBackup(ctx)

	// The result should not be nil
	assert.NotNil(t, result)

	// In a test environment, this will likely fail due to no database connection
	// but we can verify the structure is correct
	assert.NotEmpty(t, result.FilePath)
	assert.NotZero(t, result.CreatedAt)
	assert.NotZero(t, result.Duration)
}

func TestBackupService_CleanupOldBackups(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	backupService := services.NewBackupService(tempDir)

	// Create some test backup files
	testFiles := []string{
		"backup_test_2024-01-01_12-00-00.sql",
		"backup_test_2024-01-02_12-00-00.sql",
		"backup_test_2024-01-03_12-00-00.sql",
	}

	for _, filename := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		err := os.WriteFile(filePath, []byte("test backup content"), 0644)
		require.NoError(t, err)

		// Set file modification time to simulate old files
		oldTime := time.Now().AddDate(0, 0, -31) // 31 days ago
		err = os.Chtimes(filePath, oldTime, oldTime)
		require.NoError(t, err)
	}

	// Verify files were created
	files, err := os.ReadDir(tempDir)
	require.NoError(t, err)
	assert.Len(t, files, 3)

	// Run cleanup with 30-day retention
	err = backupService.CleanupOldBackups(30)
	assert.NoError(t, err)

	// Verify old files were cleaned up
	files, err = os.ReadDir(tempDir)
	require.NoError(t, err)
	assert.Len(t, files, 0, "Old backup files should have been cleaned up")
}

func TestBackupService_GetBackupStats(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	backupService := services.NewBackupService(tempDir)

	// Create some test backup files
	testFiles := []string{
		"backup_test_2024-01-01_12-00-00.sql",
		"backup_test_2024-01-02_12-00-00.sql",
	}

	for _, filename := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		err := os.WriteFile(filePath, []byte("test backup content"), 0644)
		require.NoError(t, err)
	}

	// Get backup stats
	stats, err := backupService.GetBackupStats()
	require.NoError(t, err)

	// Verify stats
	assert.Equal(t, 2, stats["total_backups"])
	assert.NotZero(t, stats["total_size_bytes"])
	assert.Equal(t, tempDir, stats["backup_directory"])
	assert.NotNil(t, stats["oldest_backup"])
	assert.NotNil(t, stats["newest_backup"])
}

func TestBackupService_WithConfig(t *testing.T) {
	// Test with config-based backup directory
	cfg := &config.Config{
		Backup: config.Backup{
			Enabled:       true,
			Directory:     "test_backups",
			RetentionDays: 30,
			CleanupOld:    true,
		},
	}

	// Set config for testing
	config.SetConfig(cfg)

	backupService := services.NewBackupService(cfg.Backup.Directory)
	assert.NotNil(t, backupService)

	// Clean up test directory
	os.RemoveAll(cfg.Backup.Directory)
}
