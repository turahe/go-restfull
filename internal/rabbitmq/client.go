package rabbitmq

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/turahe/go-restfull/config"
	"github.com/turahe/go-restfull/internal/logger"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// Client represents a RabbitMQ client
type Client struct {
	config     *config.RabbitMQ
	connection *amqp.Connection
	channel    *amqp.Channel
	mu         sync.RWMutex
	closed     bool
}

// Message represents a RabbitMQ message
type Message struct {
	Exchange   string
	RoutingKey string
	Body       []byte
	Headers    amqp.Table
	Priority   uint8
	Timestamp  time.Time
}

// Delivery represents a received message
type Delivery struct {
	amqp.Delivery
}

// ConsumerFunc represents a message consumer function
type ConsumerFunc func(ctx context.Context, delivery Delivery) error

// NewClient creates a new RabbitMQ client
func NewClient(cfg *config.RabbitMQ) *Client {
	return &Client{
		config: cfg,
	}
}

// Connect establishes a connection to RabbitMQ
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connection != nil && !c.connection.IsClosed() {
		return nil
	}

	// Build connection URL
	protocol := "amqp"
	if c.config.SSL {
		protocol = "amqps"
	}

	url := fmt.Sprintf("%s://%s:%s@%s:%d/%s",
		protocol,
		c.config.Username,
		c.config.Password,
		c.config.Host,
		c.config.Port,
		c.config.VHost,
	)

	// Configure connection
	connConfig := amqp.Config{
		Dial: amqp.DefaultDial(time.Duration(c.config.Connection.Timeout) * time.Second),
	}

	// Establish connection with retry logic
	var conn *amqp.Connection
	var err error

	for i := 0; i <= c.config.Connection.MaxRetries; i++ {
		conn, err = amqp.DialConfig(url, connConfig)
		if err == nil {
			break
		}

		if i < c.config.Connection.MaxRetries {
			if logger.Log != nil {
				logger.Log.Warn("Failed to connect to RabbitMQ, retrying...",
					zap.Error(err),
					zap.Int("attempt", i+1),
					zap.Int("maxRetries", c.config.Connection.MaxRetries),
				)
			}
			time.Sleep(time.Duration(c.config.Connection.RetryDelay) * time.Second)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ after %d retries: %w", c.config.Connection.MaxRetries, err)
	}

	c.connection = conn

	// Create channel
	channel, err := c.connection.Channel()
	if err != nil {
		c.connection.Close()
		return fmt.Errorf("failed to create channel: %w", err)
	}

	// Configure channel QoS
	if c.config.Channel.Qos > 0 {
		err = channel.Qos(c.config.Channel.Qos, 0, false)
		if err != nil {
			channel.Close()
			c.connection.Close()
			return fmt.Errorf("failed to set QoS: %w", err)
		}
	}

	c.channel = channel

	if logger.Log != nil {
		logger.Log.Info("Successfully connected to RabbitMQ",
			zap.String("host", c.config.Host),
			zap.Int("port", c.config.Port),
			zap.String("vhost", c.config.VHost),
		)
	}

	return nil
}

// DeclareExchanges declares all configured exchanges
func (c *Client) DeclareExchanges(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.channel == nil {
		return fmt.Errorf("channel not initialized")
	}

	for _, exchangeConfig := range config.GetAllExchangeConfigs() {
		err := c.channel.ExchangeDeclare(
			exchangeConfig.Name,
			exchangeConfig.Type,
			exchangeConfig.Durable,
			exchangeConfig.AutoDelete,
			exchangeConfig.Internal,
			false, // no-wait
			convertArguments(exchangeConfig.Arguments),
		)

		if err != nil {
			return fmt.Errorf("failed to declare exchange '%s': %w", exchangeConfig.Name, err)
		}

		if logger.Log != nil {
			logger.Log.Info("Declared exchange",
				zap.String("name", exchangeConfig.Name),
				zap.String("type", exchangeConfig.Type),
			)
		}
	}

	return nil
}

// DeclareQueues declares all configured queues
func (c *Client) DeclareQueues(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.channel == nil {
		return fmt.Errorf("channel not initialized")
	}

	for _, queueConfig := range config.GetAllQueueConfigs() {
		_, err := c.channel.QueueDeclare(
			queueConfig.Name,
			queueConfig.Durable,
			queueConfig.AutoDelete,
			queueConfig.Exclusive,
			false, // no-wait
			convertArguments(queueConfig.Arguments),
		)

		if err != nil {
			return fmt.Errorf("failed to declare queue '%s': %w", queueConfig.Name, err)
		}

		// Bind queue to exchanges
		for _, binding := range queueConfig.Bindings {
			err = c.channel.QueueBind(
				queueConfig.Name,
				binding.RoutingKey,
				binding.Exchange,
				false, // no-wait
				nil,   // arguments
			)

			if err != nil {
				return fmt.Errorf("failed to bind queue '%s' to exchange '%s': %w", queueConfig.Name, binding.Exchange, err)
			}
		}

		if logger.Log != nil {
			logger.Log.Info("Declared queue",
				zap.String("name", queueConfig.Name),
				zap.Int("bindings", len(queueConfig.Bindings)),
			)
		}
	}

	return nil
}

