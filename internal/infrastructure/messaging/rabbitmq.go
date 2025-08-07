package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
	"github.com/turahe/go-restfull/config"
	"github.com/turahe/go-restfull/internal/logger"

	"go.uber.org/zap"
)

// RabbitMQClient represents a RabbitMQ client
type RabbitMQClient struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	config  config.RabbitMQ
}

// Message represents a message to be sent to RabbitMQ
type Message struct {
	ID          string          `json:"id"`
	Type        string          `json:"type"`
	Payload     json.RawMessage `json:"payload"`
	CreatedAt   time.Time       `json:"created_at"`
	MaxAttempts int             `json:"max_attempts"`
	Attempts    int             `json:"attempts"`
	Delay       int             `json:"delay"`
	Priority    int             `json:"priority"`
}

// NewRabbitMQClient creates a new RabbitMQ client
func NewRabbitMQClient(cfg config.RabbitMQ) (*RabbitMQClient, error) {
	client := &RabbitMQClient{
		config: cfg,
	}

	err := client.connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	return client, nil
}

// connect establishes connection to RabbitMQ
func (r *RabbitMQClient) connect() error {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		r.config.Username,
		r.config.Password,
		r.config.Host,
		r.config.Port,
		r.config.VHost,
	)

	conn, err := amqp091.Dial(url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	r.conn = conn
	r.channel = channel

	// Declare exchanges and queues
	err = r.setupExchangesAndQueues()
	if err != nil {
		return fmt.Errorf("failed to setup exchanges and queues: %w", err)
	}

	logger.Log.Info("RabbitMQ client connected successfully")
	return nil
}

// setupExchangesAndQueues declares the necessary exchanges and queues
func (r *RabbitMQClient) setupExchangesAndQueues() error {
	// Declare main exchange
	err := r.channel.ExchangeDeclare(
		"app_exchange", // name
		"direct",       // type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare queues
	queues := []string{"email_queue", "backup_queue", "cleanup_queue", "notification_queue"}

	for _, queueName := range queues {
		_, err := r.channel.QueueDeclare(
			queueName, // name
			true,      // durable
			false,     // delete when unused
			false,     // exclusive
			false,     // no-wait
			amqp091.Table{
				"x-dead-letter-exchange":    "app_exchange",
				"x-dead-letter-routing-key": "failed",
			}, // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", queueName, err)
		}

		// Bind queue to exchange
		err = r.channel.QueueBind(
			queueName,      // queue name
			queueName,      // routing key
			"app_exchange", // exchange
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue %s: %w", queueName, err)
		}
	}

	// Declare failed queue
	_, err := r.channel.QueueDeclare(
		"failed_queue", // name
		true,           // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare failed queue: %w", err)
	}

	// Bind failed queue
	err = r.channel.QueueBind(
		"failed_queue", // queue name
		"failed",       // routing key
		"app_exchange", // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind failed queue: %w", err)
	}

	return nil
}

// PublishMessage publishes a message to a specific queue
func (r *RabbitMQClient) PublishMessage(ctx context.Context, queueName string, message *Message) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = r.channel.PublishWithContext(ctx,
		"app_exchange", // exchange
		queueName,      // routing key
		false,          // mandatory
		false,          // immediate
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp091.Persistent,
			Priority:     uint8(message.Priority),
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	logger.Log.Info("Message published successfully",
		zap.String("queue", queueName),
		zap.String("message_id", message.ID),
		zap.String("type", message.Type),
	)
	return nil
}

// ConsumeMessages starts consuming messages from a queue
func (r *RabbitMQClient) ConsumeMessages(ctx context.Context, queueName string, handler MessageHandler) error {
	msgs, err := r.channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	logger.Log.Info("Started consuming messages", zap.String("queue", queueName))

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("Stopping message consumption", zap.String("queue", queueName))
			return nil
		case msg := <-msgs:
			go func(msg amqp091.Delivery) {
				if err := r.processMessage(ctx, msg, handler); err != nil {
					logger.Log.Error("Failed to process message",
						zap.Error(err),
						zap.String("queue", queueName),
					)
					// Reject the message and requeue
					msg.Nack(false, true)
				} else {
					// Acknowledge the message
					msg.Ack(false)
				}
			}(msg)
		}
	}
}

// processMessage processes a single message
func (r *RabbitMQClient) processMessage(ctx context.Context, msg amqp091.Delivery, handler MessageHandler) error {
	var message Message
	err := json.Unmarshal(msg.Body, &message)
	if err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	logger.Log.Info("Processing message",
		zap.String("message_id", message.ID),
		zap.String("type", message.Type),
	)

	// Check if message has exceeded max attempts
	if message.Attempts >= message.MaxAttempts {
		logger.Log.Warn("Message exceeded max attempts, moving to failed queue",
			zap.String("message_id", message.ID),
			zap.Int("attempts", message.Attempts),
			zap.Int("max_attempts", message.MaxAttempts),
		)
		return r.moveToFailedQueue(ctx, &message)
	}

	// Increment attempts
	message.Attempts++

	// Process the message
	err = handler.Handle(ctx, &message)
	if err != nil {
		logger.Log.Error("Message processing failed",
			zap.Error(err),
			zap.String("message_id", message.ID),
		)
		return err
	}

	logger.Log.Info("Message processed successfully",
		zap.String("message_id", message.ID),
		zap.String("type", message.Type),
	)
	return nil
}

// moveToFailedQueue moves a message to the failed queue
func (r *RabbitMQClient) moveToFailedQueue(ctx context.Context, message *Message) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal failed message: %w", err)
	}

	err = r.channel.PublishWithContext(ctx,
		"app_exchange", // exchange
		"failed",       // routing key
		false,          // mandatory
		false,          // immediate
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp091.Persistent,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish to failed queue: %w", err)
	}

	return nil
}

// Close closes the RabbitMQ connection
func (r *RabbitMQClient) Close() error {
	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			logger.Log.Error("Failed to close channel", zap.Error(err))
		}
	}
	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			logger.Log.Error("Failed to close connection", zap.Error(err))
		}
	}
	return nil
}

// MessageHandler defines the interface for message handlers
type MessageHandler interface {
	Handle(ctx context.Context, message *Message) error
}

// NewMessage creates a new message
func NewMessage(messageType string, payload interface{}, maxAttempts, delay, priority int) (*Message, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	return &Message{
		ID:          uuid.New().String(),
		Type:        messageType,
		Payload:     payloadBytes,
		CreatedAt:   time.Now(),
		MaxAttempts: maxAttempts,
		Attempts:    0,
		Delay:       delay,
		Priority:    priority,
	}, nil
}
