package database

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"strings"
	"time"

	"go-rest/internal/config"

	"go.uber.org/zap"
	"cloud.google.com/go/cloudsqlconn"
	sqlmysql "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DB struct {
	Gorm *gorm.DB
	SQL  *sql.DB
}

func ConnectMySQL(cfg config.Config, log *zap.Logger) (DB, error) {
	driver := strings.ToLower(strings.TrimSpace(cfg.DBDriver))
	if driver == "" {
		driver = "mysql"
	}

	var dsn string
	switch driver {
	case "mysql-cloud":
		if strings.TrimSpace(cfg.CloudSQLInstanceConnectionName) == "" {
			return DB{}, fmt.Errorf("INSTANCE_CONNECTION_NAME is required for mysql-cloud")
		}

		// Cloud SQL Go Connector dialer.
		d, err := cloudsqlconn.NewDialer(context.Background(), cloudsqlconn.WithLazyRefresh())
		if err != nil {
			return DB{}, fmt.Errorf("cloudsqlconn.NewDialer: %w", err)
		}

		// Register a dialer for go-sql-driver/mysql.
		var opts []cloudsqlconn.DialOption
		if cfg.CloudSQLPrivateIP {
			opts = append(opts, cloudsqlconn.WithPrivateIP())
		}

		sqlmysql.RegisterDialContext("cloudsqlconn",
			func(ctx context.Context, addr string) (net.Conn, error) {
				// The addr is ignored by convention; we use INSTANCE_CONNECTION_NAME.
				return d.Dial(ctx, cfg.CloudSQLInstanceConnectionName, opts...)
			},
		)

		// cloudsqlconn DSN format:
		//   user:pass@cloudsqlconn(localhost:3306)/dbname?parseTime=true
		dsn = fmt.Sprintf("%s:%s@cloudsqlconn(localhost:3306)/%s?parseTime=true&charset=utf8mb4&loc=Local",
			cfg.DBUser,
			cfg.DBPassword,
			cfg.DBName,
		)

	default: // "mysql"
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=Local",
			cfg.DBUser,
			cfg.DBPassword,
			cfg.DBHost,
			cfg.DBPort,
			cfg.DBName,
		)
	}

	gormDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: NewZapGormLogger(log),
	})
	if err != nil {
		return DB{}, err
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return DB{}, err
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	return DB{Gorm: gormDB, SQL: sqlDB}, nil
}

