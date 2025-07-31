// Package services provides application-level business logic for job queue management.
// This package contains the job service implementation that handles job creation,
// processing, status management, and failed job handling while ensuring reliable
// background task execution.
package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/turahe/go-restfull/internal/application/ports"
	"github.com/turahe/go-restfull/internal/domain/entities"
	"github.com/turahe/go-restfull/internal/domain/repositories"

	"github.com/google/uuid"
)

// JobServiceImpl implements the JobService interface and provides comprehensive
// job queue management functionality. It handles job creation, processing, status
// tracking, failed job management, and retry mechanisms while ensuring reliable
// background task execution.
type JobServiceImpl struct {
	jobRepository repositories.JobRepository
}

// NewJobService creates a new job service instance with the provided repository.
// This function follows the dependency injection pattern to ensure loose coupling
// between the service layer and the data access layer.
//
// Parameters:
//   - jobRepository: Repository interface for job data access operations
//
// Returns:
//   - ports.JobService: The job service interface implementation
func NewJobService(jobRepository repositories.JobRepository) ports.JobService {
	return &JobServiceImpl{
		jobRepository: jobRepository,
	}
}

// CreateJob creates a new job in the queue with specified parameters.
// This method enforces business rules for job creation and supports
// various job types with configurable retry mechanisms.
//
// Business Rules:
//   - Queue name must be specified for job routing
//   - Handler name must be provided for job execution
//   - Payload must be serializable to JSON
//   - Max attempts must be positive for retry logic
//   - Delay must be non-negative for scheduling
//
// Parameters:
//   - ctx: Context for the operation
//   - queue: Queue name for job routing and processing
//   - handlerName: Name of the handler to execute the job
//   - payload: Job data to be processed (will be serialized to JSON)
//   - maxAttempts: Maximum number of retry attempts for failed jobs
//   - delay: Delay in seconds before job execution
//
// Returns:
//   - *entities.Job: The created job entity
//   - error: Any error that occurred during the operation
func (s *JobServiceImpl) CreateJob(ctx context.Context, queue string, handlerName string, payload interface{}, maxAttempts int, delay int) (*entities.Job, error) {
	// Convert payload to JSON string for storage and processing
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create job entity with the provided parameters
	job, err := entities.NewJob(queue, handlerName, string(payloadBytes), maxAttempts, delay)
	if err != nil {
		return nil, err
	}

	// Persist the job to the repository
	err = s.jobRepository.CreateJob(ctx, job)
	if err != nil {
		return nil, err
	}

	return job, nil
}

// GetJob retrieves a job by its unique identifier.
// This method provides access to individual job details and status information.
//
// Parameters:
//   - ctx: Context for the operation
//   - jobID: UUID of the job to retrieve
//
// Returns:
//   - *entities.Job: The job entity if found
//   - error: Error if job not found or other issues occur
func (s *JobServiceImpl) GetJob(ctx context.Context, jobID uuid.UUID) (*entities.Job, error) {
	return s.jobRepository.GetJob(ctx, jobID)
}

// GetJobs retrieves all jobs in the system.
// This method is useful for administrative purposes and job monitoring.
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns:
//   - []*entities.Job: List of all jobs
//   - error: Any error that occurred during the operation
func (s *JobServiceImpl) GetJobs(ctx context.Context) ([]*entities.Job, error) {
	return s.jobRepository.GetJobs(ctx)
}

// GetUnfinishedJobs retrieves all jobs that are not yet completed.
// This method is useful for job processing and monitoring active jobs.
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns:
//   - []*entities.Job: List of unfinished jobs (pending, processing, failed)
//   - error: Any error that occurred during the operation
func (s *JobServiceImpl) GetUnfinishedJobs(ctx context.Context) ([]*entities.Job, error) {
	return s.jobRepository.GetUnfinishedJobs(ctx)
}

// UpdateJobStatus updates the status of a specific job.
// This method is used during job processing to track job lifecycle.
//
// Parameters:
//   - ctx: Context for the operation
//   - jobID: UUID of the job to update
//   - status: New status for the job
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *JobServiceImpl) UpdateJobStatus(ctx context.Context, jobID uuid.UUID, status string) error {
	return s.jobRepository.UpdateJobStatus(ctx, jobID, status)
}

