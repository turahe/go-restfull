# Database Seeding Feature

This document describes the database seeding feature that allows you to populate your database with initial data for development and testing purposes.

## Overview

The database seeding feature provides a systematic way to populate your database with sample data. It includes:

- **Seeder Management**: Tracks which seeders have been run to avoid duplicate data
- **Multiple Seeders**: Separate seeders for different entity types
- **Dependency Management**: Ensures seeders run in the correct order
- **Idempotent Operations**: Safe to run multiple times

## Available Commands

### 1. Seed Database
```bash
./webapi.exe seed
```
Populates the database with initial data including:
- Roles (super_admin, admin, editor, author, user)
- Users (superadmin, admin, editor, author, user)
- User-Role assignments
- Taxonomies (Technology, Programming, Business, etc.)
- Tags (Go, JavaScript, Python, React, etc.)
- Menus (Dashboard, Content Management, User Management, etc.)
- Menu-Role assignments
- Posts (sample blog posts)
- Content (About Us, Privacy Policy, etc.)
- Comments (sample comments on posts)
- Settings (site configuration)

### 2. Clear Seeded Data
```bash
./webapi.exe seed:flush
```
Removes all seeded data from the database and clears the seeder tracking table.

### 3. Check Seeding Status
```bash
./webapi.exe seed:status
```
Shows which seeders have been applied and when they were run.

## Seeder Architecture

### Seeder Interface
All seeders implement the `Seeder` interface:

```go
type Seeder interface {
    GetName() string
    Run(ctx context.Context, db *pgxpool.Pool) error
}
```

### Seeder Manager
The `SeederManager` handles:
- Tracking which seeders have been run
- Preventing duplicate execution
- Error handling and logging

### Seeder Order
Seeders are executed in the following order to respect dependencies:

1. **RoleSeeder** - Creates initial roles
2. **UserSeeder** - Creates initial users
3. **UserRoleSeeder** - Assigns roles to users
4. **TaxonomySeeder** - Creates content categories
5. **TagSeeder** - Creates content tags
6. **MenuSeeder** - Creates navigation menus
7. **MenuRoleSeeder** - Assigns menu access to roles
8. **PostSeeder** - Creates sample blog posts
9. **ContentSeeder** - Creates static content pages
10. **CommentSeeder** - Creates sample comments
11. **SettingSeeder** - Creates application settings

## Sample Data

### Users
The seeder creates the following users with verified email and phone:

| Username | Email | Password | Role |
|----------|-------|----------|------|
| superadmin | superadmin@example.com | SuperAdmin123! | super_admin |
| admin | admin@example.com | Admin123! | admin |
| editor | editor@example.com | Editor123! | editor |
| author | author@example.com | Author123! | author |
| user | user@example.com | User123! | user |

### Roles
- **super_admin**: Full system access
- **admin**: Management access
- **editor**: Content management access
- **author**: Content creation access
- **user**: Basic access

### Taxonomies
- Technology
  - Programming
    - Web Development
  - Mobile Development
- Business
  - Marketing

### Tags
- Go, JavaScript, Python, React, Vue.js
- Docker, Kubernetes, API, Microservices, Database

### Sample Posts
- "Getting Started with Go Programming"
- "Building RESTful APIs with Fiber"
- "Database Design Best Practices"
- "Microservices Architecture Patterns" (draft)
- "Docker and Kubernetes for Developers"

## Database Schema

The seeding system creates a `seeders` table to track execution:

```sql
CREATE TABLE IF NOT EXISTS seeders (
    id SERIAL PRIMARY KEY,
    seeder VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

## Error Handling

- **Duplicate Data**: Uses `ON CONFLICT` clauses to handle existing data
- **Dependencies**: Ensures proper order of execution
- **Rollback**: If any seeder fails, the process stops and reports the error
- **Logging**: Comprehensive logging for debugging

## Customization

### Adding New Seeders

1. Create a new seeder file in `internal/seeders/`
2. Implement the `Seeder` interface
3. Add the seeder to the list in `cmd/seed.go`

Example:
```go
// internal/seeders/custom_seeder.go
type CustomSeeder struct{}

func NewCustomSeeder() *CustomSeeder {
    return &CustomSeeder{}
}

func (cs *CustomSeeder) GetName() string {
    return "CustomSeeder"
}

func (cs *CustomSeeder) Run(ctx context.Context, db *pgxpool.Pool) error {
    // Your seeding logic here
    return nil
}
```

### Modifying Existing Seeders

Each seeder is self-contained and can be modified independently. Common modifications:

- **Data Volume**: Adjust the number of records created
- **Content**: Modify sample data to match your needs
- **Relationships**: Change how entities are related

## Best Practices

1. **Idempotency**: Always use `ON CONFLICT` clauses
2. **Dependencies**: Respect foreign key relationships
3. **Performance**: Use batch inserts for large datasets
4. **Testing**: Test seeders in isolation
5. **Documentation**: Keep seeder data documented

## Troubleshooting

### Common Issues

1. **Unique Constraint Violations**
   - Check if data already exists
   - Use `seed:flush` to clear existing data

2. **Foreign Key Violations**
   - Ensure seeders run in correct order
   - Check that referenced entities exist

3. **Column Type Mismatches**
   - Verify data types match database schema
   - Check migration files for exact column definitions

### Debugging

- Use `seed:status` to check which seeders have run
- Check logs for detailed error messages
- Verify database schema matches seeder expectations

## Integration with Development Workflow

### Development Environment
```bash
# Initial setup
./webapi.exe migrate
./webapi.exe seed

# Reset and reseed
./webapi.exe seed:flush
./webapi.exe seed
```

### Testing Environment
```bash
# Clean slate for tests
./webapi.exe seed:flush
./webapi.exe seed
```

### Production
- Seeders are typically not run in production
- Use migrations for production data changes
- Consider data migration scripts for production deployments

## Security Considerations

- **Passwords**: Sample passwords are for development only
- **Data**: Sample data should not contain sensitive information
- **Access**: Ensure proper access controls in production

## Future Enhancements

Potential improvements to the seeding system:

1. **Environment-specific Data**: Different data for dev/staging/prod
2. **Data Factories**: Generate realistic test data
3. **Incremental Seeding**: Add data without clearing existing data
4. **Data Validation**: Validate seeded data against business rules
5. **Performance Optimization**: Parallel seeding for independent seeders 