// Publish publishes a message to an exchange
func (c *Client) Publish(ctx context.Context, msg Message) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.channel == nil {
		return fmt.Errorf("channel not initialized")
	}

	publishing := amqp.Publishing{
		ContentType:  "application/json",
		Body:         msg.Body,
		Headers:      msg.Headers,
		Priority:     msg.Priority,
		Timestamp:    msg.Timestamp,
		DeliveryMode: amqp.Persistent,
	}

	err := c.channel.PublishWithContext(ctx,
		msg.Exchange,
		msg.RoutingKey,
		false, // mandatory
		false, // immediate
		publishing,
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	if logger.Log != nil {
		logger.Log.Debug("Published message",
			zap.String("exchange", msg.Exchange),
			zap.String("routingKey", msg.RoutingKey),
			zap.Int("bodySize", len(msg.Body)),
		)
	}

	return nil
}

// PublishToQueue publishes a message directly to a queue
func (c *Client) PublishToQueue(ctx context.Context, queueName string, msg Message) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.channel == nil {
		return fmt.Errorf("channel not initialized")
	}

	publishing := amqp.Publishing{
		ContentType:  "application/json",
		Body:         msg.Body,
		Headers:      msg.Headers,
		Priority:     msg.Priority,
		Timestamp:    msg.Timestamp,
		DeliveryMode: amqp.Persistent,
	}

	err := c.channel.PublishWithContext(ctx,
		"", // empty exchange for direct queue publishing
		queueName,
		false, // mandatory
		false, // immediate
		publishing,
	)

	if err != nil {
		return fmt.Errorf("failed to publish message to queue '%s': %w", queueName, err)
	}

	if logger.Log != nil {
		logger.Log.Debug("Published message to queue",
			zap.String("queue", queueName),
			zap.Int("bodySize", len(msg.Body)),
		)
	}

	return nil
}

// Consume starts consuming messages from a queue
func (c *Client) Consume(ctx context.Context, queueName string, consumer ConsumerFunc) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.channel == nil {
		return fmt.Errorf("channel not initialized")
	}

	// Set prefetch count for this consumer
	if c.config.Channel.PrefetchCount > 0 {
		err := c.channel.Qos(c.config.Channel.PrefetchCount, 0, false)
		if err != nil {
			return fmt.Errorf("failed to set prefetch count: %w", err)
		}
	}

	deliveries, err := c.channel.Consume(
		queueName,
		"",    // consumer tag (empty for auto-generated)
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // arguments
	)

	if err != nil {
		return fmt.Errorf("failed to start consuming from queue '%s': %w", queueName, err)
	}

	if logger.Log != nil {
		logger.Log.Info("Started consuming from queue", zap.String("queue", queueName))
	}

	// Start consuming messages
	go func() {
		for {
			select {
			case <-ctx.Done():
				if logger.Log != nil {
					logger.Log.Info("Stopping consumer", zap.String("queue", queueName))
				}
				return
			case delivery := <-deliveries:
				if delivery.Acknowledger == nil {
					// Channel was closed
					return
				}

				// Process message
				err := consumer(ctx, Delivery{delivery})
				if err != nil {
					if logger.Log != nil {
						logger.Log.Error("Failed to process message",
							zap.Error(err),
							zap.String("queue", queueName),
							zap.Uint64("deliveryTag", delivery.DeliveryTag),
						)
					}
					// Reject message and requeue
					delivery.Nack(false, true)
				} else {
					// Acknowledge message
					delivery.Ack(false)
				}
			}
		}
	}()

	return nil
}

// GetChannel returns the underlying AMQP channel
func (c *Client) GetChannel() *amqp.Channel {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.channel
}

// GetConnection returns the underlying AMQP connection
func (c *Client) GetConnection() *amqp.Connection {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connection
}

// IsConnected checks if the client is connected
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connection != nil && !c.connection.IsClosed() && c.channel != nil
}

// Close closes the RabbitMQ connection
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	var errs []error

	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close channel: %w", err))
		}
		c.channel = nil
	}

	if c.connection != nil {
		if err := c.connection.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close connection: %w", err))
		}
		c.connection = nil
	}

	c.closed = true

	if logger.Log != nil {
		logger.Log.Info("RabbitMQ client closed")
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during close: %v", errs)
	}

	return nil
}

// convertArguments converts string map to amqp.Table
func convertArguments(args map[string]string) amqp.Table {
	if args == nil {
		return nil
	}

	table := make(amqp.Table)
	for k, v := range args {
		table[k] = v
	}
	return table
}
