# Go RESTful API - Hexagonal Architecture

A comprehensive RESTful API built with Go, Fiber, and PostgreSQL using Hexagonal Architecture. This Docker image provides a production-ready, scalable API with authentication, authorization, and comprehensive features.

## 🚀 Features

- **Hexagonal Architecture** - Clean separation of concerns
- **JWT Authentication** - Secure token-based authentication
- **RBAC Authorization** - Role-based access control with Casbin
- **PostgreSQL Database** - Robust data persistence
- **Redis Caching** - High-performance caching layer
- **MinIO Integration** - Object storage capabilities
- **Email Service** - SMTP email functionality
- **Swagger Documentation** - Auto-generated API docs
- **Health Checks** - Container health monitoring
- **Multi-stage Builds** - Optimized for different environments

## 📋 Prerequisites

- Docker
- PostgreSQL (if not using docker-compose)
- Redis (if not using docker-compose)

## 🐳 Quick Start

### Using Docker Compose (Recommended)

```bash
# Clone the repository
git clone https://github.com/turahe/go-restfull.git
cd go-restfull

# Start all services
docker-compose -f docker-compose.prod.yml up -d
```

### Using Docker Run

```bash
# Run with default configuration
docker run -p 8000:8000 turahe/go-restfull:latest

# Run with external configuration
docker run -p 8000:8000 \
  -v $(pwd)/custom.yml:/configs/user.yml \
  -e CONFIG_PATH=/configs/user.yml \
  turahe/go-restfull:latest
```

## 🔧 Configuration

The application supports multiple configuration methods:

### 1. Environment Variable
```bash
docker run -p 8000:8000 \
  -v $(pwd)/config.yml:/configs/config.yml \
  -e CONFIG_PATH=/configs/config.yml \
  turahe/go-restfull:latest
```

### 2. Volume Mount
```bash
docker run -p 8000:8000 \
  -v $(pwd)/config.yml:/app/config.yml \
  turahe/go-restfull:latest
```

### 3. Command Line Flag
```bash
docker run -p 8000:8000 \
  -v $(pwd)/config.yml:/app/config.yml \
  turahe/go-restfull:latest --config=/app/config.yml
```

## 📁 Configuration Priority

1. `CONFIG_PATH` environment variable (highest priority)
2. Command line `--config` flag
3. Default `config/config.yaml` (lowest priority)

## 🏗️ Multi-Stage Builds

### Available Tags

- `turahe/go-restfull:latest` - Production optimized
- `turahe/go-restfull:dev` - Development with hot reload
- `turahe/go-restfull:staging` - Staging environment
- `turahe/go-restfull:prod` - Production environment

### Build Stages

```bash
# Development (with Air hot reload)
docker run -p 8000:8000 -v $(pwd):/app turahe/go-restfull:dev

# Staging
docker run -p 8000:8000 turahe/go-restfull:staging

# Production
docker run -p 8000:8000 turahe/go-restfull:prod
```

## 🔐 Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `CONFIG_PATH` | Path to configuration file | `config/config.yaml` |
| `ENV` | Environment (dev/staging/prod) | `production` |
| `TZ` | Timezone | `Asia/Jakarta` |

## 📊 Health Checks

The container includes health checks:

```bash
# Check container health
docker inspect --format='{{.State.Health.Status}}' <container_name>

# View health check logs
docker inspect --format='{{range .State.Health.Log}}{{.Output}}{{end}}' <container_name>
```

## 🔌 Ports

- `8000` - HTTP API server
- `5432` - PostgreSQL (if using docker-compose)
- `6379` - Redis (if using docker-compose)
- `9000` - MinIO API (if using docker-compose)
- `8900` - MinIO Console (if using docker-compose)
- `8025` - Mailpit Web UI (if using docker-compose)

## 📚 API Documentation

Once the container is running, access the Swagger documentation:

- **Swagger UI**: http://localhost:8000/swagger/
- **Health Check**: http://localhost:8000/healthz

## 🗄️ Database Setup

### Using Docker Compose
The docker-compose files include PostgreSQL and Redis services.

### Manual Setup
```bash
# PostgreSQL
docker run -d \
  --name postgres \
  -e POSTGRES_DB=my_db \
  -e POSTGRES_USER=my_user \
  -e POSTGRES_PASSWORD=secret \
  -p 5432:5432 \
  postgres:17

# Redis
docker run -d \
  --name redis \
  -p 6379:6379 \
  redis:alpine
```

## 🔧 Configuration Example

Create a `config.yml` file:

```yaml
env: production
app:
  name: "My API"
  nameSlug: "my-api"
  jwtSecret: "your-super-secret-jwt-key"
  accessTokenExpiration: 24

httpServer:
  port: 8000
  swaggerURL: "http://localhost:8000/swagger/"

postgres:
  host: "postgres"
  port: 5432
  database: "my_db"
  schema: "public"
  username: "my_user"
  password: "secret"

redis:
  - host: "redis"
    port: 6379
    password: ""
    database: 0

minio:
  enable: true
  endpoint: "minio:9000"
  accessKeyID: "minioadmin"
  accessKeySecret: "minioadmin"
  bucket: "my-bucket"
```

## 🚀 Docker Compose Examples

### Development
```bash
docker-compose -f docker-compose.dev.yml up -d
```

### Staging
```bash
docker-compose -f docker-compose.staging.yml up -d
```

### Production
```bash
docker-compose -f docker-compose.prod.yml up -d
```

## 🔍 Troubleshooting

### Container Won't Start
```bash
# Check logs
docker logs <container_name>

# Check configuration
docker exec -it <container_name> cat /app/config/config.yaml
```

### Database Connection Issues
- Ensure PostgreSQL is running and accessible
- Verify database credentials in configuration
- Check network connectivity between containers

### Redis Connection Issues
- Ensure Redis is running and accessible
- Verify Redis configuration in config file
- Check if Redis is required for your use case

## 📈 Monitoring

### Health Check Endpoint
```bash
curl http://localhost:8000/healthz
```

### Container Metrics
```bash
# Resource usage
docker stats <container_name>

# Container inspection
docker inspect <container_name>
```

## 🔒 Security

- **Non-root user**: Production images run as non-root user
- **Minimal base image**: Alpine Linux for smaller attack surface
- **Optimized binary**: Stripped symbols for security
- **Health checks**: Regular health monitoring

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📄 License

This project is licensed under the Apache 2.0 License - see the [LICENSE](https://github.com/turahe/go-restfull/blob/main/LICENSE) file for details.

## 🆘 Support

- **Issues**: [GitHub Issues](https://github.com/turahe/go-restfull/issues)
- **Documentation**: [GitHub Wiki](https://github.com/turahe/go-restfull/wiki)
- **Email**: wachid@outlook.com

## 🙏 Acknowledgments

- [Fiber](https://gofiber.io/) - Fast HTTP framework
- [PostgreSQL](https://www.postgresql.org/) - Database
- [Redis](https://redis.io/) - Caching
- [MinIO](https://min.io/) - Object storage
- [Casbin](https://casbin.org/) - Authorization

---

**Made with ❤️ by Nur Wachid** 