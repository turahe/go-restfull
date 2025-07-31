package db

import (
	"context"
	"fmt"
	"github.com/turahe/go-restfull/pkg/logger"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/turahe/go-restfull/config"
	"github.com/turahe/go-restfull/internal/db/pgx"
	"go.uber.org/zap"
)

// DatabaseService handles multiple database initialization and management
type DatabaseService struct {
	manager *pgx.DatabaseManager
}

// NewDatabaseService creates a new database service
func NewDatabaseService() *DatabaseService {
	return &DatabaseService{
		manager: pgx.GetDefaultManager(),
	}
}

// InitializeAll initializes all configured databases
func (ds *DatabaseService) InitializeAll(ctx context.Context) error {
	if logger.Log != nil {
		logger.Log.Info("Initializing all databases...")
	}

	// Initialize all databases from configuration
	if err := pgx.InitAllDatabases(); err != nil {
		return fmt.Errorf("failed to initialize databases: %w", err)
	}

	// Initialize schemas for all databases
	if err := ds.initializeSchemas(ctx); err != nil {
		return fmt.Errorf("failed to initialize schemas: %w", err)
	}

	// Perform health checks on all databases
	if err := ds.healthCheckAll(ctx); err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if logger.Log != nil {
		logger.Log.Info("All databases initialized successfully")
	}

	return nil
}

// InitializeByName initializes a specific database by name
func (ds *DatabaseService) InitializeByName(ctx context.Context, name string) error {
	if logger.Log != nil {
		logger.Log.Info("Initializing database", zap.String("name", name))
	}

	// Initialize the specific database
	if err := pgx.InitDatabaseByName(name); err != nil {
		return fmt.Errorf("failed to initialize database '%s': %w", name, err)
	}

	// Initialize schema for the database
	dbConfig := config.GetDatabaseConfig(name)
	if dbConfig != nil {
		if err := pgx.InitSchema(ctx, config.Postgres{
			Host:     dbConfig.Host,
			Port:     dbConfig.Port,
			Database: dbConfig.Database,
			Schema:   dbConfig.Schema,
			Username: dbConfig.Username,
			Password: dbConfig.Password,
		}, dbConfig.Schema); err != nil {
			return fmt.Errorf("failed to initialize schema for '%s': %w", name, err)
		}
	}

	// Perform health check
	if err := ds.manager.HealthCheck(ctx, name); err != nil {
		return fmt.Errorf("health check failed for '%s': %w", name, err)
	}

	if logger.Log != nil {
		logger.Log.Info("Database initialized successfully", zap.String("name", name))
	}

	return nil
}

// initializeSchemas initializes schemas for all configured databases
func (ds *DatabaseService) initializeSchemas(ctx context.Context) error {
	databases := config.GetAllDatabaseConfigs()

	for _, dbConfig := range databases {
		if err := pgx.InitSchema(ctx, config.Postgres{
			Host:     dbConfig.Host,
			Port:     dbConfig.Port,
			Database: dbConfig.Database,
			Schema:   dbConfig.Schema,
			Username: dbConfig.Username,
			Password: dbConfig.Password,
		}, dbConfig.Schema); err != nil {
			return fmt.Errorf("failed to initialize schema for '%s': %w", dbConfig.Name, err)
		}
	}

	return nil
}

// healthCheckAll performs health checks on all databases
func (ds *DatabaseService) healthCheckAll(ctx context.Context) error {
	databases := config.GetAllDatabaseConfigs()

	for _, dbConfig := range databases {
		if err := ds.manager.HealthCheck(ctx, dbConfig.Name); err != nil {
			return fmt.Errorf("health check failed for '%s': %w", dbConfig.Name, err)
		}
	}

	return nil
}

// GetConnection returns a connection from the default database
func (ds *DatabaseService) GetConnection(ctx context.Context) (*pgxpool.Conn, error) {
	return pgx.GetPgxConnWithContext(ctx)
}

// GetConnectionByName returns a connection from a specific database
func (ds *DatabaseService) GetConnectionByName(ctx context.Context, name string) (*pgxpool.Conn, error) {
	return pgx.GetPgxConnWithContextByName(ctx, name)
}

// HealthCheck performs health check on all databases
func (ds *DatabaseService) HealthCheck(ctx context.Context) error {
	databases := config.GetAllDatabaseConfigs()

	for _, dbConfig := range databases {
		if err := ds.manager.HealthCheck(ctx, dbConfig.Name); err != nil {
			return fmt.Errorf("health check failed for '%s': %w", dbConfig.Name, err)
		}
	}

	return nil
}

// CloseAll closes all database connections
func (ds *DatabaseService) CloseAll() {
	ds.manager.CloseAll()
}

// ListConnections returns a list of all active database connections
func (ds *DatabaseService) ListConnections() []string {
	return ds.manager.ListConnections()
}

// HasConnection checks if a specific database connection exists
func (ds *DatabaseService) HasConnection(name string) bool {
	return ds.manager.HasConnection(name)
}
