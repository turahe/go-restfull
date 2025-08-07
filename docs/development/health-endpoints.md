# Health Endpoints Documentation

This document describes the health check endpoints available in the Go RESTful API.

## Overview

The API provides two health check endpoints:
1. **Simple Health Check** (`/health`) - Basic application status
2. **Comprehensive Health Check** (`/healthz`) - Detailed service health status

## Endpoints

### 1. Simple Health Check

**Endpoint:** `GET /health`

**Description:** Basic health check that returns a simple status response.

**Response:**
```json
{
  "status": "healthy",
  "message": "Server is running"
}
```

**Use Cases:**
- Load balancer health checks
- Basic application status monitoring
- Quick availability checks

### 2. Comprehensive Health Check

**Endpoint:** `GET /healthz`

**Description:** Comprehensive health check that monitors all services and dependencies.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-07-27T23:47:39+07:00",
  "environment": "dev",
  "version": "1.0.0",
  "services": [
    {
      "service": "postgresql",
      "status": "healthy",
      "message": "Connected successfully",
      "timestamp": "2025-07-27T23:47:39+07:00"
    },
    {
      "service": "redis",
      "status": "healthy",
      "message": "Connected successfully",
      "timestamp": "2025-07-27T23:47:39+07:00"
    },
    {
      "service": "application",
      "status": "healthy",
      "message": "Application is running",
      "timestamp": "2025-07-27T23:47:39+07:00"
    },
    {
      "service": "configuration",
      "status": "healthy",
      "message": "Configuration loaded successfully",
      "timestamp": "2025-07-27T23:47:39+07:00"
    },
    {
      "service": "rbac",
      "status": "healthy",
      "message": "RBAC service configured",
      "timestamp": "2025-07-27T23:47:39+07:00"
    },
    {
      "service": "job_queue",
      "status": "healthy",
      "message": "Job queue system accessible",
      "timestamp": "2025-07-27T23:47:39+07:00"
    },
    {
      "service": "email",
      "status": "healthy",
      "message": "Email service configured",
      "timestamp": "2025-07-27T23:47:39+07:00"
    }
  ]
}
```

**HTTP Status Codes:**
- `200 OK` - All services are healthy
- `503 Service Unavailable` - One or more services are unhealthy

**Services Monitored:**

#### Core Services
- **postgresql** - Database connection and ping test
- **redis** - Redis connection and ping test
- **application** - Application runtime status
- **configuration** - Configuration loading status

#### Optional Services
- **minio** - MinIO object storage (only if enabled in config)
- **rbac** - Role-Based Access Control service
- **job_queue** - Background job processing system
- **email** - Email service configuration
- **sentry** - Error tracking service (only if configured)

**Use Cases:**
- Comprehensive system monitoring
- DevOps health checks
- Service dependency monitoring
- Troubleshooting system issues
- Kubernetes liveness/readiness probes

## Implementation Details

### Health Check Logic

The comprehensive health check performs the following tests:

1. **Database Health Check:**
   - Verifies PostgreSQL connection pool exists
   - Performs a ping test to ensure connectivity
   - Returns detailed error messages if connection fails

2. **Redis Health Check:**
   - Verifies Redis client exists
   - Performs a ping test to ensure connectivity
   - Returns detailed error messages if connection fails

3. **MinIO Health Check:**
   - Only runs if MinIO is enabled in configuration
   - Verifies MinIO client exists
   - Tests bucket listing functionality
   - Returns detailed error messages if connection fails

4. **Application Services:**
   - Checks application runtime status
   - Verifies configuration loading
   - Monitors service availability

5. **RBAC Service:**
   - Verifies RBAC configuration exists
   - Checks Casbin model and policy files

6. **Job Queue:**
   - Verifies jobs table accessibility
   - Tests database connectivity for job processing

7. **Email Service:**
   - Checks email configuration
   - Returns warning status if not configured

8. **Sentry:**
   - Only runs if Sentry DSN is configured
   - Verifies error tracking service configuration

### Error Handling

- **Timeout:** Health checks have a 10-second timeout to prevent hanging
- **Graceful Degradation:** Individual service failures don't crash the entire health check
- **Detailed Messages:** Each service provides specific error messages for troubleshooting
- **Status Aggregation:** Overall status is determined by critical service health

### Performance Considerations

- **Fast Response:** Health checks are designed to respond quickly
- **Minimal Impact:** Tests use lightweight operations (ping, simple queries)
- **Caching:** No caching to ensure real-time status
- **Timeout Protection:** Prevents health checks from blocking the application

## Usage Examples

### Basic Health Check
```bash
curl -X GET http://localhost:8000/health
```

### Comprehensive Health Check
```bash
curl -X GET http://localhost:8000/healthz
```

### Health Check with Status Code
```bash
curl -X GET http://localhost:8000/healthz -w "\nHTTP Status: %{http_code}\n"
```

### Monitoring Integration
```bash
# Check if all services are healthy
curl -s http://localhost:8000/healthz | jq '.status == "healthy"'
```

## Configuration

Health check behavior can be influenced by:

1. **Service Configuration:** Enable/disable services in `config.yaml`
2. **Timeout Settings:** Adjust timeout in the health check implementation
3. **Environment Variables:** Control service availability through environment

## Troubleshooting

### Common Issues

1. **Database Connection Failed:**
   - Check PostgreSQL service status
   - Verify connection credentials
   - Ensure database is accessible

2. **Redis Connection Failed:**
   - Check Redis service status
   - Verify Redis configuration
   - Ensure Redis is accessible

3. **MinIO Connection Failed:**
   - Check MinIO service status
   - Verify MinIO configuration
   - Ensure MinIO is accessible

4. **Configuration Issues:**
   - Verify `config.yaml` file exists
   - Check configuration syntax
   - Ensure required fields are present

### Debug Mode

For detailed debugging, check the application logs for health check related messages.

## Best Practices

1. **Use `/health` for load balancers** - Fast, simple checks
2. **Use `/healthz` for monitoring** - Comprehensive service status
3. **Set appropriate timeouts** - Don't let health checks hang
4. **Monitor response times** - Health checks should be fast
5. **Log health check failures** - For troubleshooting and alerting
6. **Use health checks in CI/CD** - Ensure services are ready before deployment 