// ResetProcessingJobs resets all jobs with "processing" status back to "pending".
// This method is useful for recovery scenarios where jobs may have been interrupted.
//
// Business Rules:
//   - Only jobs with "processing" status are affected
//   - Jobs are reset to "pending" status for reprocessing
//   - This operation is useful for system recovery
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *JobServiceImpl) ResetProcessingJobs(ctx context.Context) error {
	return s.jobRepository.ResetProcessingJobs(ctx)
}

// ProcessJobs processes all pending jobs in the system.
// This method implements the job processing logic and handles job lifecycle management.
//
// Business Rules:
//   - Only pending jobs are processed
//   - Jobs are marked as processing during execution
//   - Completed jobs are marked as finished
//   - Failed jobs are moved to failed jobs table
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *JobServiceImpl) ProcessJobs(ctx context.Context) error {
	// Get all unfinished jobs for processing
	unfinishedJobs, err := s.jobRepository.GetUnfinishedJobs(ctx)
	if err != nil {
		return err
	}

	// Process each pending job
	for _, job := range unfinishedJobs {
		if job.IsPending() {
			// Mark job as processing to prevent duplicate execution
			job.MarkAsProcessing()
			err = s.jobRepository.UpdateJobStatus(ctx, job.ID, job.Status)
			if err != nil {
				return err
			}

			// TODO: Execute job handler based on HandlerName
			// For now, just mark as completed
			// In a real implementation, you would:
			// 1. Deserialize the payload
			// 2. Find the appropriate handler
			// 3. Execute the handler with the payload
			// 4. Handle success/failure appropriately
			job.MarkAsCompleted()
			err = s.jobRepository.UpdateJobStatus(ctx, job.ID, job.Status)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// GetFailedJobs retrieves all failed jobs from the failed jobs table.
// This method provides access to jobs that have exceeded their retry attempts.
//
// Parameters:
//   - ctx: Context for the operation
//
// Returns:
//   - []*entities.FailedJob: List of failed jobs
//   - error: Any error that occurred during the operation
func (s *JobServiceImpl) GetFailedJobs(ctx context.Context) ([]*entities.FailedJob, error) {
	failedJobs, err := s.jobRepository.GetFailedJobs(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to slice of pointers for consistency
	var result []*entities.FailedJob
	for _, job := range failedJobs {
		result = append(result, &job)
	}

	return result, nil
}

// RetryFailedJob retries a specific failed job by creating a new job with the same parameters.
// This method allows manual retry of failed jobs that have exceeded their automatic retry attempts.
//
// Business Rules:
//   - Failed job must exist in the failed jobs table
//   - New job is created with the same parameters as the failed job
//   - Original failed job is removed from the failed jobs table
//   - New job starts with fresh retry attempts
//
// Parameters:
//   - ctx: Context for the operation
//   - jobID: UUID of the failed job to retry
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *JobServiceImpl) RetryFailedJob(ctx context.Context, jobID uuid.UUID) error {
	// Get all failed jobs to find the specific one
	failedJobs, err := s.jobRepository.GetFailedJobs(ctx)
	if err != nil {
		return err
	}

	// Find the specific failed job
	var failedJob *entities.FailedJob
	for _, fj := range failedJobs {
		if fj.JobID == jobID {
			failedJob = &fj
			break
		}
	}

	// Check if failed job exists
	if failedJob == nil {
		return fmt.Errorf("failed job not found")
	}

	// Create a new job from the failed job parameters
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

	// Save the new job to the repository
	err = s.jobRepository.CreateJob(ctx, newJob)
	if err != nil {
		return err
	}

	// Remove the original failed job from the failed jobs table
	return s.jobRepository.RemoveFailedJob(ctx, jobID)
}

// RemoveFailedJob permanently removes a failed job from the failed jobs table.
// This method is useful for cleaning up failed jobs that are no longer needed.
//
// Parameters:
//   - ctx: Context for the operation
//   - jobID: UUID of the failed job to remove
//
// Returns:
//   - error: Any error that occurred during the operation
func (s *JobServiceImpl) RemoveFailedJob(ctx context.Context, jobID uuid.UUID) error {
	return s.jobRepository.RemoveFailedJob(ctx, jobID)
}
