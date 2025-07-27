package services

import (
	"context"
	"time"

	"webapi/internal/application/ports"
	"webapi/internal/db/model"
	"webapi/internal/domain/repositories"
	"webapi/internal/job"

	"github.com/google/uuid"
)

type JobServiceImpl struct {
	jobRepository repositories.JobRepository
}

func NewJobService(jobRepository repositories.JobRepository) ports.JobService {
	return &JobServiceImpl{
		jobRepository: jobRepository,
	}
}

func (s *JobServiceImpl) CreateJob(ctx context.Context, queue string, handlerName string, payload interface{}, maxAttempts int, delay int) (*model.Job, error) {
	jobInstance, err := job.NewJob(handlerName, payload, maxAttempts, delay)
	if err != nil {
		return nil, err
	}

	modelJob := &model.Job{
		ID:          jobInstance.ID,
		Queue:       queue,
		HandlerName: jobInstance.HandlerName,
		Payload:     jobInstance.Payload,
		MaxAttempts: jobInstance.MaxAttempts,
		Delay:       jobInstance.Delay,
		Status:      "pending",
		CreatedAt:   jobInstance.CreatedAt,
		UpdatedAt:   jobInstance.CreatedAt,
	}

	_, err = s.jobRepository.AddJob(ctx, *modelJob)
	return modelJob, err
}

func (s *JobServiceImpl) GetJob(ctx context.Context, jobID uuid.UUID) (*model.Job, error) {
	jobs, err := s.jobRepository.GetJobs(ctx)
	if err != nil {
		return nil, err
	}

	for _, job := range jobs {
		if job.ID == jobID {
			return &job, nil
		}
	}

	return nil, nil
}

func (s *JobServiceImpl) GetJobs(ctx context.Context) ([]*model.Job, error) {
	jobs, err := s.jobRepository.GetJobs(ctx)
	if err != nil {
		return nil, err
	}

	var result []*model.Job
	for _, job := range jobs {
		result = append(result, &job)
	}

	return result, nil
}

func (s *JobServiceImpl) GetUnfinishedJobs(ctx context.Context) ([]*model.Job, error) {
	jobs, err := s.jobRepository.GetUnfinishedJobs(ctx)
	if err != nil {
		return nil, err
	}

	var result []*model.Job
	for _, job := range jobs {
		result = append(result, &job)
	}

	return result, nil
}

func (s *JobServiceImpl) GetFailedJobs(ctx context.Context) ([]*model.FailedJob, error) {
	failedJobs, err := s.jobRepository.GetFailedJobs(ctx)
	if err != nil {
		return nil, err
	}

	var result []*model.FailedJob
	for _, job := range failedJobs {
		result = append(result, &job)
	}

	return result, nil
}

func (s *JobServiceImpl) UpdateJobStatus(ctx context.Context, jobID uuid.UUID, status string) error {
	return s.jobRepository.UpdateJobStatus(ctx, jobID, status)
}

func (s *JobServiceImpl) RetryFailedJob(ctx context.Context, jobID uuid.UUID) error {
	failedJobs, err := s.jobRepository.GetFailedJobs(ctx)
	if err != nil {
		return err
	}

	var failedJob *model.FailedJob
	for _, fj := range failedJobs {
		if fj.JobID == jobID {
			failedJob = &fj
			break
		}
	}

	if failedJob == nil {
		return nil
	}

	job := &model.Job{
		ID:          uuid.New(),
		Queue:       failedJob.Queue,
		HandlerName: "retry",
		Payload:     failedJob.Payload,
		MaxAttempts: 3,
		Delay:       0,
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	_, err = s.jobRepository.AddJob(ctx, *job)
	if err != nil {
		return err
	}

	return s.jobRepository.RemoveFailedJob(ctx, jobID)
}

func (s *JobServiceImpl) RemoveFailedJob(ctx context.Context, jobID uuid.UUID) error {
	return s.jobRepository.RemoveFailedJob(ctx, jobID)
}

func (s *JobServiceImpl) ResetProcessingJobs(ctx context.Context) error {
	return s.jobRepository.ResetProcessingJobsToPending(ctx)
}

func (s *JobServiceImpl) ProcessJobs(ctx context.Context) error {
	jobs, err := s.GetUnfinishedJobs(ctx)
	if err != nil {
		return err
	}

	for _, job := range jobs {
		if job.Status == "pending" {
			s.UpdateJobStatus(ctx, job.ID, "processing")
			// Process job logic here
			s.UpdateJobStatus(ctx, job.ID, "completed")
		}
	}

	return nil
}
