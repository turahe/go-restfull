package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
	internal_minio "webapi/pkg/minio"

	"github.com/getsentry/sentry-go"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"webapi/config"
	"webapi/internal/db/pgx"
	"webapi/internal/db/rdb"
	"webapi/internal/logger"
)

const defaultConfigFile = "config/config.yaml"

var RootCmdName = "main"

var configFile string
var rootCmd = &cobra.Command{
	Use: func() string {
		return RootCmdName
	}(),
	Short: "\nThis application is made with ❤️",
	Run: func(cmd *cobra.Command, _ []string) {
		cmd.Usage()
	},
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", fmt.Sprintf("config file (default is %s)", defaultConfigFile))
}

func setupAll() {
	setUpConfig()
	setUpLogger()
	setUpPostgres()
	setUpRedis()
	setUpSentry()
	setUpMinio()

}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("rootCmd.Execute() Error: %v", err)
		os.Exit(1)
	}
}

func setUpConfig() {
	if configFile == "" {
		configFile = defaultConfigFile
	}

	log.Default().Printf("Using config file: %s", configFile)
	config.SetConfig(configFile)
}

func setUpLogger() {
	log.Default().Printf("Using log level: %s", config.GetConfig().Log.Level)
	logger.InitLogger("zap")
}

func setUpPostgres() {
	// Create the database connection pool
	if config.GetConfig().Postgres.Host != "" {
		if config.GetConfig().Postgres.Schema == "" {
			logger.Log.Fatal("Postgres schema is not set")
		}

		// Initialize database schema if it doesn't exist
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		logger.Log.Info("Initializing database schema", zap.String("schema", config.GetConfig().Postgres.Schema))
		err := pgx.InitSchema(ctx, config.GetConfig().Postgres, config.GetConfig().Postgres.Schema)
		if err != nil {
			logger.Log.Fatal("pgx.InitSchema()", zap.Error(err))
		}

		logger.Log.Info("Initializing pgxPool")
		err = pgx.InitPgConnectionPool(config.GetConfig().Postgres)
		if err != nil {
			logger.Log.Fatal("pgx.InitPgConnectionPool()", zap.Error(err))
		}
		logger.Log.Info("pgxPool initialized")
	}

}

func setUpRedis() {
	// Create the database connection pool
	if config.GetConfig().Redis[0].Host != "" {
		logger.Log.Info("Initializing redis")
		err := rdb.InitRedisClient(config.GetConfig().Redis)
		if err != nil {
			logger.Log.Fatal("rdb.InitRedisClient()", zap.Error(err))
		}
		logger.Log.Info("redis initialized")
	}

}

func setUpSentry() {
	// Don't initialize sentry if DSN is not set
	if config.GetConfig().Sentry.Dsn == "" {
		return
	}

	// Initialize sentry
	logger.Log.Info("Initializing Sentry " + config.GetConfig().Sentry.Dsn)
	err := sentry.Init(sentry.ClientOptions{
		Dsn: config.GetConfig().Sentry.Dsn,
		// BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
		// 	if hint.Context != nil {
		// 		if c, ok := hint.Context.Value(sentry.RequestContextKey).(*fiber.Ctx); ok {
		// 			// You have access to the original Context if it panicked
		// 			fmt.Println(utils.CopyString(c.Hostname()))
		// 		}
		// 	}
		// 	fmt.Println(event)
		// 	return event
		// },
		Debug:            config.GetConfig().Sentry.Debug,
		AttachStacktrace: true,
		EnableTracing:    true,
		TracesSampleRate: 1.0,
		Environment:      config.GetConfig().Sentry.Environment,
		Release:          config.GetConfig().Sentry.Release,
	})

	if err != nil {
		logger.Log.Error("Create Sentry instant error: %v", zap.Error(err))
		return
	}

	logger.Log.Info("Create Sentry instant success")

	// send initial event to sentry with data
	sentry.CaptureMessage("Sentry initialized")

	defer sentry.Flush(2 * time.Second)
}

func setUpMinio() {
	// Don't initialize minio if it is not enabled
	if !config.GetConfig().Minio.Enable {
		return
	}

	logger.Log.Info("Initializing Minio")
	err := internal_minio.Setup()
	if err != nil {
		logger.Log.Fatal("internal_minio.Setup()", zap.Error(err))
	}
	logger.Log.Info("Minio initialized")
}
