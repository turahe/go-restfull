package adapters

import (
	"context"
	"webapi/internal/db/model"
	"webapi/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// JobRepository defines the interface for job operations
type JobRepository interface {
	AddJob(ctx context.Context, job model.Job) (jobID uuid.UUID, err error)
	AddFailedJob(ctx context.Context, job model.FailedJob) (failedJobID int, err error)
	UpdateJobStatus(ctx context.Context, jobID uuid.UUID, status string) error
	ResetProcessingJobsToPending(ctx context.Context) error
	GetJobs(ctx context.Context) ([]model.Job, error)
	GetUnfinishedJobs(ctx context.Context) ([]model.Job, error)
	GetFailedJobs(ctx context.Context) ([]model.FailedJob, error)
	RemoveFailedJob(ctx context.Context, jobID uuid.UUID) error
}

type PostgresJobRepository struct {
	repo repository.JobRepository
}

func NewPostgresJobRepository(db *pgxpool.Pool, redisClient redis.Cmdable) JobRepository {
	return &PostgresJobRepository{
		repo: repository.NewJobRepository(db, redisClient),
	}
}

func (r *PostgresJobRepository) AddJob(ctx context.Context, job model.Job) (jobID uuid.UUID, err error) {
	return r.repo.AddJob(ctx, job)
}

func (r *PostgresJobRepository) AddFailedJob(ctx context.Context, job model.FailedJob) (failedJobID int, err error) {
	return r.repo.AddFailedJob(ctx, job)
}

func (r *PostgresJobRepository) UpdateJobStatus(ctx context.Context, jobID uuid.UUID, status string) error {
	return r.repo.UpdateJobStatus(ctx, jobID, status)
}

func (r *PostgresJobRepository) ResetProcessingJobsToPending(ctx context.Context) error {
	return r.repo.ResetProcessingJobsToPending(ctx)
}

func (r *PostgresJobRepository) GetJobs(ctx context.Context) ([]model.Job, error) {
	return r.repo.GetJobs(ctx)
}

func (r *PostgresJobRepository) GetUnfinishedJobs(ctx context.Context) ([]model.Job, error) {
	return r.repo.GetUnfinishedJobs(ctx)
}

func (r *PostgresJobRepository) GetFailedJobs(ctx context.Context) ([]model.FailedJob, error) {
	return r.repo.GetFailedJobs(ctx)
}

func (r *PostgresJobRepository) RemoveFailedJob(ctx context.Context, jobID uuid.UUID) error {
	return r.repo.RemoveFailedJob(ctx, jobID)
}
