package logger

import (
	"sync"

	"github.com/turahe/go-restfull/config"
	"go.uber.org/zap"
)

var Log *zap.Logger
var m sync.Mutex

func InitLogger(logDriver string) {
	m.Lock()
	defer m.Unlock()

	Log = newZapLogger()

	// Log that the logger has been initialized
	if Log != nil {
		Log.Info("Logger initialized successfully",
			zap.String("driver", logDriver),
			zap.Bool("file_enabled", config.GetConfig().Log.FileEnabled),
			zap.String("file_path", config.GetConfig().Log.FilePath),
			zap.String("level", config.GetConfig().Log.Level),
		)
	}
}
