package repository

import (
	"context"
	"webapi/internal/domain/entities"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type JobRepository interface {
	Create(ctx context.Context, job *entities.Job) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Job, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entities.Job, error)
	GetByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.Job, error)
	Update(ctx context.Context, job *entities.Job) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
	CountByStatus(ctx context.Context, status string) (int64, error)
	// Additional methods for job management
	GetUnfinished(ctx context.Context) ([]*entities.Job, error)
	UpdateStatus(ctx context.Context, jobID uuid.UUID, status string) error
	AddFailedJob(ctx context.Context, job entities.FailedJob) (failedJobID int, err error)
	GetFailedJobs(ctx context.Context) ([]entities.FailedJob, error)
	RemoveFailedJob(ctx context.Context, jobID uuid.UUID) error
	ResetProcessing(ctx context.Context) error
}

type JobRepositoryImpl struct {
	pgxPool     *pgxpool.Pool
	redisClient redis.Cmdable
}

func NewJobRepository(pgxPool *pgxpool.Pool, redisClient redis.Cmdable) JobRepository {
	return &JobRepositoryImpl{
		pgxPool:     pgxPool,
		redisClient: redisClient,
	}
}

func (r *JobRepositoryImpl) Create(ctx context.Context, job *entities.Job) error {
	query := `INSERT INTO jobs (id, queue, handler_name, payload, max_attempts, delay, status, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.pgxPool.Exec(ctx, query,
		job.ID.String(), job.Queue, job.HandlerName, job.Payload, job.MaxAttempts, job.Delay, job.Status,
		job.CreatedAt, job.UpdatedAt)
	return err
}

func (r *JobRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Job, error) {
	query := `SELECT id, queue, handler_name, payload, max_attempts, delay, status, created_at, updated_at
			  FROM jobs WHERE id = $1`

	var job entities.Job
	err := r.pgxPool.QueryRow(ctx, query, id.String()).Scan(
		&job.ID, &job.Queue, &job.HandlerName, &job.Payload, &job.MaxAttempts, &job.Delay, &job.Status,
		&job.CreatedAt, &job.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &job, nil
}

func (r *JobRepositoryImpl) GetAll(ctx context.Context, limit, offset int) ([]*entities.Job, error) {
	query := `SELECT id, queue, handler_name, payload, max_attempts, delay, status, created_at, updated_at
			  FROM jobs
			  ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	rows, err := r.pgxPool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*entities.Job
	for rows.Next() {
		job, err := r.scanJobRow(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (r *JobRepositoryImpl) GetByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.Job, error) {
	query := `SELECT id, queue, handler_name, payload, max_attempts, delay, status, created_at, updated_at
			  FROM jobs WHERE status = $1
			  ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.pgxPool.Query(ctx, query, status, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*entities.Job
	for rows.Next() {
		job, err := r.scanJobRow(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (r *JobRepositoryImpl) Update(ctx context.Context, job *entities.Job) error {
	query := `UPDATE jobs SET queue = $1, handler_name = $2, payload = $3, max_attempts = $4, delay = $5, status = $6, updated_at = $7
			  WHERE id = $8`

	_, err := r.pgxPool.Exec(ctx, query, job.Queue, job.HandlerName, job.Payload, job.MaxAttempts, job.Delay, job.Status,
		job.UpdatedAt, job.ID.String())
	return err
}

func (r *JobRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM jobs WHERE id = $1`
	_, err := r.pgxPool.Exec(ctx, query, id.String())
	return err
}

func (r *JobRepositoryImpl) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM jobs`
	var count int64
	err := r.pgxPool.QueryRow(ctx, query).Scan(&count)
	return count, err
}

func (r *JobRepositoryImpl) CountByStatus(ctx context.Context, status string) (int64, error) {
	query := `SELECT COUNT(*) FROM jobs WHERE status = $1`
	var count int64
	err := r.pgxPool.QueryRow(ctx, query, status).Scan(&count)
	return count, err
}

func (r *JobRepositoryImpl) GetUnfinished(ctx context.Context) ([]*entities.Job, error) {
	query := `SELECT id, queue, handler_name, payload, max_attempts, delay, status, created_at, updated_at
			  FROM jobs WHERE status IN ('pending', 'processing')
			  ORDER BY created_at ASC`

	rows, err := r.pgxPool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*entities.Job
	for rows.Next() {
		job, err := r.scanJobRow(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (r *JobRepositoryImpl) UpdateStatus(ctx context.Context, jobID uuid.UUID, status string) error {
	query := `UPDATE jobs SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.pgxPool.Exec(ctx, query, status, jobID.String())
	return err
}

func (r *JobRepositoryImpl) AddFailedJob(ctx context.Context, job entities.FailedJob) (failedJobID int, err error) {
	query := `INSERT INTO failed_jobs (job_id, queue, handler_name, payload, max_attempts, delay, status, error, failed_at, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id`

	var id int
	err = r.pgxPool.QueryRow(ctx, query,
		job.JobID.String(), job.Queue, job.HandlerName, job.Payload, job.MaxAttempts, job.Delay, job.Status, job.Error,
		job.FailedAt, job.CreatedAt, job.UpdatedAt).Scan(&id)
	return id, err
}

func (r *JobRepositoryImpl) GetFailedJobs(ctx context.Context) ([]entities.FailedJob, error) {
	query := `SELECT id, job_id, queue, handler_name, payload, max_attempts, delay, status, error, failed_at, created_at, updated_at
			  FROM failed_jobs ORDER BY failed_at DESC`

	rows, err := r.pgxPool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var failedJobs []entities.FailedJob
	for rows.Next() {
		var failedJob entities.FailedJob
		err := rows.Scan(
			&failedJob.ID, &failedJob.JobID, &failedJob.Queue, &failedJob.HandlerName, &failedJob.Payload,
			&failedJob.MaxAttempts, &failedJob.Delay, &failedJob.Status, &failedJob.Error, &failedJob.FailedAt,
			&failedJob.CreatedAt, &failedJob.UpdatedAt)
		if err != nil {
			return nil, err
		}
		failedJobs = append(failedJobs, failedJob)
	}

	return failedJobs, nil
}

func (r *JobRepositoryImpl) RemoveFailedJob(ctx context.Context, jobID uuid.UUID) error {
	query := `DELETE FROM failed_jobs WHERE job_id = $1`
	_, err := r.pgxPool.Exec(ctx, query, jobID.String())
	return err
}

func (r *JobRepositoryImpl) ResetProcessing(ctx context.Context) error {
	query := `UPDATE jobs SET status = 'pending', updated_at = NOW() WHERE status = 'processing'`
	_, err := r.pgxPool.Exec(ctx, query)
	return err
}

// scanJobRow is a helper function to scan a job row from database
func (r *JobRepositoryImpl) scanJobRow(rows pgx.Rows) (*entities.Job, error) {
	var job entities.Job
	err := rows.Scan(
		&job.ID, &job.Queue, &job.HandlerName, &job.Payload, &job.MaxAttempts, &job.Delay, &job.Status,
		&job.CreatedAt, &job.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &job, nil
}
