package services

import (
	"context"
	"encoding/json"
	"fmt"
	"webapi/internal/application/ports"
	"webapi/internal/domain/entities"
	"webapi/internal/domain/repositories"

	"github.com/google/uuid"
)

// JobServiceImpl implements JobService interface
type JobServiceImpl struct {
	jobRepository repositories.JobRepository
}

// NewJobService creates a new job service
func NewJobService(jobRepository repositories.JobRepository) ports.JobService {
	return &JobServiceImpl{
		jobRepository: jobRepository,
	}
}

// CreateJob creates a new job in the queue
func (s *JobServiceImpl) CreateJob(ctx context.Context, queue string, handlerName string, payload interface{}, maxAttempts int, delay int) (*entities.Job, error) {
	// Convert payload to JSON string
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create job entity
	job, err := entities.NewJob(queue, handlerName, string(payloadBytes), maxAttempts, delay)
	if err != nil {
		return nil, err
	}

	// Save to repository
	err = s.jobRepository.CreateJob(ctx, job)
	if err != nil {
		return nil, err
	}

	return job, nil
}

// GetJob retrieves a job by ID
func (s *JobServiceImpl) GetJob(ctx context.Context, jobID uuid.UUID) (*entities.Job, error) {
	return s.jobRepository.GetJob(ctx, jobID)
}

// GetJobs retrieves all jobs
func (s *JobServiceImpl) GetJobs(ctx context.Context) ([]*entities.Job, error) {
	return s.jobRepository.GetJobs(ctx)
}

// GetUnfinishedJobs retrieves all unfinished jobs
func (s *JobServiceImpl) GetUnfinishedJobs(ctx context.Context) ([]*entities.Job, error) {
	return s.jobRepository.GetUnfinishedJobs(ctx)
}

// UpdateJobStatus updates the status of a job
func (s *JobServiceImpl) UpdateJobStatus(ctx context.Context, jobID uuid.UUID, status string) error {
	return s.jobRepository.UpdateJobStatus(ctx, jobID, status)
}

// ResetProcessingJobs resets all processing jobs to pending status
func (s *JobServiceImpl) ResetProcessingJobs(ctx context.Context) error {
	return s.jobRepository.ResetProcessingJobs(ctx)
}

// ProcessJobs processes pending jobs
func (s *JobServiceImpl) ProcessJobs(ctx context.Context) error {
	// Get unfinished jobs
	unfinishedJobs, err := s.jobRepository.GetUnfinishedJobs(ctx)
	if err != nil {
		return err
	}

	// Process each job
	for _, job := range unfinishedJobs {
		if job.IsPending() {
			// Mark as processing
			job.MarkAsProcessing()
			err = s.jobRepository.UpdateJobStatus(ctx, job.ID, job.Status)
			if err != nil {
				return err
			}

			// TODO: Execute job handler based on HandlerName
			// For now, just mark as completed
			job.MarkAsCompleted()
			err = s.jobRepository.UpdateJobStatus(ctx, job.ID, job.Status)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// GetFailedJobs retrieves all failed jobs
func (s *JobServiceImpl) GetFailedJobs(ctx context.Context) ([]*entities.FailedJob, error) {
	failedJobs, err := s.jobRepository.GetFailedJobs(ctx)
	if err != nil {
		return nil, err
	}

	var result []*entities.FailedJob
	for _, job := range failedJobs {
		result = append(result, &job)
	}

	return result, nil
}

// RetryFailedJob retries a failed job
func (s *JobServiceImpl) RetryFailedJob(ctx context.Context, jobID uuid.UUID) error {
	failedJobs, err := s.jobRepository.GetFailedJobs(ctx)
	if err != nil {
		return err
	}

	var failedJob *entities.FailedJob
	for _, fj := range failedJobs {
		if fj.JobID == jobID {
			failedJob = &fj
			break
		}
	}

	if failedJob == nil {
		return fmt.Errorf("failed job not found")
	}

	// Create a new job from the failed job
	newJob, err := entities.NewJob(
		failedJob.Queue,
		failedJob.HandlerName,
		failedJob.Payload,
		failedJob.MaxAttempts,
		failedJob.Delay,
	)
	if err != nil {
		return err
	}

	// Save the new job
	err = s.jobRepository.CreateJob(ctx, newJob)
	if err != nil {
		return err
	}

	// Remove the failed job
	return s.jobRepository.RemoveFailedJob(ctx, jobID)
}

// RemoveFailedJob removes a failed job
func (s *JobServiceImpl) RemoveFailedJob(ctx context.Context, jobID uuid.UUID) error {
	return s.jobRepository.RemoveFailedJob(ctx, jobID)
}
