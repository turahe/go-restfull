package repositories

import (
	"context"
	"webapi/internal/db/model"

	"github.com/google/uuid"
)

// JobRepository defines the interface for job data access
type JobRepository interface {
	// AddJob adds a new job to the queue
	AddJob(ctx context.Context, job model.Job) (jobID uuid.UUID, err error)

	// AddFailedJob adds a failed job to the failed jobs table
	AddFailedJob(ctx context.Context, job model.FailedJob) (failedJobID int, err error)

	// UpdateJobStatus updates the status of a job
	UpdateJobStatus(ctx context.Context, jobID uuid.UUID, status string) error

	// ResetProcessingJobsToPending resets all processing jobs to pending status
	ResetProcessingJobsToPending(ctx context.Context) error

	// GetJobs retrieves all jobs
	GetJobs(ctx context.Context) ([]model.Job, error)

	// GetUnfinishedJobs retrieves all unfinished jobs
	GetUnfinishedJobs(ctx context.Context) ([]model.Job, error)

	// GetFailedJobs retrieves all failed jobs
	GetFailedJobs(ctx context.Context) ([]model.FailedJob, error)

	// RemoveFailedJob removes a failed job
	RemoveFailedJob(ctx context.Context, jobID uuid.UUID) error
}
