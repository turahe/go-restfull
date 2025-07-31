# Multi-Database Setup Guide

This guide explains how to set up and use multiple databases in the Go RESTful API application.

## Overview

The application now supports multiple database connections, allowing you to:
- Connect to different databases for different purposes (primary, read replicas, analytics, etc.)
- Manage connections independently
- Perform health checks on individual databases
- Use different configurations for each database

## Configuration

### 1. Database Configuration Structure

Add a `databases` section to your `config.yaml` file:

```yaml
# Multi-database configuration
databases:
  # Primary database (default)
  - name: "primary"
    host: "localhost"
    port: 5432
    database: "primary_db"
    schema: "public"
    username: "primary_user"
    password: "primary_secret"
    maxConnections: 20
    connectionTimeout: 30
    idleTimeout: 300
    maxIdleConns: 10
    maxOpenConns: 100
    sslMode: "disable"
    isDefault: true

  # Secondary database (read replica)
  - name: "secondary"
    host: "localhost"
    port: 5433
    database: "secondary_db"
    schema: "public"
    username: "secondary_user"
    password: "secondary_secret"
    maxConnections: 15
    connectionTimeout: 30
    idleTimeout: 300
    maxIdleConns: 5
    maxOpenConns: 50
    sslMode: "disable"
    isDefault: false

  # Analytics database
  - name: "analytics"
    host: "analytics-db.example.com"
    port: 5432
    database: "analytics_db"
    schema: "analytics"
    username: "analytics_user"
    password: "analytics_secret"
    maxConnections: 10
    connectionTimeout: 30
    idleTimeout: 300
    maxIdleConns: 3
    maxOpenConns: 20
    sslMode: "require"
    isDefault: false
```

### 2. Configuration Parameters

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `name` | string | Unique name for the database connection | Required |
| `host` | string | Database host address | Required |
| `port` | int | Database port | 5432 |
| `database` | string | Database name | Required |
| `schema` | string | Database schema | "public" |
| `username` | string | Database username | Required |
| `password` | string | Database password | Required |
| `maxConnections` | int | Maximum number of connections in pool | 20 |
| `connectionTimeout` | int | Connection timeout in seconds | 30 |
| `idleTimeout` | int | Idle connection timeout in seconds | 300 |
| `maxIdleConns` | int | Maximum number of idle connections | 10 |
| `maxOpenConns` | int | Maximum number of open connections | 100 |
| `sslMode` | string | SSL mode (disable, require, verify-ca, verify-full) | "disable" |
| `isDefault` | bool | Mark as default database | false |

## Usage Examples

### 1. Basic Database Service Usage

```go
package main

import (
    "context"
    "log"

    "github.com/turahe/go-restfull/internal/db"
)

func main() {
    ctx := context.Background()
    
    // Create database service
    dbService := db.NewDatabaseService()
    
    // Initialize all databases
    if err := dbService.InitializeAll(ctx); err != nil {
        log.Fatalf("Failed to initialize databases: %v", err)
    }
    
    // Get connection from default database
    conn, err := dbService.GetConnection(ctx)
    if err != nil {
        log.Fatalf("Failed to get connection: %v", err)
    }
    defer conn.Release()
    
    // Use connection for queries
    var result string
    err = conn.QueryRow(ctx, "SELECT version()").Scan(&result)
    if err != nil {
        log.Printf("Query failed: %v", err)
    } else {
        log.Printf("Database version: %s", result)
    }
}
```

### 2. Using Specific Databases

```go
// Get connection from specific database
conn, err := dbService.GetConnectionByName(ctx, "analytics")
if err != nil {
    log.Fatalf("Failed to get analytics connection: %v", err)
}
defer conn.Release()

// Execute analytics query
var count int
err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM analytics.events").Scan(&count)
if err != nil {
    log.Printf("Analytics query failed: %v", err)
} else {
    log.Printf("Total events: %d", count)
}
```

### 3. Health Checking

```go
// Health check all databases
if err := dbService.HealthCheck(ctx); err != nil {
    log.Printf("Health check failed: %v", err)
}

// List all active connections
connections := dbService.ListConnections()
log.Printf("Active connections: %v", connections)

// Check if specific database exists
if dbService.HasConnection("analytics") {
    log.Println("Analytics database is available")
}
```

### 4. Direct PGX Usage

