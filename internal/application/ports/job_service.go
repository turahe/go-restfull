package ports

import (
	"context"
	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
)

// JobService defines the interface for job management operations
type JobService interface {
	// CreateJob creates a new job in the queue
	CreateJob(ctx context.Context, queue string, handlerName string, payload interface{}, maxAttempts int, delay int) (*entities.Job, error)

	// GetJob retrieves a job by ID
	GetJob(ctx context.Context, jobID uuid.UUID) (*entities.Job, error)

	// GetJobs retrieves all jobs with optional filtering
	GetJobs(ctx context.Context) ([]*entities.Job, error)

	// GetUnfinishedJobs retrieves all unfinished jobs
	GetUnfinishedJobs(ctx context.Context) ([]*entities.Job, error)

	// GetFailedJobs retrieves all failed jobs
	GetFailedJobs(ctx context.Context) ([]*entities.FailedJob, error)

	// UpdateJobStatus updates the status of a job
	UpdateJobStatus(ctx context.Context, jobID uuid.UUID, status string) error

	// RetryFailedJob retries a failed job
	RetryFailedJob(ctx context.Context, jobID uuid.UUID) error

	// RemoveFailedJob removes a failed job
	RemoveFailedJob(ctx context.Context, jobID uuid.UUID) error

	// ResetProcessingJobs resets all processing jobs to pending status
	ResetProcessingJobs(ctx context.Context) error

	// ProcessJobs processes pending jobs
	ProcessJobs(ctx context.Context) error
}
