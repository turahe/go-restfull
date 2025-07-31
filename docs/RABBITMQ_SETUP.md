# RabbitMQ Setup Guide

This guide explains how to set up and use RabbitMQ in the Go RESTful API application.

## Overview

RabbitMQ is a message broker that enables asynchronous communication between different parts of your application. This implementation provides:

- **Connection Management**: Automatic connection handling with retry logic
- **Exchange & Queue Declaration**: Automatic setup of exchanges and queues from configuration
- **Message Publishing**: Easy publishing of JSON messages to exchanges or queues
- **Message Consumption**: Simple consumer setup with automatic acknowledgment
- **Health Monitoring**: Built-in health checks and monitoring
- **Error Handling**: Comprehensive error handling and logging

## Configuration

### 1. RabbitMQ Configuration Structure

Add a `rabbitmq` section to your `config.yaml` file:

```yaml
rabbitmq:
  enable: true
  host: "localhost"
  port: 5672
  username: "guest"
  password: "guest"
  vhost: "/"
  ssl: false
  connection:
    maxRetries: 3
    retryDelay: 5
    timeout: 30
  channel:
    prefetchCount: 10
    qos: 10
  exchanges:
    - name: "user.events"
      type: "topic"
      durable: true
      autoDelete: false
      internal: false
      arguments: {}
  queues:
    - name: "user.registration"
      durable: true
      autoDelete: false
      exclusive: false
      arguments: {}
      bindings:
        - exchange: "user.events"
          routingKey: "user.registered"
```

### 2. Configuration Parameters

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `enable` | bool | Enable/disable RabbitMQ | false |
| `host` | string | RabbitMQ server host | localhost |
| `port` | int | RabbitMQ server port | 5672 |
| `username` | string | RabbitMQ username | guest |
| `password` | string | RabbitMQ password | guest |
| `vhost` | string | RabbitMQ virtual host | / |
| `ssl` | bool | Use SSL connection | false |
| `connection.maxRetries` | int | Maximum connection retries | 3 |
| `connection.retryDelay` | int | Delay between retries (seconds) | 5 |
| `connection.timeout` | int | Connection timeout (seconds) | 30 |
| `channel.prefetchCount` | int | Prefetch count for consumers | 10 |
| `channel.qos` | int | Quality of service setting | 10 |

### 3. Exchange Types

- **direct**: Routes messages based on exact routing key match
- **topic**: Routes messages based on wildcard routing key patterns
- **fanout**: Broadcasts messages to all bound queues
- **headers**: Routes messages based on header values

## Usage Examples

### 1. Basic Service Usage

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/turahe/go-restfull/internal/rabbitmq"
)

func main() {
    ctx := context.Background()
    
    // Create RabbitMQ service
    service := rabbitmq.NewService()
    
    // Initialize service
    if err := service.Initialize(ctx); err != nil {
        log.Fatalf("Failed to initialize RabbitMQ: %v", err)
    }
    
    // Publish a message
    data := map[string]interface{}{
        "user_id": "123",
        "action":  "login",
        "time":    time.Now(),
    }
    
    err := service.PublishJSON(ctx, "user.events", "user.login", data)
    if err != nil {
        log.Printf("Failed to publish: %v", err)
    }
    
    // Close service
    service.Close()
}
```

### 2. Message Publishing

```go
// Publish JSON message to exchange
data := UserEvent{
    UserID:    "user123",
    EventType: "registered",
    Timestamp: time.Now(),
}

err := service.PublishJSON(ctx, "user.events", "user.registered", data)

// Publish directly to queue
err = service.PublishToQueueJSON(ctx, "email.queue", emailData)

// Publish with custom headers
headers := map[string]interface{}{
    "X-User-ID":    "123",
    "X-Request-ID": "req-456",
    "X-Priority":   "high",
}

err = service.PublishWithHeaders(ctx, "system.events", "system.notification", data, headers)
```

### 3. Message Consumption

```go
// Consume JSON messages
err := service.ConsumeJSON(ctx, "user.registration", func(ctx context.Context, data interface{}) error {
    // Handle the message
    log.Printf("Received message: %+v", data)
    return nil
})

// Consume with specific type
err = service.ConsumeWithType(ctx, "email.queue", &EmailNotification{}, func(ctx context.Context, data interface{}) error {
    notification := data.(*EmailNotification)
    // Process email notification
    return sendEmail(notification)
})

// Consume raw messages
err = service.ConsumeRaw(ctx, "raw.queue", func(ctx context.Context, delivery rabbitmq.Delivery) error {
    log.Printf("Raw message: %s", string(delivery.Body))
    log.Printf("Headers: %+v", delivery.Headers)
    return nil
})
```

### 4. Complete Example with Producer and Consumer

```go
type UserService struct {
    rabbitMQ *rabbitmq.Service
}

func NewUserService(rabbitMQ *rabbitmq.Service) *UserService {
    return &UserService{rabbitMQ: rabbitMQ}
}

func (s *UserService) RegisterUser(ctx context.Context, user User) error {
    // Save user to database
    if err := s.saveUser(user); err != nil {
        return err
    }
    
    // Publish user registration event
    event := UserEvent{
        UserID:    user.ID,
        EventType: "registered",
        Timestamp: time.Now(),
    }
    
    return s.rabbitMQ.PublishJSON(ctx, "user.events", "user.registered", event)
}