```go
import (
    "github.com/turahe/go-restfull/internal/db/pgx"
)

// Initialize specific database
err := pgx.InitDatabaseByName("analytics")
if err != nil {
    log.Fatalf("Failed to initialize analytics database: %v", err)
}

// Get connection with context
conn, err := pgx.GetPgxConnWithContextByName(ctx, "analytics")
if err != nil {
    log.Fatalf("Failed to get connection: %v", err)
}
defer conn.Release()

// Get connection pool
pool := pgx.GetPgxPoolByName("analytics")
if pool == nil {
    log.Fatal("Analytics pool is not available")
}
```

### 5. Repository Pattern with Multiple Databases

```go
type UserRepository struct {
    primaryDB   *pgxpool.Pool
    analyticsDB *pgxpool.Pool
}

func NewUserRepository() *UserRepository {
    return &UserRepository{
        primaryDB:   pgx.GetPgxPoolByName("primary"),
        analyticsDB: pgx.GetPgxPoolByName("analytics"),
    }
}

func (r *UserRepository) CreateUser(ctx context.Context, user *User) error {
    // Use primary database for writes
    conn, err := r.primaryDB.Acquire(ctx)
    if err != nil {
        return err
    }
    defer conn.Release()
    
    // Insert user into primary database
    return conn.QueryRow(ctx, "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id",
        user.Name, user.Email).Scan(&user.ID)
}

func (r *UserRepository) GetUserAnalytics(ctx context.Context, userID int) (*UserAnalytics, error) {
    // Use analytics database for reads
    conn, err := r.analyticsDB.Acquire(ctx)
    if err != nil {
        return nil, err
    }
    defer conn.Release()
    
    // Query analytics data
    var analytics UserAnalytics
    err = conn.QueryRow(ctx, "SELECT * FROM analytics.user_stats WHERE user_id = $1", userID).
        Scan(&analytics.UserID, &analytics.LoginCount, &analytics.LastLogin)
    
    return &analytics, err
}
```

## Migration from Single Database

### 1. Backward Compatibility

The application maintains backward compatibility with the existing single database configuration:

```yaml
# Legacy configuration (still works)
postgres:
  host: "localhost"
  port: 5432
  database: "my_db"
  schema: "public"
  username: "my_user"
  password: "secret"
  maxConnections: 20
```

### 2. Gradual Migration

1. **Start with existing configuration**: The application will use the legacy `postgres` config as the default database.

2. **Add new databases**: Add the `databases` section to your configuration file.

3. **Update code gradually**: Replace direct `pgx` calls with the new multi-database API.

4. **Remove legacy config**: Once migration is complete, you can remove the legacy `postgres` section.

## Best Practices

### 1. Database Naming

Use descriptive names for your databases:
- `primary` - Main application database
- `secondary` - Read replica
- `analytics` - Analytics/warehouse database
- `testing` - Test database
- `staging` - Staging environment database

### 2. Connection Management

- Always release connections after use
- Use context with timeouts for database operations
- Implement proper error handling
- Use connection pooling effectively

### 3. Health Monitoring

```go
// Regular health checks
func monitorDatabases() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        if err := dbService.HealthCheck(context.Background()); err != nil {
            log.Printf("Database health check failed: %v", err)
        }
    }
}
```

### 4. Error Handling

```go
// Robust error handling
func handleDatabaseError(err error, dbName string) {
    if err != nil {
        log.Printf("Database error in %s: %v", dbName, err)
        // Implement retry logic, circuit breaker, etc.
    }
}
```

## Troubleshooting

### Common Issues

1. **Connection refused**: Check host, port, and firewall settings
2. **Authentication failed**: Verify username and password
3. **Database does not exist**: Create the database or check database name
4. **Schema not found**: Ensure the schema exists or set `isDefault: true`

### Debugging

Enable debug logging to troubleshoot connection issues:

```yaml
log:
  level: "debug"
```

### Monitoring

Use the provided health check functions to monitor database status:

```go
// Check specific database
if err := dbService.manager.HealthCheck(ctx, "primary"); err != nil {
    log.Printf("Primary database is unhealthy: %v", err)
}

// List all connections
connections := dbService.ListConnections()
log.Printf("Active databases: %v", connections)
```

## Performance Considerations

1. **Connection Pooling**: Configure appropriate pool sizes based on your workload
2. **Connection Limits**: Set reasonable limits to prevent resource exhaustion
3. **Timeout Settings**: Configure timeouts based on your application needs
4. **SSL Mode**: Use appropriate SSL settings for security

## Security Considerations

1. **Password Management**: Use environment variables for sensitive data
2. **SSL Configuration**: Enable SSL for production environments
3. **Network Security**: Use firewalls and VPNs for database access
4. **Access Control**: Implement proper database user permissions 