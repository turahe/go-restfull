// Package main provides the main entry point for the Go RESTful API
//	@title			Go RESTful API - Hexagonal Architecture
//	@version		1.0
//	@description	A comprehensive RESTful API built with Go, Fiber, and PostgreSQL using Hexagonal Architecture
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		localhost:8000
//	@BasePath	/api/v1

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.

package main

import (
	"context"
	"log"
	"os"

	"github.com/turahe/go-restfull/pkg/logger"

	"github.com/turahe/go-restfull/cmd"
	_ "github.com/turahe/go-restfull/docs/api" // Import for Swagger documentation
	"github.com/turahe/go-restfull/internal/db/pgx"
	"github.com/turahe/go-restfull/internal/db/seeds"
	"github.com/turahe/go-restfull/internal/infrastructure/container"
	"github.com/turahe/go-restfull/internal/router"
	"go.uber.org/automaxprocs/maxprocs"
)

// Initialize Swagger documentation
func init() {
	// This ensures the docs package is imported and its init function runs
	// SwaggerInfo is now imported from docs/api package
}

func main() {
	// Check if command line arguments are provided
	if len(os.Args) > 1 {
		// Execute command (like migrate, etc.)
		cmd.Execute()
		return
	}

	// Default behavior: start the server
	defer func() {
		_ = os.Remove("./tmp/live")

		// Close the database connection pool
		pgx.ClosePgxPool()

		// Flush the log buffer
		if logger.Log != nil {
			logger.Log.Sync()
		}
	}()

	// Liveness probe for Kubernetes
	_, err := os.Create("./tmp/live")
	if err != nil {
		log.Fatalf("Cannot create a Liveness file: %v", err)
	}

	// Set the maximum number of CPUs to use
	nopLog := func(string, ...interface{}) {}
	_, err = maxprocs.Set(maxprocs.Logger(nopLog))
	if err != nil {
		log.Fatalf("Cannot set maxprocs: %v", err)
	}

	// Setup all dependencies
	cmd.SetupAll()

	// Initialize the dependency injection container
	db := pgx.GetPgxPool()
	container := container.NewContainer(db)

	// Ensure default roles exist
	if err := ensureDefaultRoles(); err != nil {
		log.Printf("Warning: Failed to ensure default roles: %v", err)
	}

	// Start messaging consumers
	err = container.MessagingService.StartConsumers(context.Background())
	if err != nil {
		log.Fatalf("Failed to start messaging consumers: %v", err)
	}

	// Use the proper router setup from the router package
	app := router.NewFiberRouter()

	// Start the server
	log.Fatal(app.Listen(":8000"))
}

// ensureDefaultRoles ensures that default roles exist in the database
func ensureDefaultRoles() error {
	return seeds.EnsureDefaultUserRole()
}
