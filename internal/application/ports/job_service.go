package ports

import (
	"context"
	"webapi/internal/db/model"

	"github.com/google/uuid"
)

// JobService defines the interface for job management operations
type JobService interface {
	// CreateJob creates a new job in the queue
	CreateJob(ctx context.Context, queue string, handlerName string, payload interface{}, maxAttempts int, delay int) (*model.Job, error)
	
	// GetJob retrieves a job by ID
	GetJob(ctx context.Context, jobID uuid.UUID) (*model.Job, error)
	
	// GetJobs retrieves all jobs with optional filtering
	GetJobs(ctx context.Context) ([]*model.Job, error)
	
	// GetUnfinishedJobs retrieves all unfinished jobs
	GetUnfinishedJobs(ctx context.Context) ([]*model.Job, error)
	
	// GetFailedJobs retrieves all failed jobs
	GetFailedJobs(ctx context.Context) ([]*model.FailedJob, error)
	
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