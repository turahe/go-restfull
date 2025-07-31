package testutils

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/turahe/go-restfull/pkg/logger"
	"log"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/turahe/go-restfull/config"
	"github.com/turahe/go-restfull/internal/db/pgx"
	"github.com/turahe/go-restfull/internal/db/rdb"
	"github.com/turahe/go-restfull/internal/infrastructure/container"
)

// TestSetup contains all test dependencies
type TestSetup struct {
	DB          *pgxpool.Pool
	RedisClient redis.Cmdable
	Container   *container.Container
	Cleanup     func()
}

// SetupTestDatabase initializes a test database
func SetupTestDatabase(t *testing.T) *pgxpool.Pool {
	// Load test configuration
	cfg := config.GetConfig()

	// Create database connection string
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable search_path=%s",
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.Username,
		cfg.Postgres.Password,
		cfg.Postgres.Database,
		cfg.Postgres.Schema,
	)

	// Connect to database
	connConfig, err := pgxpool.ParseConfig(dsn)
	require.NoError(t, err, "Failed to parse database config")

	connConfig.MaxConns = int32(cfg.Postgres.MaxConnections)

	db, err := pgxpool.NewWithConfig(context.Background(), connConfig)
	require.NoError(t, err, "Failed to connect to test database")

	// Test connection
	err = db.Ping(context.Background())
	require.NoError(t, err, "Failed to ping test database")

	return db
}

// SetupTestRedis initializes a test Redis client
func SetupTestRedis(t *testing.T) redis.Cmdable {
	cfg := config.GetConfig()

	// Use the first Redis configuration
	if len(cfg.Redis) == 0 {
		t.Skip("No Redis configuration available")
	}

	redisConfig := cfg.Redis[0]
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
		Password: redisConfig.Password,
		DB:       redisConfig.Database,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Ping(ctx).Err()
	require.NoError(t, err, "Failed to connect to test Redis")

	return client
}

// SetupTestContainer creates a test container with all dependencies
func SetupTestContainer(t *testing.T) *TestSetup {
	// Initialize logger
	logger.InitLogger("zap")

	// Setup database
	db := SetupTestDatabase(t)

	// Setup Redis
	redisClient := SetupTestRedis(t)

	// Initialize database connections
	pgx.InitPgConnectionPool(config.GetConfig().Postgres)
	rdb.InitRedisClient(config.GetConfig().Redis)

	// Create container
	container := container.NewContainer(db)

	// Cleanup function
	cleanup := func() {
		// Clean up test data
		cleanupTestData(t, db)

		// Close connections
		db.Close()
		pgx.ClosePgxPool()
	}

	return &TestSetup{
		DB:          db,
		RedisClient: redisClient,
		Container:   container,
		Cleanup:     cleanup,
	}
}

// cleanupTestData cleans up test data from the database
func cleanupTestData(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()

	// List of tables to clean up (in reverse dependency order)
	tables := []string{
		"user_roles",
		"menu_roles",
		"comments",
		"posts",
		"tags",
		"taxonomies",
		"contents",
		"media",
		"settings",
		"jobs",
		"users",
		"roles",
		"menus",
	}

	for _, table := range tables {
		_, err := db.Exec(ctx, fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			log.Printf("Warning: Failed to clean up table %s: %v", table, err)
		}
	}
}

// CreateTestUser creates a test user for testing
func CreateTestUser(t *testing.T, db *pgxpool.Pool, username, email, password string) string {
	ctx := context.Background()

	// Hash password
	hashedPassword := "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi" // "password"

	// Insert test user
	var userID string
	err := db.QueryRow(ctx, `
		INSERT INTO users (id, username, email, phone, password, email_verified_at, phone_verified_at, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW(), NOW(), NOW(), NOW())
		RETURNING id::text
	`, username, email, "+1234567890", hashedPassword).Scan(&userID)

	require.NoError(t, err, "Failed to create test user")
	return userID
}

// CreateTestRole creates a test role for testing
func CreateTestRole(t *testing.T, db *pgxpool.Pool, name, slug, description string) string {
	ctx := context.Background()

	var roleID string
	err := db.QueryRow(ctx, `
		INSERT INTO roles (id, name, slug, description, is_active, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, true, NOW(), NOW())
		RETURNING id::text
	`, name, slug, description).Scan(&roleID)

	require.NoError(t, err, "Failed to create test role")
	return roleID
}

