package pgx

import (
	"context"
	"fmt"
	"github.com/turahe/go-restfull/pkg/logger"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/turahe/go-restfull/config"
	"go.uber.org/zap"
)

// DatabaseManager manages multiple database connections
type DatabaseManager struct {
	pools map[string]*pgxpool.Pool
	once  sync.Once
	mu    sync.RWMutex
}

// NewDatabaseManager creates a new database manager
func NewDatabaseManager() *DatabaseManager {
	return &DatabaseManager{
		pools: make(map[string]*pgxpool.Pool),
	}
}

// Global database manager instance
var (
	defaultManager *DatabaseManager
	managerOnce    sync.Once
)

// GetDefaultManager returns the default database manager instance
func GetDefaultManager() *DatabaseManager {
	managerOnce.Do(func() {
		defaultManager = NewDatabaseManager()
	})
	return defaultManager
}

// Initialize the database connection pool with thread-safe lazy initialization.
func InitPgConnectionPool(postgresConfig config.Postgres) error {
	return GetDefaultManager().InitConnectionLegacy("default", postgresConfig)
}

// InitAllDatabases initializes all configured databases
func InitAllDatabases() error {
	manager := GetDefaultManager()

	// Initialize all databases from config
	for _, dbConfig := range config.GetAllDatabaseConfigs() {
		if err := manager.InitConnection(dbConfig.Name, dbConfig); err != nil {
			return fmt.Errorf("failed to initialize database '%s': %w", dbConfig.Name, err)
		}
	}

	// If no databases are configured, initialize with legacy config
	if len(config.GetAllDatabaseConfigs()) == 0 {
		legacyConfig := config.GetConfig().Postgres
		if err := manager.InitConnectionLegacy("default", legacyConfig); err != nil {
			return fmt.Errorf("failed to initialize default database: %w", err)
		}
	}

	return nil
}

// InitDatabaseByName initializes a specific database by name
func InitDatabaseByName(name string) error {
	manager := GetDefaultManager()

	dbConfig := config.GetDatabaseConfig(name)
	if dbConfig == nil {
		return fmt.Errorf("database configuration '%s' not found", name)
	}

	return manager.InitConnection(name, *dbConfig)
}

// InitConnection initializes a named database connection
func (dm *DatabaseManager) InitConnection(name string, dbConfig config.DatabaseConfig) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// Check if connection already exists
	if _, exists := dm.pools[name]; exists {
		return nil
	}

	// Convert DatabaseConfig to Postgres for connection string
	postgresConfig := config.Postgres{
		Host:           dbConfig.Host,
		Port:           dbConfig.Port,
		Database:       dbConfig.Database,
		Schema:         dbConfig.Schema,
		Username:       dbConfig.Username,
		Password:       dbConfig.Password,
		MaxConnections: dbConfig.MaxConnections,
	}

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable search_path=%s",
		postgresConfig.Host,
		postgresConfig.Port,
		postgresConfig.Username,
		postgresConfig.Password,
		postgresConfig.Database,
		postgresConfig.Schema,
	)

	connConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		if logger.Log != nil {
			logger.Log.Error("Failed to parse config:", zap.Error(err), zap.String("connection", name))
		}
		return err
	}

	// Set maximum number of connections
	connConfig.MaxConns = int32(postgresConfig.MaxConnections)
	// connConfig.MaxConnIdleTime = time.Duration(postgresConfig.MaxConnIdleTime) * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), connConfig)
	if err != nil {
		if logger.Log != nil {
			logger.Log.Error("Failed to create connection pool:", zap.Error(err), zap.String("connection", name))
		}
		return err
	}

	dm.pools[name] = pool

	if logger.Log != nil {
		logger.Log.Info("PostgreSQL connection pool initialized successfully", zap.String("connection", name))
	}
	return nil
}

// InitConnectionLegacy initializes a named database connection with legacy Postgres config
func (dm *DatabaseManager) InitConnectionLegacy(name string, postgresConfig config.Postgres) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// Check if connection already exists
	if _, exists := dm.pools[name]; exists {
		return nil
	}

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable search_path=%s",
		postgresConfig.Host,
		postgresConfig.Port,
		postgresConfig.Username,
		postgresConfig.Password,
		postgresConfig.Database,
		postgresConfig.Schema,
	)

	connConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		if logger.Log != nil {
			logger.Log.Error("Failed to parse config:", zap.Error(err), zap.String("connection", name))
		}
		return err
	}

	// Set maximum number of connections
	connConfig.MaxConns = int32(postgresConfig.MaxConnections)
	// connConfig.MaxConnIdleTime = time.Duration(postgresConfig.MaxConnIdleTime) * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), connConfig)
	if err != nil {
		if logger.Log != nil {
			logger.Log.Error("Failed to create connection pool:", zap.Error(err), zap.String("connection", name))
		}
		return err
	}

	dm.pools[name] = pool

	if logger.Log != nil {
		logger.Log.Info("PostgreSQL connection pool initialized successfully", zap.String("connection", name))
	}
	return nil
}

// GetPool returns a specific named connection pool
func (dm *DatabaseManager) GetPool(name string) *pgxpool.Pool {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.pools[name]
}

