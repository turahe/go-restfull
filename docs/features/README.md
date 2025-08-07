# Features Documentation

This section contains documentation for specific application features and functionality.

## 📋 Contents

### Organization Management
- **[Organization Management](./organization-management.md)** - Organization feature documentation and implementation

### Backup System
- **[Backup Scheduler](./backup-scheduler.md)** - Database backup scheduling system
- **[Backup Scheduler Readme](./backup-scheduler-readme.md)** - Quick reference for backup operations

### Messaging System
- **[Messaging System](./messaging-system.md)** - RabbitMQ integration and message queuing (see migration section)

## 🏢 Organization Management

### Features
- **Multi-tenant Support** - Organization-based data isolation
- **User Management** - Organization member management
- **Role Assignment** - Organization-specific roles
- **Resource Sharing** - Shared resources within organizations

### Implementation
- **Database Design** - Organization entity relationships
- **API Endpoints** - CRUD operations for organizations
- **Authorization** - Organization-based access control
- **Validation** - Data validation and business rules

## 💾 Backup System

### Features
- **Automated Backups** - Scheduled database backups
- **Compression** - Backup file compression
- **Retention Policy** - Configurable backup retention
- **Cleanup** - Automatic old backup cleanup

### Configuration
```yaml
backup:
  enabled: true
  directory: "/backups"
  retentionDays: 30
  cleanupOld: true
  compressBackup: true
```

### Scheduling
- **Cron Expressions** - Flexible scheduling
- **Multiple Schedules** - Different backup frequencies
- **Error Handling** - Failed backup recovery
- **Monitoring** - Backup status monitoring

## 📨 Messaging System

### RabbitMQ Integration
- **Message Queuing** - Asynchronous message processing
- **Multiple Queues** - Email, backup, cleanup, notification
- **Reliability** - Persistent messages and acknowledgments
- **Scalability** - Horizontal scaling support

### Message Types
- **Email Messages** - Asynchronous email sending
- **Backup Messages** - Database backup operations
- **Cleanup Messages** - System cleanup tasks
- **Notification Messages** - User notifications

## 🔧 Configuration

### Organization Settings
```yaml
# Organization-specific configuration
organizations:
  maxMembers: 100
  allowPublic: false
  requireApproval: true
```

### Backup Settings
```yaml
# Backup configuration
backup:
  schedule: "0 2 * * *"  # Daily at 2 AM
  compress: true
  encrypt: false
```

## 🔗 Related Documentation

- [Architecture Documentation](../architecture/) - System architecture
- [Migration Documentation](../migration/) - RabbitMQ migration
- [Security Documentation](../security/) - Feature security
