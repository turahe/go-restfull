# Notification System

The notification system provides a comprehensive way to send, manage, and deliver notifications to users across multiple channels.

## Features

### Core Functionality
- **Multi-channel notifications**: Email, SMS, Push, In-app, and Webhook
- **Notification templates**: Predefined templates for common scenarios
- **User preferences**: Configurable notification preferences per user
- **Priority levels**: Low, Normal, High, and Urgent priorities
- **Status tracking**: Unread, Read, Archived, and Deleted states
- **Bulk operations**: Send notifications to multiple users at once
- **Delivery tracking**: Monitor delivery status and retry failed deliveries

### Notification Types

#### User-related
- User registration
- Password changes
- Email verification
- Profile updates

#### System
- System alerts
- Maintenance notices
- Security alerts

#### Content
- New posts
- Comment replies
- Mentions
- Likes

#### Organization
- Invitations
- Role changes
- Permission updates

## API Endpoints

### Notifications

#### Get User Notifications
```http
GET /api/v1/notifications?page=1&per_page=10&sort_by=created_at&sort_desc=true
Authorization: Bearer <token>
```

#### Get Unread Notifications
```http
GET /api/v1/notifications/unread?page=1&per_page=10
Authorization: Bearer <token>
```

#### Get Notification by ID
```http
GET /api/v1/notifications/{id}
Authorization: Bearer <token>
```

#### Mark as Read
```http
PUT /api/v1/notifications/{id}/read
Authorization: Bearer <token>
```

#### Mark as Unread
```http
PUT /api/v1/notifications/{id}/unread
Authorization: Bearer <token>
```

#### Archive Notification
```http
PUT /api/v1/notifications/{id}/archive
Authorization: Bearer <token>
```

#### Delete Notification
```http
DELETE /api/v1/notifications/{id}
Authorization: Bearer <token>
```

#### Bulk Mark as Read
```http
POST /api/v1/notifications/bulk/read
Authorization: Bearer <token>
Content-Type: application/json

{
  "ids": ["uuid1", "uuid2", "uuid3"]
}
```

#### Get Notification Count
```http
GET /api/v1/notifications/count
Authorization: Bearer <token>
```

### Notification Preferences

#### Get User Preferences
```http
GET /api/v1/notifications/preferences
Authorization: Bearer <token>
```

#### Update User Preferences
```http
PUT /api/v1/notifications/preferences
Authorization: Bearer <token>
Content-Type: application/json

[
  {
    "type": "new_post",
    "email": true,
    "sms": false,
    "push": true,
    "in_app": true,
    "webhook": false
  }
]
```

## Database Schema

### Notifications Table
```sql
CREATE TABLE notifications (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    data JSONB,
    priority VARCHAR(20) NOT NULL DEFAULT 'normal',
    status VARCHAR(20) NOT NULL DEFAULT 'unread',
    channels JSONB NOT NULL,
    read_at TIMESTAMP WITH TIME ZONE,
    archived_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE
);
```

### Notification Templates Table
```sql
CREATE TABLE notification_templates (
    id UUID PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    subject VARCHAR(255),
    channels JSONB NOT NULL,
    priority VARCHAR(20) NOT NULL DEFAULT 'normal',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);
```

### Notification Preferences Table
```sql
CREATE TABLE notification_preferences (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL,
    email BOOLEAN NOT NULL DEFAULT true,
    sms BOOLEAN NOT NULL DEFAULT false,
    push BOOLEAN NOT NULL DEFAULT true,
    in_app BOOLEAN NOT NULL DEFAULT true,
    webhook BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE(user_id, type)
);
```

### Notification Deliveries Table
```sql
CREATE TABLE notification_deliveries (
    id UUID PRIMARY KEY,
    notification_id UUID NOT NULL,
    channel VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    attempts INTEGER NOT NULL DEFAULT 0,
    last_attempt_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    error TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);
```

## Usage Examples

### Sending a Welcome Notification
```go
// Using template
err := notificationService.SendNotificationFromTemplate(
    ctx, 
    userID, 
    "welcome_user", 
    map[string]interface{}{
        "username": "john_doe",
    },
)

// Custom notification
err := notificationService.SendCustomNotification(
    ctx,
    userID,
    entities.NotificationTypeUserRegistration,
    "Welcome!",
    "Welcome to our platform, John!",
    map[string]interface{}{"username": "john_doe"},
    entities.NotificationPriorityNormal,
    []entities.NotificationChannel{
        entities.NotificationChannelEmail,
        entities.NotificationChannelInApp,
    },
)
```

### Sending System Maintenance Notice
```go
err := notificationService.SendMaintenanceNotification(
    ctx,
    userIDs,
    "Scheduled Maintenance",
    "We will be performing maintenance on Sunday from 2-4 AM UTC",
    "2024-01-14 02:00:00",
)
```

### Sending Security Alert
```go
err := notificationService.SendSecurityAlert(
    ctx,
    userIDs,
    "Security Alert",
    "Unusual login activity detected from a new location",
)
```

## Configuration

### Default Templates
The system comes with predefined templates:
- `welcome_user`: Welcome message for new users
- `password_reset`: Password reset requests
- `email_verification`: Email verification
- `new_post_notification`: New post alerts
- `comment_reply`: Comment reply notifications
- `system_maintenance`: Maintenance notices
- `security_alert`: Security alerts

### Default Preferences
Users get default preferences for all notification types:
- **Email**: Enabled for most notifications
- **SMS**: Disabled by default (enabled for security alerts)
- **Push**: Enabled for most notifications
- **In-app**: Enabled for most notifications
- **Webhook**: Disabled by default

## Security

- All notification endpoints require authentication
- Users can only access their own notifications
- RBAC middleware controls access to notification management
- Sensitive notifications (security alerts) have additional delivery channels

## Performance

- Indexes on frequently queried fields
- Pagination support for large notification lists
- Soft deletes to maintain data integrity
- Automatic cleanup of expired and old archived notifications

## Future Enhancements

- **Real-time notifications**: WebSocket support for instant delivery
- **Advanced templating**: More sophisticated template engine
- **Scheduling**: Send notifications at specific times
- **Analytics**: Track notification engagement and delivery rates
- **Mobile push**: Native mobile push notification support
- **Internationalization**: Multi-language notification support
