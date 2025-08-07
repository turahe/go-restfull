# Hexagonal Architecture Implementation

This document explains the implementation of Hexagonal Architecture (also known as Ports and Adapters pattern) in the Go RESTful API project.

## Overview

Hexagonal Architecture is a software design pattern that promotes separation of concerns by organizing code into layers with clear boundaries. The core business logic is isolated from external concerns like databases, web frameworks, and external services.

## Architecture Layers

### 1. Domain Layer (Core)

The domain layer contains the core business logic and is completely independent of external concerns.

#### Entities (`internal/domain/entities/`)
- **User**: Core user entity with business logic
- **Post**: Core post entity with business logic

#### Repository Interfaces (`internal/domain/repositories/`)
- **UserRepository**: Defines the contract for user data access
- **PostRepository**: Defines the contract for post data access

#### Domain Services (`internal/domain/services/`)
- **PasswordService**: Interface for password operations
- **EmailService**: Interface for email operations

### 2. Application Layer

The application layer contains use cases and orchestrates domain entities.

#### Ports (`internal/application/ports/`)
- **UserService**: Application service interface for user operations
- **PostService**: Application service interface for post operations

#### Services (`internal/application/services/`)
- **UserService**: Implementation of user use cases
- **PostService**: Implementation of post use cases

### 3. Infrastructure Layer

The infrastructure layer contains adapters that implement the interfaces defined in the domain layer.

#### Adapters (`internal/infrastructure/adapters/`)
- **PostgresUserRepository**: PostgreSQL implementation of UserRepository
- **BcryptPasswordService**: Bcrypt implementation of PasswordService
- **SmtpEmailService**: SMTP implementation of EmailService

#### Container (`internal/infrastructure/container/`)
- **Container**: Dependency injection container that wires all components

### 4. Interface Layer

The interface layer handles external communication (HTTP, gRPC, etc.).

#### HTTP Controllers (`internal/interfaces/http/controllers/`)
- **UserController**: HTTP controller for user operations

#### Requests/Responses (`internal/interfaces/http/`)
- **Requests**: DTOs for incoming HTTP requests
- **Responses**: DTOs for HTTP responses

## Directory Structure

```
internal/
├── domain/
│   ├── entities/
│   │   ├── user.go
│   │   └── post.go
│   ├── repositories/
│   │   ├── user_repository.go
│   │   └── post_repository.go
│   └── services/
│       ├── password_service.go
│       └── email_service.go
├── application/
│   ├── ports/
│   │   ├── user_service.go
│   │   └── post_service.go
│   └── services/
│       ├── user_service.go
│       └── post_service.go
├── infrastructure/
│   ├── adapters/
│   │   ├── user_repository.go
│   │   ├── password_service.go
│   │   └── email_service.go
│   └── container/
│       └── container.go
└── interfaces/
    └── http/
        ├── controllers/
        │   └── user_controller.go
        ├── requests/
        │   └── user_requests.go
        └── responses/
            └── common_responses.go
```

## Key Benefits

### 1. Separation of Concerns
- Business logic is isolated from infrastructure concerns
- Each layer has a single responsibility
- Easy to understand and maintain

### 2. Testability
- Domain logic can be tested without external dependencies
- Easy to mock interfaces for unit testing
- Integration tests can focus on specific layers

### 3. Flexibility
- Easy to swap implementations (e.g., PostgreSQL to MongoDB)
- Can add new interfaces without changing existing code
- Supports multiple delivery mechanisms (HTTP, gRPC, CLI)

### 4. Independence
- Domain logic doesn't depend on frameworks
- Business rules are framework-agnostic
- Easy to migrate to different technologies

## Dependency Flow

```
HTTP Controller → Application Service → Domain Entity
     ↓                    ↓                    ↓
Infrastructure ← Repository Interface ← Domain Repository
```

## Example Usage

### Creating a User

1. **HTTP Request** → `UserController.CreateUser()`
2. **Controller** → `UserService.CreateUser()`
3. **Service** → `UserRepository.Create()` + `EmailService.SendWelcomeEmail()`
4. **Repository** → PostgreSQL database
5. **Email Service** → SMTP server

### Flow Diagram

```
[HTTP Request] → [Controller] → [Application Service] → [Domain Entity]
                                                              ↓
[HTTP Response] ← [Controller] ← [Application Service] ← [Repository] → [Database]
```

## Testing Strategy

### Unit Tests
- Test domain entities independently
- Mock repository interfaces
- Test application services with mocked dependencies

### Integration Tests
- Test repository implementations with test database
- Test HTTP controllers with mocked services
- Test complete use cases

### End-to-End Tests
- Test complete API endpoints
- Use test containers for external dependencies
- Validate business workflows

## Migration Guide

To migrate from the current architecture to Hexagonal Architecture:

1. **Extract Domain Entities**: Move business logic to domain entities
2. **Define Interfaces**: Create repository and service interfaces
3. **Implement Adapters**: Create infrastructure implementations
4. **Update Controllers**: Refactor controllers to use application services
5. **Setup DI Container**: Wire all dependencies together

## Best Practices

1. **Dependency Inversion**: Depend on abstractions, not concretions
2. **Single Responsibility**: Each class/module has one reason to change
3. **Interface Segregation**: Keep interfaces small and focused
4. **Dependency Injection**: Use constructor injection for dependencies
5. **Error Handling**: Handle errors at appropriate layers
6. **Validation**: Validate input at the interface layer
7. **Logging**: Log at infrastructure and interface layers

## Configuration

The hexagonal architecture is configured through the dependency injection container in `internal/infrastructure/container/container.go`. This file wires all the dependencies together and can be easily modified to swap implementations.

## Future Enhancements

1. **Event Sourcing**: Add domain events for better decoupling
2. **CQRS**: Separate read and write operations
3. **Microservices**: Split into multiple services
4. **GraphQL**: Add GraphQL interface
5. **gRPC**: Add gRPC interface for internal communication

## Conclusion

The hexagonal architecture provides a clean, maintainable, and testable structure for the Go RESTful API. It separates concerns effectively and makes the codebase more flexible for future changes. 