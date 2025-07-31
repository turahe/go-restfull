package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/turahe/go-restfull/pkg/logger"
	"sync"
	"time"

	"github.com/turahe/go-restfull/config"
	"go.uber.org/zap"
)

// Service represents a RabbitMQ service
type Service struct {
	client *Client
	mu     sync.RWMutex
}

// NewService creates a new RabbitMQ service
func NewService() *Service {
	return &Service{}
}

// Initialize initializes the RabbitMQ service
func (s *Service) Initialize(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !config.IsRabbitMQEnabled() {
		if logger.Log != nil {
			logger.Log.Info("RabbitMQ is disabled, skipping initialization")
		}
		return nil
	}

	rabbitMQConfig := config.GetRabbitMQConfig()
	s.client = NewClient(rabbitMQConfig)

	// Connect to RabbitMQ
	if err := s.client.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Declare exchanges
	if err := s.client.DeclareExchanges(ctx); err != nil {
		return fmt.Errorf("failed to declare exchanges: %w", err)
	}

	// Declare queues
	if err := s.client.DeclareQueues(ctx); err != nil {
		return fmt.Errorf("failed to declare queues: %w", err)
	}

	if logger.Log != nil {
		logger.Log.Info("RabbitMQ service initialized successfully")
	}

	return nil
}

// PublishJSON publishes a JSON message to an exchange
func (s *Service) PublishJSON(ctx context.Context, exchange, routingKey string, data interface{}) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.client == nil {
		return fmt.Errorf("RabbitMQ client not initialized")
	}

	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	msg := Message{
		Exchange:   exchange,
		RoutingKey: routingKey,
		Body:       body,
		Timestamp:  time.Now(),
	}

	return s.client.Publish(ctx, msg)
}

// PublishToQueueJSON publishes a JSON message directly to a queue
func (s *Service) PublishToQueueJSON(ctx context.Context, queueName string, data interface{}) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.client == nil {
		return fmt.Errorf("RabbitMQ client not initialized")
	}

	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	msg := Message{
		Body:      body,
		Timestamp: time.Now(),
	}

	return s.client.PublishToQueue(ctx, queueName, msg)
}

// PublishWithHeaders publishes a message with custom headers
func (s *Service) PublishWithHeaders(ctx context.Context, exchange, routingKey string, data interface{}, headers map[string]interface{}) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.client == nil {
		return fmt.Errorf("RabbitMQ client not initialized")
	}

	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	msg := Message{
		Exchange:   exchange,
		RoutingKey: routingKey,
		Body:       body,
		Headers:    headers,
		Timestamp:  time.Now(),
	}

	return s.client.Publish(ctx, msg)
}

// ConsumeJSON starts consuming JSON messages from a queue
func (s *Service) ConsumeJSON(ctx context.Context, queueName string, handler func(ctx context.Context, data interface{}) error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.client == nil {
		return fmt.Errorf("RabbitMQ client not initialized")
	}

	consumer := func(ctx context.Context, delivery Delivery) error {
		var data interface{}
		if err := json.Unmarshal(delivery.Body, &data); err != nil {
			return fmt.Errorf("failed to unmarshal JSON: %w", err)
		}

		return handler(ctx, data)
	}

	return s.client.Consume(ctx, queueName, consumer)
}

// ConsumeWithType starts consuming messages with a specific type
func (s *Service) ConsumeWithType(ctx context.Context, queueName string, dataType interface{}, handler func(ctx context.Context, data interface{}) error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.client == nil {
		return fmt.Errorf("RabbitMQ client not initialized")
	}

	consumer := func(ctx context.Context, delivery Delivery) error {
		if err := json.Unmarshal(delivery.Body, dataType); err != nil {
			return fmt.Errorf("failed to unmarshal JSON: %w", err)
		}

		return handler(ctx, dataType)
	}

	return s.client.Consume(ctx, queueName, consumer)
}

// ConsumeRaw starts consuming raw messages from a queue
func (s *Service) ConsumeRaw(ctx context.Context, queueName string, handler func(ctx context.Context, delivery Delivery) error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.client == nil {
		return fmt.Errorf("RabbitMQ client not initialized")
	}

	return s.client.Consume(ctx, queueName, handler)
}

// IsConnected checks if the service is connected to RabbitMQ
func (s *Service) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.client == nil {
		return false
	}

	return s.client.IsConnected()
}

// GetClient returns the underlying RabbitMQ client
func (s *Service) GetClient() *Client {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.client
}

// Close closes the RabbitMQ service
func (s *Service) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		return nil
	}

	return s.client.Close()
}

// HealthCheck performs a health check on the RabbitMQ service
func (s *Service) HealthCheck(ctx context.Context) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !config.IsRabbitMQEnabled() {
		return nil // RabbitMQ is disabled, consider it healthy
	}

	if s.client == nil {
		return fmt.Errorf("RabbitMQ client not initialized")
	}

	if !s.client.IsConnected() {
		return fmt.Errorf("RabbitMQ client not connected")
	}

	// Try to get channel info to verify connection is working
	channel := s.client.GetChannel()
	if channel == nil {
		return fmt.Errorf("RabbitMQ channel not available")
	}

	// Check if channel is open
	if channel.IsClosed() {
		return fmt.Errorf("RabbitMQ channel is closed")
	}

	return nil
}

// GetQueueInfo returns information about a queue
func (s *Service) GetQueueInfo(ctx context.Context, queueName string) (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.client == nil {
		return nil, fmt.Errorf("RabbitMQ client not initialized")
	}

	channel := s.client.GetChannel()
	if channel == nil {
		return nil, fmt.Errorf("RabbitMQ channel not available")
	}

	queue, err := channel.QueueInspect(queueName)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect queue '%s': %w", queueName, err)
	}

	info := map[string]interface{}{
		"name":      queue.Name,
		"messages":  queue.Messages,
		"consumers": queue.Consumers,
	}

	return info, nil
}

// PurgeQueue purges all messages from a queue
func (s *Service) PurgeQueue(ctx context.Context, queueName string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.client == nil {
		return fmt.Errorf("RabbitMQ client not initialized")
	}

	channel := s.client.GetChannel()
	if channel == nil {
		return fmt.Errorf("RabbitMQ channel not available")
	}

	count, err := channel.QueuePurge(queueName, false)
	if err != nil {
		return fmt.Errorf("failed to purge queue '%s': %w", queueName, err)
	}

	if logger.Log != nil {
		logger.Log.Info("Purged queue",
			zap.String("queue", queueName),
			zap.Int("messagesRemoved", count),
		)
	}

	return nil
}

// DeleteQueue deletes a queue
func (s *Service) DeleteQueue(ctx context.Context, queueName string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.client == nil {
		return fmt.Errorf("RabbitMQ client not initialized")
	}

	channel := s.client.GetChannel()
	if channel == nil {
		return fmt.Errorf("RabbitMQ channel not available")
	}

	count, err := channel.QueueDelete(queueName, false, false, false)
	if err != nil {
		return fmt.Errorf("failed to delete queue '%s': %w", queueName, err)
	}

	if logger.Log != nil {
		logger.Log.Info("Deleted queue",
			zap.String("queue", queueName),
			zap.Int("messagesRemoved", count),
		)
	}

	return nil
}
