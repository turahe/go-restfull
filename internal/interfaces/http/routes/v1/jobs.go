package v1

import (
	"webapi/internal/infrastructure/container"

	"github.com/gofiber/fiber/v2"
)

// RegisterJobRoutes registers all job-related routes
func RegisterJobRoutes(protected fiber.Router, container *container.Container) {
	jobController := container.GetJobController()

	// Job Management routes (admin only)
	jobs := protected.Group("/jobs")
	jobs.Get("/", jobController.GetJobs)
	jobs.Get("/:id", jobController.GetJob)
	jobs.Get("/failed", jobController.GetFailedJobs)
	jobs.Post("/:id/retry", jobController.RetryFailedJob)
	jobs.Delete("/failed/:id", jobController.RemoveFailedJob)
	jobs.Post("/process", jobController.ProcessJobs)
	jobs.Post("/reset", jobController.ResetProcessingJobs)
}