// AssignUserRole assigns a role to a user for testing
func AssignUserRole(t *testing.T, db *pgxpool.Pool, userID, roleID string) {
	ctx := context.Background()

	_, err := db.Exec(ctx, `
		INSERT INTO user_roles (id, user_id, role_id, created_at, updated_at)
		VALUES (gen_random_uuid(), $1::uuid, $2::uuid, NOW(), NOW())
	`, userID, roleID)

	require.NoError(t, err, "Failed to assign user role")
}

// GenerateTestToken generates a test JWT token for testing
func GenerateTestToken(t *testing.T, userID string) string {
	// This is a simplified token generation for testing
	// In real tests, you might want to use the actual JWT service
	return "test-token-" + userID
}

// AssertJSONResponse asserts that a JSON response matches expected structure
func AssertJSONResponse(t *testing.T, body []byte, expectedCode int, expectedMessage string) {
	var response map[string]interface{}
	err := json.Unmarshal(body, &response)
	require.NoError(t, err, "Failed to unmarshal JSON response")

	if expectedCode != 0 {
		require.Equal(t, float64(expectedCode), response["response_code"], "Response code mismatch")
	}

	if expectedMessage != "" {
		require.Equal(t, expectedMessage, response["response_message"], "Response message mismatch")
	}
}

// CreateTestPost creates a test post for testing
func CreateTestPost(t *testing.T, db *pgxpool.Pool, title, content, authorID string) string {
	ctx := context.Background()

	var postID string
	err := db.QueryRow(ctx, `
		INSERT INTO posts (id, title, slug, content, excerpt, author_id, status, published_at, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5::uuid, 'draft', NULL, NOW(), NOW())
		RETURNING id::text
	`, title, title, content, content[:100], authorID).Scan(&postID)

	require.NoError(t, err, "Failed to create test post")
	return postID
}

// CreateTestTaxonomy creates a test taxonomy for testing
func CreateTestTaxonomy(t *testing.T, db *pgxpool.Pool, name, slug, description string) string {
	ctx := context.Background()

	var taxonomyID string
	err := db.QueryRow(ctx, `
		INSERT INTO taxonomies (id, name, slug, description, record_ordering, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, 0, NOW(), NOW())
		RETURNING id::text
	`, name, slug, description).Scan(&taxonomyID)

	require.NoError(t, err, "Failed to create test taxonomy")
	return taxonomyID
}

// CreateTestMenu creates a test menu for testing
func CreateTestMenu(t *testing.T, db *pgxpool.Pool, name, slug, url string) string {
	ctx := context.Background()

	var menuID string
	err := db.QueryRow(ctx, `
		INSERT INTO menus (id, name, slug, url, parent_id, record_ordering, is_active, is_visible, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, NULL, 0, true, true, NOW(), NOW())
		RETURNING id::text
	`, name, slug, url).Scan(&menuID)

	require.NoError(t, err, "Failed to create test menu")
	return menuID
}

// WaitForCondition waits for a condition to be true with timeout
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatal("Condition not met within timeout")
}

// SkipIfNoDatabase skips the test if database is not available
func SkipIfNoDatabase(t *testing.T) {
	cfg := config.GetConfig()
	if cfg == nil {
		t.Skip("Config not loaded, skipping database test")
		return
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.Username,
		cfg.Postgres.Password,
		cfg.Postgres.Database,
	)

	connConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Skip("Database not available, skipping test")
	}

	db, err := pgxpool.NewWithConfig(context.Background(), connConfig)
	if err != nil {
		t.Skip("Database not available, skipping test")
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		t.Skip("Database not available, skipping test")
	}
}

// SkipIfNoRedis skips the test if Redis is not available
func SkipIfNoRedis(t *testing.T) {
	cfg := config.GetConfig()

	if len(cfg.Redis) == 0 {
		t.Skip("Redis not available, skipping test")
	}

	redisConfig := cfg.Redis[0]
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
		Password: redisConfig.Password,
		DB:       redisConfig.Database,
	})
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available, skipping test")
	}
}
