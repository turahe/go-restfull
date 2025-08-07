# Go RESTful API - Hexagonal Architecture

This project demonstrates the implementation of Hexagonal Architecture (Ports and Adapters pattern) in a Go RESTful API using Fiber, PostgreSQL, and other modern technologies.

## 🏗️ Architecture Overview

The application follows Hexagonal Architecture principles with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                    Interface Layer                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │ HTTP Contr. │  │ gRPC Contr. │  │ CLI Commands│        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                  Application Layer                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │ User Service│  │ Post Service│  │ Auth Service│        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Domain Layer                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │   Entities  │  │ Repositories│  │   Services  │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                Infrastructure Layer                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │ PostgreSQL  │  │   Redis     │  │   Email     │        │
│  │ Repository  │  │   Cache     │  │   Service   │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────────────────────────────────────────────┘
```

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 13+
- Redis (optional, for caching)
- Docker & Docker Compose (optional)

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd go-restfull
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Setup environment**
   ```bash
   cp config/config.example.yaml config/config.yaml
   # Edit config/config.yaml with your database credentials
   ```

4. **Run database migrations**
   ```bash
   go run cmd/migrate.go
   ```

5. **Start the application**
   ```bash
   go run main.go
   ```

### Using Docker

```bash
# Start all services
docker-compose up -d

# Run migrations
docker-compose exec app go run cmd/migrate.go

# View logs
docker-compose logs -f app
```

## 📁 Project Structure

```
├── cmd/                    # Command line tools
├── config/                 # Configuration files
├── docs/                   # Documentation
├── internal/               # Internal application code
│   ├── domain/            # Domain layer (core business logic)
│   │   ├── entities/      # Domain entities
│   │   ├── repositories/  # Repository interfaces
│   │   └── services/      # Domain services
│   ├── application/       # Application layer (use cases)
│   │   ├── ports/         # Application service interfaces
│   │   └── services/      # Application service implementations
│   ├── infrastructure/    # Infrastructure layer (adapters)
│   │   ├── adapters/      # External service adapters
│   │   └── container/     # Dependency injection container
│   └── interfaces/        # Interface layer (HTTP, gRPC)
│       └── http/          # HTTP interface
│           ├── controllers/
│           ├── requests/
│           └── responses/
├── pkg/                   # Public packages
└── test/                  # Integration tests
```

## 🔧 Key Features

### Domain Layer
- **Rich Domain Models**: Entities with business logic and validation
- **Repository Pattern**: Clean data access interfaces
- **Domain Services**: Business logic that doesn't belong to entities

### Application Layer
- **Use Cases**: Application-specific business logic
- **Ports**: Interfaces for external dependencies
- **Orchestration**: Coordinates between domain entities

### Infrastructure Layer
- **Adapters**: Implementations of domain interfaces
- **Dependency Injection**: Clean wiring of components
- **External Services**: Database, email, caching, etc.

### Interface Layer
- **HTTP Controllers**: REST API endpoints
- **Request/Response DTOs**: Data transfer objects
- **Validation**: Input validation and sanitization

## 🛠️ API Endpoints

### Users
- `POST /api/v1/users` - Create a new user
- `GET /api/v1/users` - List users (with pagination)
- `GET /api/v1/users/:id` - Get user by ID
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user
- `PUT /api/v1/users/:id/password` - Change password

### Authentication
- `POST /api/v1/auth/login` - User login

### Health Check
- `GET /health` - Application health status

## 🧪 Testing

### Unit Tests
```bash
go test ./internal/domain/...
go test ./internal/application/...
```

### Integration Tests
```bash
go test ./test/...
```

### Run All Tests
```bash
go test ./...
```

## 📊 Database Schema

The application uses PostgreSQL with the following main tables:

- `users` - User accounts and profiles
- `posts` - Blog posts and content
- `settings` - User and application settings
- `media` - File uploads and media
- `tags` - Content tagging system
- `taxonomies` - Content categorization

## 🔒 Security Features

- **Password Hashing**: Bcrypt with configurable cost
- **JWT Authentication**: Stateless authentication
- **Input Validation**: Request validation and sanitization
- **CORS**: Configurable cross-origin resource sharing
- **Rate Limiting**: API rate limiting (configurable)

## 📈 Performance

- **Connection Pooling**: Efficient database connections
- **Caching**: Redis-based caching (optional)
- **Compression**: Response compression
- **Logging**: Structured logging with Zap

## 🚀 Deployment

### Production Build
```bash
go build -o bin/api main.go
```

### Docker Build
```bash
docker build -t go-restfull-api .
```

### Environment Variables
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=go_restfull
export DB_USER=postgres
export DB_PASSWORD=password
export JWT_SECRET=your-secret-key
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes following the hexagonal architecture principles
4. Add tests for new functionality
5. Submit a pull request

## 📚 Documentation

- [Hexagonal Architecture Guide](docs/HEXAGONAL_ARCHITECTURE.md)
- [API Documentation](docs/swagger.yaml)
- [JWT Authentication](docs/JWT_AUTHENTICATION.md)

## 🏆 Benefits of Hexagonal Architecture

### ✅ Maintainability
- Clear separation of concerns
- Easy to understand and modify
- Reduced coupling between components

### ✅ Testability
- Easy to mock dependencies
- Isolated unit tests
- Comprehensive test coverage

### ✅ Flexibility
- Easy to swap implementations
- Framework independence
- Multiple delivery mechanisms

### ✅ Scalability
- Modular design
- Easy to extend
- Support for microservices

## 🔮 Future Enhancements

- [ ] Event Sourcing
- [ ] CQRS (Command Query Responsibility Segregation)
- [ ] GraphQL API
- [ ] gRPC interface
- [ ] Message queues (RabbitMQ/Kafka)
- [ ] Kubernetes deployment
- [ ] Monitoring and observability
- [ ] API versioning

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Fiber](https://gofiber.io/) - Fast HTTP framework
- [PostgreSQL](https://www.postgresql.org/) - Reliable database
- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/) - Design pattern by Alistair Cockburn 