func (s *UserService) StartEventConsumers(ctx context.Context) error {
    // Start user registration consumer
    return s.rabbitMQ.ConsumeWithType(ctx, "user.registration", &UserEvent{}, func(ctx context.Context, data interface{}) error {
        event := data.(*UserEvent)
        
        // Send welcome email
        email := EmailNotification{
            To:      event.UserID + "@example.com",
            Subject: "Welcome!",
            Body:    "Thank you for registering.",
        }
        
        return s.rabbitMQ.PublishJSON(ctx, "email.notifications", "email.send", email)
    })
}
```

## Advanced Features

### 1. Health Monitoring

```go
// Health check
if err := service.HealthCheck(ctx); err != nil {
    log.Printf("RabbitMQ health check failed: %v", err)
}

// Get queue information
queueInfo, err := service.GetQueueInfo(ctx, "user.registration")
if err != nil {
    log.Printf("Failed to get queue info: %v", err)
} else {
    log.Printf("Queue messages: %d", queueInfo["messages"])
    log.Printf("Queue consumers: %d", queueInfo["consumers"])
}
```

### 2. Queue Management

```go
// Purge queue (remove all messages)
err := service.PurgeQueue(ctx, "dead.letter")

// Delete queue
err := service.DeleteQueue(ctx, "temp.queue")
```

### 3. Error Handling

```go
// Publish with error handling
err := service.PublishJSON(ctx, "user.events", "user.login", data)
if err != nil {
    // Handle publishing error
    log.Printf("Failed to publish user login event: %v", err)
    
    // Maybe retry or store for later processing
    return err
}

// Consumer with error handling
err = service.ConsumeJSON(ctx, "user.registration", func(ctx context.Context, data interface{}) error {
    // Process message
    if err := processUserRegistration(data); err != nil {
        // Log error but don't return it to avoid message acknowledgment
        log.Printf("Failed to process user registration: %v", err)
        
        // Maybe send to dead letter queue
        return nil
    }
    
    return nil
})
```

## Best Practices

### 1. Message Structure

```go
// Use structured message types
type UserEvent struct {
    UserID    string    `json:"user_id"`
    EventType string    `json:"event_type"`
    Data      string    `json:"data"`
    Timestamp time.Time `json:"timestamp"`
}

// Include metadata in headers
headers := map[string]interface{}{
    "X-User-ID":    userID,
    "X-Request-ID": requestID,
    "X-Version":    "1.0",
}
```

### 2. Error Handling

```go
// Always handle errors in consumers
err = service.ConsumeJSON(ctx, "queue.name", func(ctx context.Context, data interface{}) error {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Panic in consumer: %v", r)
        }
    }()
    
    // Process message
    return processMessage(data)
})
```

### 3. Connection Management

```go
// Initialize once at startup
func (app *Application) Start() error {
    // Initialize RabbitMQ
    if err := app.rabbitMQ.Initialize(context.Background()); err != nil {
        return fmt.Errorf("failed to initialize RabbitMQ: %w", err)
    }
    
    // Start consumers
    if err := app.startConsumers(); err != nil {
        return fmt.Errorf("failed to start consumers: %w", err)
    }
    
    return nil
}

// Clean shutdown
func (app *Application) Shutdown() error {
    return app.rabbitMQ.Close()
}
```

### 4. Monitoring and Logging

```go
// Regular health checks
func monitorRabbitMQ() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        if err := service.HealthCheck(context.Background()); err != nil {
            log.Printf("RabbitMQ health check failed: %v", err)
            // Maybe trigger alert or restart connection
        }
    }
}
```

## Troubleshooting

### Common Issues

1. **Connection Refused**
   - Check if RabbitMQ server is running
   - Verify host and port configuration
   - Check firewall settings

2. **Authentication Failed**
   - Verify username and password
   - Check virtual host permissions

3. **Queue Not Found**
   - Ensure queue is declared in configuration
   - Check if queue name matches exactly

4. **Message Not Delivered**
   - Verify exchange and routing key
   - Check queue bindings
   - Ensure consumer is running

### Debugging

Enable debug logging:

```yaml
log:
  level: "debug"
```

Check RabbitMQ management interface:
- Default URL: http://localhost:15672
- Username: guest
- Password: guest

### Performance Tuning

```yaml
rabbitmq:
  channel:
    prefetchCount: 50  # Increase for high throughput
    qos: 50
  connection:
    timeout: 60        # Increase for slow networks
```

## Security Considerations

1. **Use SSL in Production**
   ```yaml
   rabbitmq:
     ssl: true
   ```

2. **Strong Authentication**
   ```yaml
   rabbitmq:
     username: "app_user"
     password: "strong_password"
   ```

3. **Virtual Host Isolation**
   ```yaml
   rabbitmq:
     vhost: "/app_production"
   ```

4. **Environment Variables**
   ```yaml
   rabbitmq:
     password: "${RABBITMQ_PASSWORD}"
   ```

## Integration with Existing Code

The RabbitMQ service can be easily integrated with existing services:

```go
// In your user controller
func (c *UserController) Register(ctx *fiber.Ctx) error {
    // Create user
    user, err := c.userService.CreateUser(ctx.Context(), userData)
    if err != nil {
        return err
    }
    
    // Publish event
    event := UserEvent{
        UserID:    user.ID,
        EventType: "registered",
        Timestamp: time.Now(),
    }
    
    if err := c.rabbitMQ.PublishJSON(ctx.Context(), "user.events", "user.registered", event); err != nil {
        // Log error but don't fail the request
        log.Printf("Failed to publish user event: %v", err)
    }
    
    return ctx.JSON(user)
}
```

This setup provides a robust, scalable message queuing solution that integrates seamlessly with your existing application architecture. 