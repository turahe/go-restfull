// Package main provides the main entry point for the Go RESTful API
// @title Go RESTful API - Hexagonal Architecture
// @version 1.0
// @description A comprehensive RESTful API built with Go, Fiber, and PostgreSQL using Hexagonal Architecture
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
	"webapi/internal/infrastructure/container"
	"webapi/internal/interfaces/http/routes"
	"webapi/internal/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
	"go.uber.org/automaxprocs/maxprocs"
)

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

	// Setup Fiber app with enhanced configuration
	app := fiber.New(fiber.Config{
		ErrorHandler:          customErrorHandler,
		DisableStartupMessage: true,
		EnablePrintRoutes:     false,
	})

	// Middleware
	app.Use(cors.New())
	app.Use(fiberlogger.New())

	// Swagger documentation
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Setup routes using Hexagonal Architecture
	routes.RegisterHexagonalRoutes(app, container)

	// Start the server
	log.Fatal(app.Listen(":8000"))
}

// customErrorHandler handles application errors
func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	return c.Status(code).JSON(fiber.Map{
		"status":  "error",
		"message": err.Error(),
	})
}
