# Multi-Provider Storage System

This project now supports multiple storage providers for media files using the [QOR OSS library](https://github.com/qor/oss). This provides a unified interface for file operations across different storage backends.

## Supported Storage Providers

### 1. Local File System
- **Provider**: `local`
- **Use Case**: Development, testing, single-server deployments
- **Configuration**: Simple path configuration

### 2. MinIO (S3-Compatible)
- **Provider**: `minio`
- **Use Case**: Self-hosted object storage, development
- **Features**: S3-compatible API, easy local setup

### 3. AWS S3
- **Provider**: `s3`
- **Use Case**: Production cloud storage
- **Features**: High availability, global CDN, enterprise features

### 4. Azure Blob Storage
- **Provider**: `azure`
- **Use Case**: Microsoft Azure ecosystem
- **Features**: Enterprise integration, compliance features

### 5. Google Cloud Storage
- **Provider**: `gcs`
- **Use Case**: Google Cloud Platform ecosystem
- **Features**: Multi-region, lifecycle management

### 6. Alibaba Cloud OSS
- **Provider**: `alibaba`
- **Use Case**: Asia-Pacific region, Alibaba ecosystem
- **Features**: Regional optimization, CDN integration

### 7. Tencent Cloud COS
- **Provider**: `tencent`
- **Use Case**: China region, Tencent ecosystem
- **Features**: Regional compliance, cost optimization

### 8. Qiniu Cloud Kodo
- **Provider**: `qiniu`
- **Use Case**: China region, specialized media storage
- **Features**: Media optimization, CDN acceleration

## Configuration

### Storage Configuration File

Create a `config/storage.yaml` file:

```yaml
# Default storage provider
default: local

# Local filesystem storage
local:
  provider: local
  local_path: "./uploads"

# MinIO storage (S3-compatible)
minio:
  provider: minio
  access_key: "your-minio-access-key"
  secret_key: "your-minio-secret-key"
  region: "us-east-1"
  bucket: "your-bucket-name"
  endpoint: "http://localhost:9000"

# AWS S3 storage
s3:
  provider: s3
  access_key: "your-aws-access-key"
  secret_key: "your-aws-secret-key"
  region: "us-east-1"
  bucket: "your-s3-bucket"
  endpoint: "https://s3.amazonaws.com"
```

### Environment Variables

For production, use environment variables:

```bash
# Storage Provider
STORAGE_PROVIDER=s3

# AWS S3
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
AWS_REGION=us-east-1
AWS_S3_BUCKET=your-bucket

# MinIO
MINIO_ACCESS_KEY=your-minio-key
MINIO_SECRET_KEY=your-minio-secret
MINIO_ENDPOINT=http://localhost:9000
MINIO_BUCKET=your-bucket
```

## Usage

### Basic Usage

```go
package main

import (
    "github.com/turahe/go-restfull/internal/infrastructure/storage"
)

func main() {
    // Create storage loader
    loader, err := storage.NewStorageLoader("config/storage.yaml")
    if err != nil {
        log.Fatal(err)
    }

    // Load default storage service
    storageService, err := loader.LoadDefaultStorage()
    if err != nil {
        log.Fatal(err)
    }

    // Use storage service
    // ... upload, download, delete files
}
```

### Dynamic Provider Switching

```go
// Load specific provider
minioStorage, err := loader.LoadStorage("minio")
if err != nil {
    log.Fatal(err)
}

// Load all providers
allProviders, err := loader.LoadAllProviders()
if err != nil {
    log.Fatal(err)
}

// Use specific provider
s3Storage := allProviders["s3"]
```

### Media Service Integration

The media service now automatically uses the configured storage provider:

```go
// Media service automatically uses the storage service
mediaService := services.NewMediaService(mediaRepository, storageService)

// Upload file (automatically uses configured storage)
media, err := mediaService.UploadMedia(ctx, file, userID)
```

## File Operations

### Upload File

```go
// Upload file to storage
media, err := storageService.UploadFile(ctx, fileHeader, userID)
if err != nil {
    return err
}
```

### Download File

```go
// Get file as *os.File
file, err := storageService.GetFile(ctx, "/uploads/user/file.jpg")
if err != nil {
    return err
}
defer file.Close()

// Get file as stream
stream, err := storageService.GetFileStream(ctx, "/uploads/user/file.jpg")
if err != nil {
    return err
}
defer stream.Close()
```

### Delete File

```go
// Delete file from storage
err := storageService.DeleteFile(ctx, "/uploads/user/file.jpg")
if err != nil {
    return err
}
```

### Get File URL

```go
// Generate public URL for file
url, err := storageService.GetFileURL(ctx, "/uploads/user/file.jpg")
if err != nil {
    return err
}
```

### List Files

```go
// List files in directory
objects, err := storageService.ListFiles(ctx, "/uploads/user/")
if err != nil {
    return err
}

for _, obj := range objects {
    fmt.Printf("File: %s, Size: %d\n", obj.Path, obj.Size)
}
```

## Security Features

### File Validation
- File type validation through MIME type checking
- File size limits and validation
- Unique filename generation to prevent conflicts

### Access Control
- User ownership tracking for uploaded files
- Secure file path generation
- Prevention of path traversal attacks

### Storage Isolation
- User-specific upload directories
- Separate storage paths for different content types
- Configurable access permissions per provider

## Performance Optimization

### Local Storage
- Direct file system access
- No network latency
- Suitable for high-frequency operations

### Cloud Storage
- CDN integration for global access
- Automatic scaling
- Cost optimization through lifecycle policies

### MinIO
- S3-compatible API
- Local development and testing
- Easy migration to cloud providers

## Migration Guide

### From Local to Cloud Storage

1. **Update Configuration**
   ```yaml
   default: s3  # Change from local to s3
   ```

2. **Migrate Existing Files**
   ```go
   // Use storage service to migrate files
   localStorage := allProviders["local"]
   s3Storage := allProviders["s3"]
   
   // Copy files from local to S3
   // ... migration logic
   ```

3. **Update File URLs**
   - Update database records with new URLs
   - Implement URL rewriting if needed

### Between Cloud Providers

1. **Update Configuration**
   ```yaml
   default: azure  # Change from s3 to azure
   ```

2. **Cross-Provider Migration**
   - Use storage service interfaces
   - Implement migration scripts
   - Update file metadata

## Troubleshooting

### Common Issues

1. **Configuration Errors**
   - Verify YAML syntax
   - Check required fields for each provider
   - Validate credentials and endpoints

2. **Permission Issues**
   - Check file system permissions for local storage
   - Verify IAM roles for cloud providers
   - Ensure bucket/container access

3. **Network Issues**
   - Check endpoint URLs
   - Verify firewall settings
   - Test connectivity to storage services

### Debug Mode

Enable debug logging:

```go
// Set log level for storage operations
log.SetLevel(log.DebugLevel)
```

## Best Practices

### 1. Provider Selection
- **Development**: Use local storage
- **Testing**: Use MinIO for S3 compatibility
- **Production**: Use cloud providers for scalability

### 2. File Organization
- Organize files by user ID and date
- Use meaningful file naming conventions
- Implement file versioning if needed

### 3. Security
- Validate file types and sizes
- Implement user authentication
- Use secure file paths

### 4. Performance
- Use appropriate storage classes
- Implement caching strategies
- Monitor storage usage and costs

## Examples

### Complete Storage Setup

```go
package main

import (
    "log"
    "github.com/turahe/go-restfull/internal/infrastructure/storage"
)

func main() {
    // Initialize storage system
    loader, err := storage.NewStorageLoader("config/storage.yaml")
    if err != nil {
        log.Fatal("Failed to load storage configuration:", err)
    }

    // Validate storage configuration
    if err := loader.ValidateStorage("local"); err != nil {
        log.Fatal("Invalid local storage configuration:", err)
    }

    // Load default storage service
    storageService, err := loader.LoadDefaultStorage()
    if err != nil {
        log.Fatal("Failed to load default storage:", err)
    }

    log.Printf("Storage provider: %s", storageService.GetProvider())
    log.Printf("Is cloud storage: %t", storageService.IsCloudStorage())
}
```

### File Upload Handler

```go
func uploadHandler(c *fiber.Ctx) error {
    // Get uploaded file
    file, err := c.FormFile("file")
    if err != nil {
        return c.Status(400).JSON(fiber.Map{
            "status": "error",
            "message": "No file uploaded",
        })
    }

    // Get user ID from context
    userID := getUserIDFromContext(c)

    // Upload file using storage service
    media, err := storageService.UploadFile(c.Context(), file, userID)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "status": "error",
            "message": "Failed to upload file",
        })
    }

    return c.Status(201).JSON(media)
}
```

## Contributing

To add support for new storage providers:

1. Implement the `oss.StorageInterface`
2. Add provider configuration
3. Update the storage factory
4. Add tests and documentation

## License

This storage system is part of the Go RESTful API project and follows the same license terms.
