# Database Backup Scheduler

This document describes the database backup scheduler feature that automatically creates PostgreSQL database backups using `pg_dump`.

## Features

- **Automated Backups**: Scheduled database backups using cron expressions
- **Retention Management**: Automatic cleanup of old backup files
- **Backup Statistics**: Track backup history and storage usage
- **Manual Triggers**: HTTP endpoints for manual backup operations
- **Configurable**: Customizable backup directory, retention period, and schedule

## Configuration

### Backup Configuration

Add the following configuration to your `config.yaml`:

```yaml
backup:
  enabled: true
  directory: "backups"
  retentionDays: 30
  cleanupOld: true
  compressBackup: false

scheduler:
  timezone: "Asia/Jakarta"

schedules:
  - cron: "0 2 * * *"  # Daily at 2 AM
    job: "DatabaseBackup"
    isEnabled: true
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | false | Enable backup functionality |
| `directory` | string | "backups" | Directory to store backup files |
| `retentionDays` | int | 30 | Number of days to keep backup files |
| `cleanupOld` | bool | true | Automatically cleanup old backups |
| `compressBackup` | bool | false | Compress backup files (future feature) |

## Scheduled Jobs

The scheduler supports the following backup job types:

### DatabaseBackup

Creates a full database backup using `pg_dump`.

**Cron Examples:**
- `"0 2 * * *"` - Daily at 2 AM
- `"0 2 * * 0"` - Weekly on Sunday at 2 AM
- `"0 2 1 * *"` - Monthly on the 1st at 2 AM

## HTTP Endpoints

The backup system provides the following HTTP endpoints:

### Manual Backup Operations

- `POST /api/v1/backup/create` - Create a new backup
- `POST /api/v1/backup/create-custom` - Create backup with custom parameters
- `GET /api/v1/backup/stats` - Get backup statistics
- `POST /api/v1/backup/cleanup` - Manually trigger cleanup of old backups

### Example Usage

```bash
# Create a backup
curl -X POST http://localhost:8000/api/v1/backup/create

# Get backup statistics
curl -X GET http://localhost:8000/api/v1/backup/stats

# Create backup with custom parameters
curl -X POST http://localhost:8000/api/v1/backup/create-custom \
  -H "Content-Type: application/json" \
  -d '{
    "backup_dir": "/custom/backup/path",
    "retention_days": 60,
    "cleanup_old": true
  }'
```

## Backup File Format

Backup files are named using the following format:
```
backup_{database_name}_{timestamp}.sql
```

Example: `backup_my_db_2024-01-15_14-30-25.sql`

## Requirements

### System Requirements

- PostgreSQL client tools (`pg_dump`)
- Sufficient disk space for backup files
- Proper database connection credentials

### Database Permissions

The database user must have sufficient permissions to:
- Connect to the database
- Read all tables and data
- Execute `pg_dump` operations

## Monitoring

### Logs

The backup system logs all operations:

```
INFO    Database backup job started
INFO    Database backup completed successfully    file_path=backups/backup_my_db_2024-01-15_14-30-25.sql    size_bytes=1048576    duration_seconds=45.2
INFO    Cleanup completed    deleted_files=3
```

### Statistics

Backup statistics include:
- Total number of backups
- Total size of backup files
- Oldest and newest backup timestamps
- Backup directory location

## Troubleshooting

### Common Issues

1. **pg_dump not found**
   - Ensure PostgreSQL client tools are installed
   - Verify `pg_dump` is in the system PATH

2. **Permission denied**
   - Check database connection credentials
   - Verify backup directory permissions

3. **Insufficient disk space**
   - Monitor backup directory size
   - Adjust retention period if needed

4. **Backup timeout**
   - Increase timeout in backup service
   - Check database performance

### Debug Mode

Enable debug logging to troubleshoot issues:

```yaml
log:
  level: "debug"
```

## Security Considerations

- Backup files contain sensitive data
- Ensure backup directory has proper permissions
- Consider encrypting backup files
- Implement secure backup file transfer
- Regular security audits of backup access

## Performance Impact

- Backups can impact database performance
- Schedule backups during low-traffic periods
- Monitor backup duration and adjust timeout as needed
- Consider using `pg_dump` with `--jobs` for large databases

## Future Enhancements

- Backup compression
- Incremental backups
- Backup encryption
- Cloud storage integration
- Backup verification
- Email notifications
- Backup restoration endpoints 