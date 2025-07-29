package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Job represents a job entity in the domain layer
type Job struct {
	ID          uuid.UUID `json:"id"`
	Queue       string    `json:"queue"`
	HandlerName string    `json:"handler_name"`
	Payload     string    `json:"payload"`
	MaxAttempts int       `json:"max_attempts"`
	Delay       int       `json:"delay"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewJob creates a new job instance
func NewJob(queue, handlerName, payload string, maxAttempts, delay int) (*Job, error) {
	if queue == "" {
		return nil, errors.New("queue is required")
	}
	if handlerName == "" {
		return nil, errors.New("handler_name is required")
	}
	if payload == "" {
		return nil, errors.New("payload is required")
	}
	if maxAttempts <= 0 {
		return nil, errors.New("max_attempts must be greater than 0")
	}

	now := time.Now()
	return &Job{
		ID:          uuid.New(),
		Queue:       queue,
		HandlerName: handlerName,
		Payload:     payload,
		MaxAttempts: maxAttempts,
		Delay:       delay,
		Status:      "pending",
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// UpdateJob updates job information
func (j *Job) UpdateJob(queue, handlerName, payload, status string, maxAttempts, delay int) error {
	if queue != "" {
		j.Queue = queue
	}
	if handlerName != "" {
		j.HandlerName = handlerName
	}
	if payload != "" {
		j.Payload = payload
	}
	if status != "" {
		j.Status = status
	}
	if maxAttempts > 0 {
		j.MaxAttempts = maxAttempts
	}
	if delay >= 0 {
		j.Delay = delay
	}
	j.UpdatedAt = time.Now()
	return nil
}

// MarkAsProcessing marks the job as processing
func (j *Job) MarkAsProcessing() {
	j.Status = "processing"
	j.UpdatedAt = time.Now()
}

// MarkAsCompleted marks the job as completed
func (j *Job) MarkAsCompleted() {
	j.Status = "completed"
	j.UpdatedAt = time.Now()
}

// MarkAsFailed marks the job as failed
func (j *Job) MarkAsFailed() {
	j.Status = "failed"
	j.UpdatedAt = time.Now()
}

// IsPending checks if the job is pending
func (j *Job) IsPending() bool {
	return j.Status == "pending"
}

// IsProcessing checks if the job is processing
func (j *Job) IsProcessing() bool {
	return j.Status == "processing"
}

// IsCompleted checks if the job is completed
func (j *Job) IsCompleted() bool {
	return j.Status == "completed"
}

// IsFailed checks if the job is failed
func (j *Job) IsFailed() bool {
	return j.Status == "failed"
}

// Validate validates the job
func (j *Job) Validate() error {
	if j.Queue == "" {
		return errors.New("queue is required")
	}
	if j.HandlerName == "" {
		return errors.New("handler_name is required")
	}
	if j.Payload == "" {
		return errors.New("payload is required")
	}
	if j.MaxAttempts <= 0 {
		return errors.New("max_attempts must be greater than 0")
	}
	return nil
}

// FailedJob represents a failed job entity in the domain layer
type FailedJob struct {
	ID          int       `json:"id"`
	JobID       uuid.UUID `json:"job_id"`
	Queue       string    `json:"queue"`
	HandlerName string    `json:"handler_name"`
	Payload     string    `json:"payload"`
	MaxAttempts int       `json:"max_attempts"`
	Delay       int       `json:"delay"`
	Status      string    `json:"status"`
	Error       string    `json:"error"`
	FailedAt    time.Time `json:"failed_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewFailedJob creates a new failed job instance
func NewFailedJob(jobID uuid.UUID, queue, handlerName, payload, status, errorMsg string, maxAttempts, delay int) (*FailedJob, error) {
	if jobID == uuid.Nil {
		return nil, errors.New("job_id is required")
	}
	if queue == "" {
		return nil, errors.New("queue is required")
	}
	if handlerName == "" {
		return nil, errors.New("handler_name is required")
	}
	if payload == "" {
		return nil, errors.New("payload is required")
	}
	if maxAttempts <= 0 {
		return nil, errors.New("max_attempts must be greater than 0")
	}

	now := time.Now()
	return &FailedJob{
		JobID:       jobID,
		Queue:       queue,
		HandlerName: handlerName,
		Payload:     payload,
		MaxAttempts: maxAttempts,
		Delay:       delay,
		Status:      status,
		Error:       errorMsg,
		FailedAt:    now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// UpdateFailedJob updates failed job information
func (fj *FailedJob) UpdateFailedJob(queue, handlerName, payload, status, errorMsg string, maxAttempts, delay int) error {
	if queue != "" {
		fj.Queue = queue
	}
	if handlerName != "" {
		fj.HandlerName = handlerName
	}
	if payload != "" {
		fj.Payload = payload
	}
	if status != "" {
		fj.Status = status
	}
	if errorMsg != "" {
		fj.Error = errorMsg
	}
	if maxAttempts > 0 {
		fj.MaxAttempts = maxAttempts
	}
	if delay >= 0 {
		fj.Delay = delay
	}
	fj.UpdatedAt = time.Now()
	return nil
}

// Validate validates the failed job
func (fj *FailedJob) Validate() error {
	if fj.JobID == uuid.Nil {
		return errors.New("job_id is required")
	}
	if fj.Queue == "" {
		return errors.New("queue is required")
	}
	if fj.HandlerName == "" {
		return errors.New("handler_name is required")
	}
	if fj.Payload == "" {
		return errors.New("payload is required")
	}
	if fj.MaxAttempts <= 0 {
		return errors.New("max_attempts must be greater than 0")
	}
	return nil
}
