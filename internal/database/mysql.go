package database

import (
	"database/sql"
	"fmt"
	"time"

	"go-rest/internal/config"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DB struct {
	Gorm *gorm.DB
	SQL  *sql.DB
}

func ConnectMySQL(cfg config.Config, log *zap.Logger) (DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=Local",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

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

