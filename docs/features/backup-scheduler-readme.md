# Database Backup Scheduler

A comprehensive database backup solution for the Go RESTful API that provides automated PostgreSQL backups using `pg_dump`.

## Features

✅ **Automated Backups** - Scheduled database backups using cron expressions  
✅ **Retention Management** - Automatic cleanup of old backup files  
✅ **Backup Statistics** - Track backup history and storage usage  
✅ **Manual Triggers** - HTTP endpoints for manual backup operations  
✅ **Configurable** - Customizable backup directory, retention period, and schedule  
✅ **Error Handling** - Robust error handling and logging  
✅ **Testing** - Comprehensive test coverage  

## Quick Start

### 1. Configuration

Add backup configuration to your `config.yaml`:

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

### 2. Manual Backup

Create a backup manually via HTTP:

```bash
curl -X POST http://localhost:8000/api/v1/backup/create
```

### 3. Check Backup Statistics

```bash
curl -X GET http://localhost:8000/api/v1/backup/stats
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/backup/create` | Create a new backup |
| `POST` | `/api/v1/backup/create-custom` | Create backup with custom parameters |
| `GET` | `/api/v1/backup/stats` | Get backup statistics |
| `POST` | `/api/v1/backup/cleanup` | Manually trigger cleanup |

## Cron Examples

| Schedule | Description |
|----------|-------------|
| `"0 2 * * *"` | Daily at 2 AM |
| `"0 2 * * 0"` | Weekly on Sunday at 2 AM |
| `"0 2 1 * *"` | Monthly on the 1st at 2 AM |
| `"0 */6 * * *"` | Every 6 hours |

## File Structure

```
backups/
├── backup_my_db_2024-01-15_14-30-25.sql
├── backup_my_db_2024-01-16_14-30-25.sql
└── backup_my_db_2024-01-17_14-30-25.sql
```

## Requirements

- PostgreSQL client tools (`pg_dump`)
- Sufficient disk space
- Proper database permissions

## Testing

Run the backup service tests:

```bash
go test ./internal/domain/services -v
go test ./internal/job -v
```

## Monitoring

The system logs all backup operations:

```
INFO    Database backup job started
INFO    Database backup completed successfully    file_path=backups/backup_my_db_2024-01-15_14-30-25.sql    size_bytes=1048576    duration_seconds=45.2
INFO    Cleanup completed    deleted_files=3
```

## Security Notes

- Backup files contain sensitive data
- Ensure proper file permissions
- Consider encrypting backup files
- Regular security audits recommended

## Future Enhancements

- [ ] Backup compression
- [ ] Incremental backups  
- [ ] Backup encryption
- [ ] Cloud storage integration
- [ ] Backup verification
- [ ] Email notifications
- [ ] Backup restoration endpoints 