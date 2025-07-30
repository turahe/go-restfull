package adapters

import (
	"context"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"
	"github.com/turahe/go-restfull/internal/repository"

	"github.com/google/uuid"
)

// PostgresJobRepository implements JobRepository interface
type PostgresJobRepository struct {
	repo repository.JobRepository
}

// NewPostgresJobRepository creates a new postgres job repository
func NewPostgresJobRepository(repo repository.JobRepository) repositories.JobRepository {
	return &PostgresJobRepository{
		repo: repo,
	}
}

// CreateJob creates a new job
func (r *PostgresJobRepository) CreateJob(ctx context.Context, job *entities.Job) error {
	return r.repo.Create(ctx, job)
}

// GetJob retrieves a job by ID
func (r *PostgresJobRepository) GetJob(ctx context.Context, jobID uuid.UUID) (*entities.Job, error) {
	return r.repo.GetByID(ctx, jobID)
}

// GetJobs retrieves all jobs
func (r *PostgresJobRepository) GetJobs(ctx context.Context) ([]*entities.Job, error) {
	return r.repo.GetAll(ctx, 100, 0) // Default limit and offset
}

// GetUnfinishedJobs retrieves all unfinished jobs
func (r *PostgresJobRepository) GetUnfinishedJobs(ctx context.Context) ([]*entities.Job, error) {
	return r.repo.GetUnfinished(ctx)
}

// UpdateJobStatus updates the status of a job
func (r *PostgresJobRepository) UpdateJobStatus(ctx context.Context, jobID uuid.UUID, status string) error {
	return r.repo.UpdateStatus(ctx, jobID, status)
}

// AddFailedJob adds a failed job to the failed jobs table
func (r *PostgresJobRepository) AddFailedJob(ctx context.Context, job entities.FailedJob) (failedJobID int, err error) {
	return r.repo.AddFailedJob(ctx, job)
}

// GetFailedJobs retrieves all failed jobs
func (r *PostgresJobRepository) GetFailedJobs(ctx context.Context) ([]entities.FailedJob, error) {
	return r.repo.GetFailedJobs(ctx)
}

// RemoveFailedJob removes a failed job
func (r *PostgresJobRepository) RemoveFailedJob(ctx context.Context, jobID uuid.UUID) error {
	return r.repo.RemoveFailedJob(ctx, jobID)
}

// ResetProcessingJobs resets all processing jobs to pending status
func (r *PostgresJobRepository) ResetProcessingJobs(ctx context.Context) error {
	return r.repo.ResetProcessing(ctx)
}
