# Seeder Architecture - Using Models/Entities

This document explains the improved seeder architecture that follows hexagonal architecture principles by using domain entities and repositories instead of raw SQL queries.

## Overview

The seeder system has been refactored to follow the same architectural patterns as the rest of the application, ensuring consistency, maintainability, and proper separation of concerns.

## Why Use Models/Entities in Seeders?

### **1. Architectural Consistency**
- **Follows Hexagonal Architecture**: Seeders now use the same patterns as controllers and services
- **Domain-Driven Design**: Uses domain entities that encapsulate business logic
- **Repository Pattern**: Leverages repository interfaces for data access
- **Service Layer**: Utilizes domain services for business operations

### **2. Benefits**

#### **✅ Type Safety**
```go
// Before: Raw SQL - Error prone
_, err = db.Exec(ctx, `
    INSERT INTO users (id, username, email, phone, password, email_verified_at, phone_verified_at, created_at, updated_at)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`, user.ID, user.Username, user.Email, user.Phone, hashedPassword, user.EmailVerifiedAt, user.PhoneVerifiedAt, user.CreatedAt, user.UpdatedAt)

// After: Entity + Repository - Type safe
user, err := entities.NewUser(username, email, phone, password)
err = userRepository.Create(ctx, user)
```

#### **✅ Business Logic Validation**
```go
// Domain entity validates business rules
user, err := entities.NewUser(username, email, phone, password)
if err != nil {
    return err // Validation failed
}

// Domain methods handle business logic
user.VerifyEmail()
user.VerifyPhone()
```

#### **✅ Repository Abstraction**
```go
// Repository handles data persistence details
err = userRepository.Create(ctx, user)
// Repository can handle:
// - Database-specific logic
// - Caching
// - Transaction management
// - Error handling
// - Duplicate handling
```

#### **✅ Maintainability**
- Changes to database schema only affect repository layer
- Business logic changes only affect domain entities
- Seeders remain focused on data creation
- Consistent error handling across the application

## Implementation Examples

### **User Seeder - Before vs After**