// GetDefaultPool returns the default connection pool
func (dm *DatabaseManager) GetDefaultPool() *pgxpool.Pool {
	return dm.GetPool("default")
}

// GetPgxPool returns the default connection pool, initializing it if necessary.
func GetPgxPool() *pgxpool.Pool {
	manager := GetDefaultManager()

	// Ensure default pool is initialized
	if err := manager.InitConnectionLegacy("default", config.GetConfig().Postgres); err != nil {
		if logger.Log != nil {
			logger.Log.Error("Failed to initialize default connection pool", zap.Error(err))
		}
		return nil
	}
	return manager.GetDefaultPool()
}

// GetPgxPoolByName returns a specific named connection pool
func GetPgxPoolByName(name string) *pgxpool.Pool {
	manager := GetDefaultManager()
	return manager.GetPool(name)
}

// GetPgxConn returns a connection from the default pool.
func GetPgxConn() *pgxpool.Conn {
	pool := GetPgxPool()
	if pool == nil {
		if logger.Log != nil {
			logger.Log.Error("Cannot acquire connection: default pool is nil")
		}
		return nil
	}

	conn, err := pool.Acquire(context.Background())
	if err != nil {
		if logger.Log != nil {
			logger.Log.Error("Failed to acquire connection from default pool", zap.Error(err))
		}
		return nil
	}

	return conn
}

// GetPgxConnByName returns a connection from a specific named pool.
func GetPgxConnByName(name string) *pgxpool.Conn {
	pool := GetPgxPoolByName(name)
	if pool == nil {
		if logger.Log != nil {
			logger.Log.Error("Cannot acquire connection: pool is nil", zap.String("connection", name))
		}
		return nil
	}

	conn, err := pool.Acquire(context.Background())
	if err != nil {
		if logger.Log != nil {
			logger.Log.Error("Failed to acquire connection from pool", zap.Error(err), zap.String("connection", name))
		}
		return nil
	}

	return conn
}

// GetPgxConnWithContext returns a connection from the default pool with a context.
func GetPgxConnWithContext(ctx context.Context) (*pgxpool.Conn, error) {
	pool := GetPgxPool()
	if pool == nil {
		return nil, fmt.Errorf("default connection pool is not initialized")
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %w", err)
	}

	return conn, nil
}

// GetPgxConnWithContextByName returns a connection from a specific named pool with a context.
func GetPgxConnWithContextByName(ctx context.Context, name string) (*pgxpool.Conn, error) {
	pool := GetPgxPoolByName(name)
	if pool == nil {
		return nil, fmt.Errorf("connection pool '%s' is not initialized", name)
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection from '%s': %w", name, err)
	}

	return conn, nil
}

// InitSchema initializes a schema for a specific database connection
func InitSchema(ctx context.Context, postgresConfig config.Postgres, schema string) (err error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		postgresConfig.Host,
		postgresConfig.Port,
		postgresConfig.Username,
		postgresConfig.Password,
		postgresConfig.Database,
	)

	pgConn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return err
	}
	defer pgConn.Close(ctx)

	// Create schema if it doesn't exist
	// Ignore error if schema already exists or if the user doesn't have permission to create schema
	pgConn.Exec(
		ctx,
		fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s`, schema),
	)

	// Set search path to schema so that we don't have to specify the schema name
	_, err = pgConn.Exec(
		ctx,
		fmt.Sprintf(`SET search_path TO %s`, schema),
	)
	if err != nil {
		return err
	}

	return nil
}

// Close closes a specific named connection pool
func (dm *DatabaseManager) Close(name string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if pool, exists := dm.pools[name]; exists {
		pool.Close()
		delete(dm.pools, name)
		if logger.Log != nil {
			logger.Log.Info("PostgreSQL connection pool closed", zap.String("connection", name))
		}
	}
}

// CloseAll closes all connection pools
func (dm *DatabaseManager) CloseAll() {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	for name, pool := range dm.pools {
		pool.Close()
		if logger.Log != nil {
			logger.Log.Info("PostgreSQL connection pool closed", zap.String("connection", name))
		}
	}
	dm.pools = make(map[string]*pgxpool.Pool)
}

// Close the default database connection pool.
func ClosePgxPool() {
	GetDefaultManager().Close("default")
}

// CloseAll closes all database connection pools.
func CloseAllPools() {
	GetDefaultManager().CloseAll()
}

// HealthCheck performs a health check on the default connection pool.
func HealthCheck(ctx context.Context) error {
	return GetDefaultManager().HealthCheck(ctx, "default")
}

// HealthCheckByName performs a health check on a specific named connection pool.
func (dm *DatabaseManager) HealthCheck(ctx context.Context, name string) error {
	pool := dm.GetPool(name)
	if pool == nil {
		return fmt.Errorf("connection pool '%s' is not initialized", name)
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection for health check: %w", err)
	}
	defer conn.Release()

	return conn.Ping(ctx)
}

// ListConnections returns a list of all active connection names
func (dm *DatabaseManager) ListConnections() []string {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	connections := make([]string, 0, len(dm.pools))
	for name := range dm.pools {
		connections = append(connections, name)
	}
	return connections
}

// HasConnection checks if a specific named connection exists
func (dm *DatabaseManager) HasConnection(name string) bool {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	_, exists := dm.pools[name]
	return exists
}
