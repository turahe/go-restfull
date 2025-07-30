package job

import (
	"context"
	"time"

	"github.com/turahe/go-restfull/internal/logger"

	"go.uber.org/zap"
)

// QueueProcessor handles job processing
type QueueProcessor struct {
	handlerMap HandlerMap
	running    bool
	stopChan   chan bool
}

// NewQueueProcessor creates a new queue processor
func NewQueueProcessor() *QueueProcessor {
	return &QueueProcessor{
		handlerMap: NewHandlerMap(),
		stopChan:   make(chan bool),
	}
}

// Start starts the queue processor
func (qp *QueueProcessor) Start(ctx context.Context) {
	qp.running = true
	logger.Log.Info("Job queue processor started")

	for qp.running {
		select {
		case <-ctx.Done():
			qp.running = false
			return
		case <-qp.stopChan:
			qp.running = false
			return
		default:
			// Process jobs here
			time.Sleep(1 * time.Second)
		}
	}
}

// Stop stops the queue processor
func (qp *QueueProcessor) Stop() {
	qp.running = false
	qp.stopChan <- true
	logger.Log.Info("Job queue processor stopped")
}

// ProcessJob processes a single job
func (qp *QueueProcessor) ProcessJob(job *Job) error {
	handlerFactory, exists := qp.handlerMap[job.HandlerName]
	if !exists {
		logger.Log.Warn("Handler not found", zap.String("handler", job.HandlerName))
		return nil
	}

	handler := handlerFactory()
	return handler.Handle()
}
