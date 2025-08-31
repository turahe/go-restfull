# MinIO Storage Troubleshooting Guide

## Issue: InvalidAccessKeyId Error

If you encounter the error `InvalidAccessKeyId: UnknownError` when trying to upload files, this guide will help you resolve it.

## Root Cause

The error occurs when there's a mismatch between:
1. The MinIO credentials configured in `docker-compose.yml`
2. The MinIO credentials configured in `config/storage.yaml`

## Solution

### 1. Verify MinIO Service is Running

First, ensure MinIO is running:

```bash
# Check if MinIO container is running
docker ps | grep minio

# Start MinIO if not running
docker-compose up minio -d
```

### 2. Check MinIO Credentials

The MinIO service in `docker-compose.yml` uses these credentials:
- **Access Key**: `minioadmin`
- **Secret Key**: `minioadmin`

### 3. Update Storage Configuration

Ensure your `config/storage.yaml` has matching credentials:

```yaml
minio:
  provider: minio
  access_key: "minioadmin"      # Must match docker-compose.yml
  secret_key: "minioadmin"      # Must match docker-compose.yml
  region: "us-east-1"
  bucket: "uploads"
  endpoint: "http://127.0.0.1:9000"
```

### 4. Test MinIO Connection

Use the provided test scripts:

```bash
# Test with shell script (Linux/Mac)
./scripts/test_minio.sh

# Test with Go script
cd scripts
go run test_minio.go
```

### 5. Verify Bucket Exists

Access the MinIO console at http://localhost:8900:
- Login with: `minioadmin` / `minioadmin`
- Check if the `uploads` bucket exists
- Create it if it doesn't exist

## Testing the Fix

### 1. Restart the Application

After updating the configuration:

```bash
# Restart the application
docker-compose restart app-prod

# Or rebuild and restart
docker-compose up --build app-prod -d
```

### 2. Test File Upload

Try uploading a file through the API:

```bash
curl -X POST http://localhost:8000/api/v1/media \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "file=@test-file.txt"
```

### 3. Check Health Endpoint

Verify storage health:

```bash
curl http://localhost:8000/healthz
```

Look for the storage service status in the response.

## Common Issues and Solutions

### Issue: "NoSuchBucket" Error
**Solution**: Create the `uploads` bucket in MinIO console

### Issue: "AccessDenied" Error
**Solution**: Check if the bucket has proper permissions

### Issue: Connection Refused
**Solution**: Ensure MinIO is running and accessible on port 9000

### Issue: Invalid Endpoint
**Solution**: Verify the endpoint URL in storage.yaml matches your MinIO setup

## Configuration Validation

The application now includes automatic validation:

1. **Storage Configuration**: Validates required fields
2. **Connection Testing**: Tests storage connectivity
3. **Bucket Access**: Verifies bucket existence and access
4. **Health Monitoring**: Includes storage status in health checks

## Monitoring

### Health Check Endpoint
- **URL**: `/healthz`
- **Storage Status**: Included in health response
- **Connection Test**: Performed automatically

### Logs
Look for these log entries:
- `Storage connection test failed`
- `Storage service upload failed`
- `Storage service is not accessible`

## Fallback Options

If MinIO continues to have issues, consider:

1. **Local Storage**: Switch to local filesystem storage
2. **AWS S3**: Use AWS S3 instead of MinIO
3. **Other Providers**: Azure Blob, Google Cloud Storage

## Support

If issues persist:
1. Check application logs for detailed error messages
2. Verify MinIO container logs: `docker-compose logs minio`
3. Test MinIO connectivity manually
4. Review storage configuration validation
