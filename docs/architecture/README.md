# Architecture Documentation

This section contains documentation related to the application's architecture and design patterns.

## 📋 Contents

### Core Architecture
- **[Hexagonal Architecture](./hexagonal-architecture.md)** - Detailed explanation of the hexagonal architecture pattern used in this application
- **[Hexagonal Readme](./hexagonal-readme.md)** - Quick reference guide for hexagonal architecture implementation
- **[Architecture Overview](./overview.md)** - High-level architecture overview and design decisions

### Database Design
- **[Database Design](./database-design.md)** - Entity Relationship Diagrams and database schema documentation
- **[Database Schema](./database-schema.mermaid)** - Mermaid diagram source for database schema
- **[Database Schema SVG](./database-schema.svg)** - Visual representation of the database schema

## 🏗️ Architecture Principles

### Hexagonal Architecture
The application follows the Hexagonal Architecture (also known as Ports and Adapters) pattern, which provides:

- **Separation of Concerns** - Clear boundaries between business logic and infrastructure
- **Testability** - Easy to test business logic in isolation
- **Flexibility** - Easy to swap implementations (e.g., different databases)
- **Maintainability** - Clear structure and dependencies

### Key Components
- **Domain Layer** - Core business logic and entities
- **Application Layer** - Use cases and application services
- **Infrastructure Layer** - External dependencies and adapters
- **Interface Layer** - HTTP controllers and API endpoints

## 📊 Database Design

The database design follows these principles:
- **Normalization** - Proper database normalization
- **Indexing** - Strategic indexing for performance
- **Constraints** - Data integrity through constraints
- **Relationships** - Clear entity relationships

## 🔗 Related Documentation

- [Development Documentation](../development/) - Implementation details
- [API Documentation](../api/) - API specifications
- [Security Documentation](../security/) - Security architecture
