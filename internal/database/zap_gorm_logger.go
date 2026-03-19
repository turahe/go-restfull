package database

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const requestIDContextKey = "request_id"

// zapGormLogger bridges GORM's logger.Interface to zap.
// It logs slow SQL, SQL errors, and request-scoped trace data.
type zapGormLogger struct {
	base               *zap.Logger
	logLevel           logger.LogLevel
	slowThreshold     time.Duration
	ignoreRecordNotFound bool
}

func NewZapGormLogger(base *zap.Logger) logger.Interface {
	if base == nil {
		base = zap.NewNop()
	}

	// Optional global switch for noisy SQL logging.
	// - `GORM_SQL_LOG_ALL=1` => log every query at Info level
	// - `GORM_SQL_LOG_LEVEL=silent|error|warn|info` => explicit override
	level := logger.Warn
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

	return &zapGormLogger{
		base:           base,
		logLevel:       level,
		slowThreshold: 200 * time.Millisecond,
		// These "not found" errors are expected in many read paths (e.g. config
		// optional rows). Avoid noisy logs for them.
		ignoreRecordNotFound: true,
	}
}

func (l *zapGormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.logLevel = level
	return &newLogger
}

func requestIDFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	v := ctx.Value(requestIDContextKey)
	s, ok := v.(string)
	return s, ok && s != ""
}

func (l *zapGormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel < logger.Info {
		return
	}
	fields := []zap.Field{
		zap.String("msg", msg),
		zap.Any("data", data),
	}
	if rid, ok := requestIDFromContext(ctx); ok {
		fields = append(fields, zap.String("request_id", rid))
	}
	l.base.Info(msg, fields...)
}

func (l *zapGormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel < logger.Warn {
		return
	}
	fields := []zap.Field{
		zap.String("msg", msg),
		zap.Any("data", data),
	}
	if rid, ok := requestIDFromContext(ctx); ok {
		fields = append(fields, zap.String("request_id", rid))
	}
	l.base.Warn(msg, fields...)
}

func (l *zapGormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel < logger.Error {
		return
	}
	fields := []zap.Field{
		zap.String("msg", msg),
		zap.Any("data", data),
	}
	if rid, ok := requestIDFromContext(ctx); ok {
		fields = append(fields, zap.String("request_id", rid))
	}
	l.base.Error(msg, fields...)
}

func (l *zapGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	// If we were configured to be silent at this level, skip.
	if l.logLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	rid, hasRID := requestIDFromContext(ctx)

	withRID := func(fields []zap.Field) []zap.Field {
		if hasRID {
			fields = append(fields, zap.String("request_id", rid))
		}
		return fields
	}

	// Mimic gorm's built-in logic: errors first.
	if err != nil {
		if l.logLevel < logger.Error {
			return
		}
		// Optionally ignore record-not-found errors.
		if l.ignoreRecordNotFound && errors.Is(err, gorm.ErrRecordNotFound) {
			return
		}

		fields := []zap.Field{
			zap.Duration("duration", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
			zap.Error(err),
		}
		fields = withRID(fields)
		l.base.Error("gorm sql error", fields...)
		return
	}

	// Slow SQL
	if l.slowThreshold > 0 && elapsed >= l.slowThreshold && l.logLevel >= logger.Warn {
		fields := []zap.Field{
			zap.Duration("duration", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		}
		fields = withRID(fields)
		l.base.Warn("gorm slow sql", fields...)
		return
	}

	// Info-level SQL logging
	if l.logLevel >= logger.Info && elapsed > 0 {
		fields := []zap.Field{
			zap.Duration("duration", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		}
		fields = withRID(fields)
		l.base.Info("gorm sql", fields...)
	}
}

