package testutil

import (
	"os"
	"strings"

	"gorm.io/gorm/logger"
)

// GormLogLevelFromEnv returns the GORM SQL logging level used by unit tests/benchmarks.
//
// Env:
//   - GORM_SQL_LOG_ALL=1  => logger.Info
//   - GORM_SQL_LOG_LEVEL=silent|error|warn|info
//
// Default is `logger.Silent` to avoid noisy test output.
func GormLogLevelFromEnv() logger.LogLevel {
	level := logger.Silent

	if strings.TrimSpace(os.Getenv("GORM_SQL_LOG_ALL")) != "" {
		switch strings.ToLower(strings.TrimSpace(os.Getenv("GORM_SQL_LOG_ALL"))) {
		case "1", "true", "yes", "y", "on":
			level = logger.Info
		}
	}

	if v := strings.TrimSpace(os.Getenv("GORM_SQL_LOG_LEVEL")); v != "" {
		switch strings.ToLower(v) {
		case "silent":
			level = logger.Silent
		case "error":
			level = logger.Error
		case "warn", "warning":
			level = logger.Warn
		case "info":
			level = logger.Info
		}
	}

	return level
}

