package repositories

import (
	"context"
	"github.com/turahe/go-restfull/internal/domain/entities"

	"github.com/google/uuid"
)

// JobRepository defines the interface for job data access
type JobRepository interface {
	// CreateJob creates a new job
	CreateJob(ctx context.Context, job *entities.Job) error

	// GetJob retrieves a job by ID
	GetJob(ctx context.Context, jobID uuid.UUID) (*entities.Job, error)

	// GetJobs retrieves all jobs
	GetJobs(ctx context.Context) ([]*entities.Job, error)

	// GetUnfinishedJobs retrieves all unfinished jobs
	GetUnfinishedJobs(ctx context.Context) ([]*entities.Job, error)

	// UpdateJobStatus updates the status of a job
	UpdateJobStatus(ctx context.Context, jobID uuid.UUID, status string) error

	// AddFailedJob adds a failed job to the failed jobs table
	AddFailedJob(ctx context.Context, job entities.FailedJob) (failedJobID int, err error)

	// GetFailedJobs retrieves all failed jobs
	GetFailedJobs(ctx context.Context) ([]entities.FailedJob, error)

	// RemoveFailedJob removes a failed job
	RemoveFailedJob(ctx context.Context, jobID uuid.UUID) error

	// ResetProcessingJobs resets all processing jobs to pending status
	ResetProcessingJobs(ctx context.Context) error
}
