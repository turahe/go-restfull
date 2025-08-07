# Messaging System

The messaging system provides asynchronous message processing using RabbitMQ for reliable, scalable message queuing.

## 🚀 Overview

The messaging system replaces the previous job service with a more robust, enterprise-grade solution using RabbitMQ. This provides better reliability, scalability, and monitoring capabilities.

## 📋 Features

### Message Types
- **Email Messages** - Asynchronous email sending
- **Backup Messages** - Database backup operations
- **Cleanup Messages** - System cleanup tasks
- **Notification Messages** - User notifications

### Queues
- `email_queue` - Email processing
- `backup_queue` - Database backup operations
- `cleanup_queue` - System cleanup tasks
- `notification_queue` - User notifications
- `failed_queue` - Failed message handling

## 🔧 Configuration

### RabbitMQ Configuration
```yaml
rabbitmq:
  host: "localhost"
  port: 5672
  username: "guest"
  password: "guest"
  vhost: "/"
```

### Message Configuration
```go
// Message structure
type Message struct {
    ID          string          `json:"id"`
    Type        string          `json:"type"`
    Payload     json.RawMessage `json:"payload"`
    MaxAttempts int             `json:"max_attempts"`
    Delay       int             `json:"delay"`
    Priority    int             `json:"priority"`
    CreatedAt   time.Time       `json:"created_at"`
}
```

## 📨 Usage Examples

### Sending Email
```go
err := messagingService.SendEmail(ctx, "user@example.com", "Subject", "Body", "<html>Body</html>")
```

### Sending Backup
```go
err := messagingService.SendBackup(ctx, "my_database", "/backups/", true)
```

### Sending Cleanup
```go
err := messagingService.SendCleanup(ctx, "logs", 30)
```

### Sending Notification
```go
err := messagingService.SendNotification(ctx, "user123", "email", "Title", "Message", nil)
```

## 🔄 Message Processing

### Handler Implementation
```go
// EmailHandler handles email messages
type EmailHandler struct{}

func (h *EmailHandler) Handle(ctx context.Context, message *Message) error {
    var emailMsg EmailMessage
    err := json.Unmarshal(message.Payload, &emailMsg)
    if err != nil {
        return fmt.Errorf("failed to unmarshal email message: %w", err)
    }
    
    // Process email logic here
    return nil
}
```

### Handler Factory
```go
// CreateHandler creates a handler based on the message type
func (f *HandlerFactory) CreateHandler(messageType string) (MessageHandler, error) {
    switch messageType {
    case "email":
        return NewEmailHandler(), nil
    case "backup":
        return NewBackupHandler(), nil
    case "cleanup":
        return NewCleanupHandler(), nil
    case "notification":
        return NewNotificationHandler(), nil
    default:
        return nil, fmt.Errorf("unknown message type: %s", messageType)
    }
}
```

## 🚀 Deployment

### RabbitMQ Setup
```bash
# Using Docker
docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management

# Using Docker Compose
version: '3.8'
services:
  rabbitmq:
    image: rabbitmq:3-management
    ports:
      - "5672:5672"
      - "15672:15672"
    environment:
      RABBITMQ_DEFAULT_USER: guest
      RABBITMQ_DEFAULT_PASS: guest
```

### Application Startup
The messaging consumers start automatically when the application starts:

```go
// Start messaging consumers
err = container.MessagingService.StartConsumers(context.Background())
if err != nil {
    log.Fatalf("Failed to start messaging consumers: %v", err)
}
```

## 📊 Monitoring

### RabbitMQ Management Interface
- **URL**: http://localhost:15672
- **Username**: guest
- **Password**: guest

### Metrics Available
- Queue depths
- Message rates
- Consumer status
- Connection health
- Failed message analysis

## 🔍 Troubleshooting

### Common Issues

#### Connection Failed
```bash
# Check RabbitMQ status
docker ps | grep rabbitmq

# Check logs
docker logs rabbitmq
```

#### Messages Not Processing
```bash
# Check queue status
curl -u guest:guest http://localhost:15672/api/queues

# Check consumer status
curl -u guest:guest http://localhost:15672/api/consumers
```

#### Failed Messages
```bash
# Check failed queue
curl -u guest:guest http://localhost:15672/api/queues/%2F/failed_queue
```

## 🔗 Related Documentation

- [RabbitMQ Migration](../migration/rabbitmq-migration.md) - Migration details
- [Architecture Documentation](../architecture/) - System architecture
- [Development Documentation](../development/) - Development practices