#### **❌ Before (Raw SQL Approach):**
```go
// Problems:
// 1. Direct database coupling
// 2. No type safety
// 3. No business validation
// 4. Hard to maintain
// 5. Inconsistent with application architecture

users := []struct {
    ID              uuid.UUID
    Username        string
    Email           string
    Phone           string
    Password        string
    EmailVerifiedAt *time.Time
    PhoneVerifiedAt *time.Time
    CreatedAt       time.Time
    UpdatedAt       time.Time
}{
    {
        ID:              uuid.New(),
        Username:        "superadmin",
        Email:           "superadmin@example.com",
        Phone:           "+1234567890",
        Password:        "SuperAdmin123!",
        EmailVerifiedAt: nil,
        PhoneVerifiedAt: nil,
        CreatedAt:       time.Now(),
        UpdatedAt:       time.Now(),
    },
    // ... more users
}

for _, user := range users {
    // Hash the password
    hashedPassword, err := passwordService.HashPassword(user.Password)
    if err != nil {
        return err
    }

    // Set verification times to current time
    now := time.Now()
    user.EmailVerifiedAt = &now
    user.PhoneVerifiedAt = &now

    _, err = db.Exec(ctx, `
        INSERT INTO users (id, username, email, phone, password, email_verified_at, phone_verified_at, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        ON CONFLICT (email) DO NOTHING
    `, user.ID, user.Username, user.Email, user.Phone, hashedPassword, user.EmailVerifiedAt, user.PhoneVerifiedAt, user.CreatedAt, user.UpdatedAt)

    if err != nil {
        return err
    }
}
```

#### **✅ After (Entity + Repository Approach):**
```go
// Benefits:
// 1. Follows hexagonal architecture
// 2. Type safe and validated
// 3. Business logic included
// 4. Easy to maintain
// 5. Consistent with application patterns

// Create password service for hashing passwords
passwordService := adapters.NewBcryptPasswordService()

// Create user repository using the adapter
userRepository := adapters.NewPostgresUserRepository(db, nil)

// Define users using the proper User entity
userData := []struct {
    username string
    email    string
    phone    string
    password string
}{
    {"superadmin", "superadmin@example.com", "+1234567890", "SuperAdmin123!"},
    {"admin", "admin@example.com", "+1234567891", "Admin123!"},
    {"editor", "editor@example.com", "+1234567892", "Editor123!"},
    {"author", "author@example.com", "+1234567893", "Author123!"},
    {"user", "user@example.com", "+1234567894", "User123!"},
}

for _, data := range userData {
    // Create user entity using domain constructor
    user, err := entities.NewUser(data.username, data.email, data.phone, data.password)
    if err != nil {
        return err // Validation failed
    }

    // Hash password using domain service
    hashedPassword, err := passwordService.HashPassword(user.Password)
    if err != nil {
        return err
    }
    user.Password = hashedPassword

    // Apply business logic
    user.VerifyEmail()
    user.VerifyPhone()

    // Use repository for persistence
    err = userRepository.Create(ctx, user)
    if err != nil {
        return err
    }
}
```

### **Role Seeder - Before vs After**

#### **❌ Before (Raw SQL Approach):**
```go
roles := []struct {
    ID          uuid.UUID
    Name        string
    Description string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}{
    {
        ID:          uuid.New(),
        Name:        "super_admin",
        Description: "Super Administrator with full system access",
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    },
    // ... more roles
}

for _, role := range roles {
    _, err := db.Exec(ctx, `
        INSERT INTO roles (id, name, slug, description, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (slug) DO NOTHING
    `, role.ID, role.Name, role.Name, role.Description, role.CreatedAt, role.UpdatedAt)

    if err != nil {
        return err
    }
}
```

#### **✅ After (Entity + Repository Approach):**
```go
// Create role repository using the adapter
roleRepository := adapters.NewPostgresRoleRepository(db, nil)

// Define roles using the proper Role entity
roleData := []struct {
    name        string
    slug        string
    description string
}{
    {"Super Administrator", "super_admin", "Super Administrator with full system access"},
    {"Administrator", "admin", "Administrator with management access"},
    {"Editor", "editor", "Editor with content management access"},
    {"Author", "author", "Author with content creation access"},
    {"User", "user", "Regular user with basic access"},
}

for _, data := range roleData {
    // Create role entity using domain constructor
    role, err := entities.NewRole(data.name, data.slug, data.description)
    if err != nil {
        return err
    }

    // Use repository for persistence
    err = roleRepository.Create(ctx, role)
    if err != nil {
        return err
    }
}
```

## Key Architectural Principles

### **1. Domain Entities**
- **Validation**: Entities validate business rules during creation
- **Business Logic**: Entities contain domain-specific methods
- **Immutability**: Entities are created with valid state
- **Encapsulation**: Business logic is encapsulated within entities

### **2. Repository Pattern**
- **Abstraction**: Hides database implementation details
- **Consistency**: Same interface used across the application
- **Error Handling**: Centralized error handling for data operations
- **Caching**: Repository can implement caching strategies

### **3. Service Layer**
- **Business Operations**: Services handle complex business logic
- **Password Hashing**: Domain services for security operations
- **Validation**: Services can perform additional validation
- **Transaction Management**: Services can manage transactions

### **4. Dependency Injection**
- **Testability**: Easy to mock dependencies for testing
- **Flexibility**: Can swap implementations without changing seeders
- **Configuration**: Dependencies can be configured externally

## Benefits Summary

### **✅ Maintainability**
- Changes to database schema only affect repository layer
- Business logic changes only affect domain entities
- Seeders remain focused on data creation
- Consistent patterns across the application

### **✅ Testability**
- Easy to mock repositories for unit testing
- Domain entities can be tested independently
- Services can be tested in isolation
- Integration tests can use real repositories

### **✅ Scalability**
- Repository pattern supports multiple database implementations
- Domain entities can be extended with new business logic
- Services can be enhanced with new functionality
- Seeders can be easily extended for new data types

### **✅ Consistency**
- Same patterns used across controllers, services, and seeders
- Consistent error handling throughout the application
- Uniform data access patterns
- Standardized business logic implementation

## Migration Guide

### **Steps to Refactor Existing Seeders:**

1. **Identify Domain Entity**: Find the corresponding domain entity for the seeder
2. **Check Repository Interface**: Verify the repository methods available
3. **Replace Raw SQL**: Replace direct database queries with repository calls
4. **Use Entity Constructors**: Use domain entity constructors for validation
5. **Apply Business Logic**: Use entity methods for business operations
6. **Test Thoroughly**: Ensure the refactored seeder works correctly

### **Example Migration:**
```go
// Step 1: Import domain entities and adapters
import (
    "webapi/internal/domain/entities"
    "webapi/internal/infrastructure/adapters"
)

// Step 2: Create repository instance
repository := adapters.NewPostgresEntityRepository(db, nil)

// Step 3: Use entity constructor
entity, err := entities.NewEntity(data)
if err != nil {
    return err
}

// Step 4: Apply business logic
entity.SomeBusinessMethod()

// Step 5: Use repository for persistence
err = repository.Create(ctx, entity)
```

## Conclusion

Using models/entities in seeders is the **correct architectural approach** because it:

1. **Maintains Consistency** with the rest of the application
2. **Ensures Type Safety** and validation
3. **Follows Domain-Driven Design** principles
4. **Improves Maintainability** and testability
5. **Provides Better Error Handling** and business logic

This approach aligns with hexagonal architecture principles and ensures that seeders are as robust and maintainable as the rest of the application. 