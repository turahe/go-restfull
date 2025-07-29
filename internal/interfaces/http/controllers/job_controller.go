package controllers

import (
	"webapi/internal/application/ports"
	"webapi/internal/interfaces/http/responses"
	"webapi/internal/router/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type JobController struct {
	jobService ports.JobService
}

func NewJobController(jobService ports.JobService) *JobController {
	return &JobController{
		jobService: jobService,
	}
}

// GetJobs retrieves all jobs
func (c *JobController) GetJobs(ctx *fiber.Ctx) error {
	// Get pagination parameters from middleware
	pagination := middleware.GetPaginationParams(ctx)

	jobs, err := c.jobService.GetJobs(ctx.Context())
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    int(responses.INTERNAL_SERVER_ERROR),
			ResponseMessage: "Failed to retrieve jobs",
		})
	}

	// For now, use simple count. In real implementation, get total count
	total := int64(len(jobs))

	// Create paginated response using helper
	paginatedResult := responses.CreatePaginatedResult(jobs, pagination.Page, pagination.PerPage, total)

	return ctx.JSON(responses.CommonResponse{
		ResponseCode:    int(responses.SYSTEM_OPERATION_SUCCESS),
		ResponseMessage: "Jobs retrieved successfully",
		Data:            paginatedResult,
	})
}

// GetJob retrieves a specific job by ID
func (c *JobController) GetJob(ctx *fiber.Ctx) error {
	jobID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    int(responses.BAD_REQUEST),
			ResponseMessage: "Invalid job ID",
		})
	}

	job, err := c.jobService.GetJob(ctx.Context(), jobID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    int(responses.INTERNAL_SERVER_ERROR),
			ResponseMessage: "Failed to retrieve job",
		})
	}

	if job == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(responses.CommonResponse{
			ResponseCode:    int(responses.NOT_FOUND),
			ResponseMessage: "Job not found",
		})
	}

	return ctx.JSON(responses.CommonResponse{
		ResponseCode:    int(responses.SYSTEM_OPERATION_SUCCESS),
		ResponseMessage: "Job retrieved successfully",
		Data:            job,
	})
}

// GetFailedJobs retrieves all failed jobs
func (c *JobController) GetFailedJobs(ctx *fiber.Ctx) error {
	// Get pagination parameters from middleware
	pagination := middleware.GetPaginationParams(ctx)

	failedJobs, err := c.jobService.GetFailedJobs(ctx.Context())
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    int(responses.INTERNAL_SERVER_ERROR),
			ResponseMessage: "Failed to retrieve failed jobs",
		})
	}

	// For now, use simple count. In real implementation, get total count
	total := int64(len(failedJobs))

	// Create paginated response using helper
	paginatedResult := responses.CreatePaginatedResult(failedJobs, pagination.Page, pagination.PerPage, total)

	return ctx.JSON(responses.CommonResponse{
		ResponseCode:    int(responses.SYSTEM_OPERATION_SUCCESS),
		ResponseMessage: "Failed jobs retrieved successfully",
		Data:            paginatedResult,
	})
}

// RetryFailedJob retries a failed job
func (c *JobController) RetryFailedJob(ctx *fiber.Ctx) error {
	jobID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    int(responses.BAD_REQUEST),
			ResponseMessage: "Invalid job ID",
		})
	}

	err = c.jobService.RetryFailedJob(ctx.Context(), jobID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    int(responses.INTERNAL_SERVER_ERROR),
			ResponseMessage: "Failed to retry job",
		})
	}

	return ctx.JSON(responses.CommonResponse{
		ResponseCode:    int(responses.SYSTEM_OPERATION_SUCCESS),
		ResponseMessage: "Job retried successfully",
	})
}

// RemoveFailedJob removes a failed job
func (c *JobController) RemoveFailedJob(ctx *fiber.Ctx) error {
	jobID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(responses.CommonResponse{
			ResponseCode:    int(responses.BAD_REQUEST),
			ResponseMessage: "Invalid job ID",
		})
	}

	err = c.jobService.RemoveFailedJob(ctx.Context(), jobID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    int(responses.INTERNAL_SERVER_ERROR),
			ResponseMessage: "Failed to remove failed job",
		})
	}

	return ctx.JSON(responses.CommonResponse{
		ResponseCode:    int(responses.SYSTEM_OPERATION_SUCCESS),
		ResponseMessage: "Failed job removed successfully",
	})
}

// ProcessJobs processes pending jobs
func (c *JobController) ProcessJobs(ctx *fiber.Ctx) error {
	err := c.jobService.ProcessJobs(ctx.Context())
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    int(responses.INTERNAL_SERVER_ERROR),
			ResponseMessage: "Failed to process jobs",
		})
	}

	return ctx.JSON(responses.CommonResponse{
		ResponseCode:    int(responses.SYSTEM_OPERATION_SUCCESS),
		ResponseMessage: "Jobs processed successfully",
	})
}

// ResetProcessingJobs resets all processing jobs to pending
func (c *JobController) ResetProcessingJobs(ctx *fiber.Ctx) error {
	err := c.jobService.ResetProcessingJobs(ctx.Context())
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(responses.CommonResponse{
			ResponseCode:    int(responses.INTERNAL_SERVER_ERROR),
			ResponseMessage: "Failed to reset processing jobs",
		})
	}

	return ctx.JSON(responses.CommonResponse{
		ResponseCode:    int(responses.SYSTEM_OPERATION_SUCCESS),
		ResponseMessage: "Processing jobs reset successfully",
	})
}
