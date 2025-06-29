// Package main provides the main entry point for the Go RESTful API
// @title Go RESTful API
// @version 1.0
// @description A comprehensive RESTful API built with Go, Fiber, and PostgreSQL
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8000
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

package main

import (
	"log"
	"os"

	"webapi/cmd"
	"webapi/internal/db/pgx"
	"webapi/internal/logger"

	"go.uber.org/automaxprocs/maxprocs"
)

// @title Go RESTful API
// @version 1.0
// @description A comprehensive RESTful API built with Go, Fiber, and PostgreSQL
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8000
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
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

	// Start the app here via CLI commands
	cmd.Execute()
}
