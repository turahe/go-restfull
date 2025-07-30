# Docker Multi-Stage Builds

This project uses multi-stage Docker builds to create optimized images for different environments: development, staging, and production.

## Build Stages

### 1. Builder Stage
- **Purpose**: Common build stage for all environments
- **Base Image**: `golang:1.24.5-alpine`
- **Features**: 
  - Compiles the Go application with optimizations (`-ldflags="-w -s"`)
  - Installs build dependencies (git, gcc, g++)
  - Downloads Go modules

### 2. Development Stage
- **Purpose**: Hot reload development environment
- **Base Image**: `golang:1.24.5-alpine`
- **Features**:
  - Includes Air for hot reloading
  - Full source code with volume mounting
  - Development tools and dependencies
  - Health checks

### 3. Staging Stage
- **Purpose**: Testing environment
- **Base Image**: `alpine:latest`
- **Features**:
  - Minimal production-like environment
  - Non-root user for security
  - Health checks
  - Optimized binary

### 4. Production Stage
- **Purpose**: Production deployment
- **Base Image**: `alpine:latest`
- **Features**:
  - Minimal footprint
  - Non-root user for security
  - Health checks
  - Optimized binary
  - Production environment variables

## Building Images

### Using Makefile
```bash
# Build all environments
make docker-build

# Build specific environment
make docker-build-dev      # Development
make docker-build-staging  # Staging
make docker-build-prod     # Production
```

### Using Docker directly
```bash
# Development
docker build --target development -t go-restfull:dev .

# Staging
docker build --target staging -t go-restfull:staging .

# Production
docker build --target production -t go-restfull:prod .
```

## Running Containers

### Using Makefile
```bash
# Run production container
make docker-run

# Run development container with volume mounting
make docker-run-dev

# Run staging container
make docker-run-staging

# Run with external configuration file
make docker-run-with-config

# Run with external configuration via environment variable
make docker-run-with-config-env

# Run development with external configuration
make docker-run-dev-with-config

# Run development with external configuration via environment variable
make docker-run-dev-with-config-env
```

### Using Docker directly
```bash
# Production
docker run -p 8000:8000 go-restfull:prod

# Development with volume mounting
docker run -p 8000:8000 -v $(PWD):/app go-restfull:dev

# Staging
docker run -p 8000:8000 go-restfull:staging

# With external configuration file
docker run -p 8000:8000 -v $(pwd)/custom.yml:/app/config.yml go-restfull:prod

# With external configuration via environment variable
docker run -p 8000:8000 -v $(pwd)/custom.yml:/configs/user.yml -e CONFIG_PATH=/configs/user.yml go-restfull:prod
```

## Docker Compose

### Development Environment
```bash
docker-compose -f docker-compose.dev.yml up -d
```

### Staging Environment
```bash
docker-compose -f docker-compose.staging.yml up -d
```

### Production Environment
```bash
docker-compose -f docker-compose.prod.yml up -d
```

## Environment Variables

Each stage sets appropriate environment variables:

- **Development**: `ENV=development`
- **Staging**: `ENV=staging`
- **Production**: `ENV=production`

### External Configuration Support

The application supports external configuration files through:

1. **Volume mounting**: Mount your config file directly to `/app/config.yml`
2. **Environment variable**: Use `CONFIG_PATH` to specify the config file location

#### Configuration Priority:
1. `CONFIG_PATH` environment variable (highest priority)
2. Command line `--config` flag
3. Default `config/config.yaml` (lowest priority)

#### Example Usage:

```bash
# Method 1: Direct volume mount
docker run -p 8000:8000 -v $(pwd)/custom.yml:/app/config.yml go-restfull:prod

# Method 2: Environment variable with custom path
docker run -p 8000:8000 -v $(pwd)/custom.yml:/configs/user.yml -e CONFIG_PATH=/configs/user.yml go-restfull:prod

# Method 3: Command line flag
docker run -p 8000:8000 -v $(pwd)/custom.yml:/app/custom.yml go-restfull:prod --config=/app/custom.yml
```

## Security Features

### Production & Staging
- Non-root user (`appuser:appgroup`)
- Minimal base image (Alpine)
- Optimized binary with stripped symbols
- Health checks for monitoring

### Development
- Full development tools
- Hot reloading with Air
- Volume mounting for live code changes

## Health Checks

All environments include health checks:

```dockerfile
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8000/healthz || exit 1
```

## Image Sizes

- **Development**: ~500MB (includes Go toolchain and Air)
- **Staging**: ~50MB (minimal Alpine + binary)
- **Production**: ~50MB (minimal Alpine + binary)

## Best Practices

1. **Use specific targets**: Always specify the target stage when building
2. **Environment-specific configs**: Use different config files for each environment
3. **Security**: Production images run as non-root user
4. **Monitoring**: Health checks are included for all environments
5. **Optimization**: Production builds use `-ldflags="-w -s"` for smaller binaries

## Troubleshooting

### Build Issues
- Ensure Go version matches `go.mod` requirements
- Check that all source files are included (not in `.dockerignore`)
- Verify Swagger docs are generated before building

### Runtime Issues
- Check container logs: `docker logs <container-name>`
- Verify health checks: `docker inspect <container-name>`
- Ensure ports are properly exposed and mapped

### Development Issues
- Ensure volume mounting works correctly
- Check Air configuration in `.air.toml`
- Verify hot reloading is working 