# Go RESTful API Documentation

Welcome to the comprehensive documentation for the Go RESTful API built with Hexagonal Architecture.

## 📚 Documentation Overview

This documentation is organized into several categories to help you find the information you need quickly:

### 🏗️ [Architecture](./architecture/)
- **Hexagonal Architecture** - Core architectural patterns and principles
- **Database Design** - ERD, schema, and database-related documentation
- **API Design** - Swagger documentation and API specifications

### 🔧 [Development](./development/)
- **Testing Strategy** - Comprehensive testing approach and implementation
- **Database Seeding** - Data seeding architecture and implementation
- **Pagination** - Pagination service setup and implementation
- **Health Endpoints** - Health check and monitoring endpoints

### 🔐 [Security](./security/)
- **JWT Authentication** - JWT implementation and usage
- **RBAC Implementation** - Role-Based Access Control system
- **Authorization** - Security and authorization patterns

### 🚀 [Deployment](./deployment/)
- **Docker Hub** - Docker Hub setup and deployment guides
- **Docker Configuration** - Multi-stage builds and containerization
- **Environment Setup** - Development, staging, and production environments

### 📊 [Features](./features/)
- **Organization Management** - Organization feature documentation
- **Backup Scheduler** - Database backup scheduling system
- **Messaging System** - RabbitMQ integration and message queuing

### 🔄 [Migration & Optimization](./migration/)
- **RabbitMQ Migration** - Migration from job service to RabbitMQ
- **Code Optimization** - Performance improvements and code cleanup

## 🚀 Quick Start

### Prerequisites
- Go 1.24.5+
- PostgreSQL
- Redis
- RabbitMQ (for messaging)

### Installation
```bash
# Clone the repository
git clone <repository-url>
cd go-restfull

# Install dependencies
go mod download

# Set up environment
cp config/config.example.yaml config/config.yaml
# Edit config/config.yaml with your settings

# Run migrations
go run main.go migrate

# Start the server
go run main.go
```

### Docker Setup
```bash
# Start all services
docker-compose up -d

# Run the application
docker-compose up app
```

## 📖 API Documentation

- **Swagger UI**: http://localhost:8000/swagger/
- **API Base URL**: http://localhost:8000/api/v1

## 🔧 Configuration

The application uses YAML configuration files:

- `config/config.yaml` - Main configuration
- `config/config.example.yaml` - Example configuration template
- `config/config.testing.yaml` - Testing configuration

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run integration tests
go test -tags=integration ./...
```

## 📊 Monitoring

- **Health Check**: `GET /api/v1/health`
- **Metrics**: Application metrics and monitoring
- **Logging**: Structured logging with Zap

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📝 License

This project is licensed under the MIT License.

---

## 📋 Documentation Index

### Architecture
- [Hexagonal Architecture](./architecture/hexagonal-architecture.md)
- [Database Design](./architecture/database-design.md)
- [API Documentation](./architecture/api-documentation.md)

### Development
- [Testing Strategy](./development/testing-strategy.md)
- [Database Seeding](./development/database-seeding.md)
- [Pagination Implementation](./development/pagination-implementation.md)
- [Health Endpoints](./development/health-endpoints.md)

### Security
- [JWT Authentication](./security/jwt-authentication.md)
- [RBAC Implementation](./security/rbac-implementation.md)

### Deployment
- [Docker Hub Guide](./deployment/docker-hub-guide.md)
- [Docker Configuration](./deployment/docker-configuration.md)

### Features
- [Organization Management](./features/organization-management.md)
- [Backup Scheduler](./features/backup-scheduler.md)
- [Messaging System](./features/messaging-system.md)

### Migration & Optimization
- [RabbitMQ Migration](./migration/rabbitmq-migration.md)
- [Code Optimization](./migration/code-optimization.md)

---

**Last Updated**: January 2025  
**Version**: 1.0.0
