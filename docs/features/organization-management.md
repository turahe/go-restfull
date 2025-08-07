# Organization Feature with Nested Set Hierarchy

## Overview

The Organization feature provides a comprehensive hierarchical organization management system using the Nested Set Model (NSM) for efficient tree operations. This implementation allows for unlimited depth organizations with fast querying capabilities for hierarchical relationships.

## Features

### Core Features
- **Hierarchical Organization Management**: Create, update, and manage organizations in a tree structure
- **Nested Set Model**: Efficient tree operations with left/right/depth values
- **Unlimited Depth**: Support for organizations with unlimited nesting levels
- **Status Management**: Active, inactive, and suspended organization states
- **Code-based Identification**: Unique organization codes for easy reference
- **Soft Delete**: Organizations are soft deleted to maintain referential integrity

### Hierarchy Operations
- **Get Root Organizations**: Retrieve all top-level organizations
- **Get Children**: Get direct child organizations
- **Get Descendants**: Get all descendant organizations (children, grandchildren, etc.)
- **Get Ancestors**: Get all ancestor organizations (parent, grandparent, etc.)
- **Get Siblings**: Get organizations at the same level
- **Get Path**: Get the complete path from root to a specific organization
- **Move Organizations**: Move organizations within the hierarchy
- **Tree View**: Get complete organization tree or subtree

### Search and Filtering
- **Full-text Search**: Search across name, description, code, and email
- **Status Filtering**: Filter organizations by status
- **Pagination**: Efficient pagination for large datasets
- **Statistics**: Get organization statistics (children count, descendants count)

## Database Schema

### Organizations Table

```sql
CREATE TABLE organizations (
    "id" UUID NOT NULL,
    "name" VARCHAR(255) NOT NULL,
    "description" TEXT,
    "code" VARCHAR(50) UNIQUE,
    "email" VARCHAR(255),
    "phone" VARCHAR(50),
    "address" TEXT,
    "website" VARCHAR(255),
    "logo_url" VARCHAR(500),
    "status" VARCHAR(20) DEFAULT 'active',
    "parent_id" UUID,
    "record_left" INTEGER NULL,
    "record_right" INTEGER NULL,
    "record_depth" INTEGER NULL,
    "record_ordering" INTEGER NULL,
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    "updated_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    "deleted_at" TIMESTAMP WITH TIME ZONE,
    CONSTRAINT "organizations_pkey" PRIMARY KEY ("id"),
    CONSTRAINT "organizations_parent_id_fkey" FOREIGN KEY ("parent_id") REFERENCES "organizations"("id") ON DELETE SET NULL,
    CONSTRAINT "organizations_status_check" CHECK ("status" IN ('active', 'inactive', 'suspended')),
    CONSTRAINT "organizations_record_left_right_check" CHECK ("record_left" < "record_right"),
    CONSTRAINT "organizations_record_ordering_check" CHECK ("record_ordering" >= 0)
);
```

### Nested Set Model Fields

- **record_left**: Left boundary value for nested set operations
- **record_right**: Right boundary value for nested set operations
- **record_depth**: Depth level in the hierarchy (0 for root)
- **record_ordering**: Ordering within siblings

## API Endpoints

### Basic CRUD Operations

#### Create Organization
```http
POST /api/v1/organizations
Content-Type: application/json

{
    "name": "New Organization",
    "description": "Organization description",
    "code": "ORG-001",
    "email": "info@org.com",
    "phone": "+1-555-0123",
    "address": "123 Main St",
    "website": "https://org.com",
    "logo_url": "https://org.com/logo.png",
    "parent_id": "uuid-of-parent-org"
}
```

#### Get Organization
```http
GET /api/v1/organizations/{id}
```

#### Get Organization by Code
```http
GET /api/v1/organizations/code/{code}
```

#### Update Organization
```http
PUT /api/v1/organizations/{id}
Content-Type: application/json

{
    "name": "Updated Organization",
    "description": "Updated description",
    "email": "updated@org.com"
}
```

#### Delete Organization
```http
DELETE /api/v1/organizations/{id}
```

#### List Organizations
```http
GET /api/v1/organizations?page=1&per_page=10&search=tech&status=active
```

### Hierarchy Operations

#### Get Root Organizations
```http
GET /api/v1/organizations/roots
```

#### Get Child Organizations
```http
GET /api/v1/organizations/{id}/children
```

#### Get Descendant Organizations
```http
GET /api/v1/organizations/{id}/descendants
```

#### Get Ancestor Organizations
```http
GET /api/v1/organizations/{id}/ancestors
```

#### Get Sibling Organizations
```http
GET /api/v1/organizations/{id}/siblings
```

#### Get Organization Path
```http
GET /api/v1/organizations/{id}/path
```

#### Get Organization Tree
```http
GET /api/v1/organizations/tree
```

#### Get Organization Subtree
```http
GET /api/v1/organizations/{id}/subtree
```

### Organization Management

#### Move Organization
```http
PUT /api/v1/organizations/{id}/move
Content-Type: application/json

{
    "new_parent_id": "uuid-of-new-parent"
}
```

#### Set Organization Status
```http
PUT /api/v1/organizations/{id}/status
Content-Type: application/json

{
    "status": "active"
}
```

### Search and Statistics

#### Search Organizations
```http
GET /api/v1/organizations/search?q=tech&page=1&per_page=10
```

#### Get Organization Statistics
```http
GET /api/v1/organizations/{id}/stats
```

## Usage Examples

### Creating a Hierarchical Organization Structure

```go
// Create root organization
rootOrg, err := organizationService.CreateOrganization(
    ctx,
    "TechCorp Global",
    "Global technology corporation",
    "TECH-CORP",
    "info@techcorp.com",
    "",
    "",
    "",
    "",
    nil, // No parent
)

// Create child organization
childOrg, err := organizationService.CreateOrganization(
    ctx,
    "Software Development",
    "Software development division",
    "SOFT-DEV",
    "dev@techcorp.com",
    "",
    "",
    "",
    "",
    &rootOrg.ID, // Set parent
)
```

### Moving Organizations

```go
// Move an organization to a new parent
err := organizationService.MoveOrganization(
    ctx,
    childOrgID,
    newParentID,
)
```

### Getting Hierarchy Information

```go
// Get all descendants of an organization
descendants, err := organizationService.GetDescendantOrganizations(ctx, orgID)

// Get the path from root to an organization
path, err := organizationService.GetOrganizationPath(ctx, orgID)

// Get organization statistics
childrenCount, err := organizationService.GetChildrenCount(ctx, orgID)
descendantsCount, err := organizationService.GetDescendantsCount(ctx, orgID)
```

## Nested Set Model Benefits

### Performance Advantages
- **Fast Ancestor Queries**: O(1) complexity for finding ancestors
- **Fast Descendant Queries**: O(1) complexity for finding descendants
- **Efficient Tree Traversal**: Single query to get entire subtree
- **Path Queries**: Easy to get complete path from root to any node

### Query Examples

```sql
-- Get all descendants
SELECT * FROM organizations 
WHERE record_left > (SELECT record_left FROM organizations WHERE id = ?)
  AND record_right < (SELECT record_right FROM organizations WHERE id = ?)

-- Get all ancestors
SELECT * FROM organizations 
WHERE record_left < (SELECT record_left FROM organizations WHERE id = ?)
  AND record_right > (SELECT record_right FROM organizations WHERE id = ?)

-- Get complete tree
SELECT * FROM organizations 
ORDER BY record_left
```

## Validation Rules

### Organization Code
- Must be unique across all organizations
- Can contain alphanumeric characters and hyphens only
- Cannot start or end with a hyphen
- Maximum length: 50 characters

### Organization Name
- Required field
- Maximum length: 255 characters

### Email
- Must be a valid email format if provided
- Maximum length: 255 characters

### Status
- Must be one of: `active`, `inactive`, `suspended`
- Default: `active`

## Error Handling

The API returns appropriate HTTP status codes and error messages:

- **400 Bad Request**: Invalid input data or validation errors
- **404 Not Found**: Organization not found
- **409 Conflict**: Duplicate organization code
- **422 Unprocessable Entity**: Business logic validation errors
- **500 Internal Server Error**: Server-side errors

## Security Considerations

- All endpoints (except public tree/roots) require JWT authentication
- Input validation prevents SQL injection and XSS attacks
- Soft delete maintains referential integrity
- Circular reference prevention in hierarchy operations

## Testing

### Unit Tests
- Entity validation tests
- Service layer business logic tests
- Repository layer data access tests

### Integration Tests
- API endpoint tests
- Database operation tests
- Hierarchy operation tests

### Performance Tests
- Large dataset query performance
- Tree traversal performance
- Concurrent operation tests

## Migration and Seeding

### Database Migration
```bash
# Run migrations
go run cmd/migrate.go

# Seed sample data
go run cmd/seed.go
```

### Sample Data
The seeder creates a sample organization hierarchy:
- TechCorp Global (root)
  - Software Development Division
    - Web Development Team
    - Mobile Development Team
  - Marketing Division
    - Digital Marketing Team
  - Sales Division
    - Enterprise Sales Team
    - SMB Sales Team
  - HR Division
    - Recruitment Team
    - Employee Relations Team

## Future Enhancements

### Planned Features
- **Bulk Operations**: Bulk create, update, and delete operations
- **Audit Trail**: Track changes to organization hierarchy
- **Versioning**: Support for organization structure versioning
- **Import/Export**: CSV/JSON import and export functionality
- **Advanced Search**: Full-text search with filters
- **Caching**: Redis caching for frequently accessed data
- **API Rate Limiting**: Rate limiting for API endpoints

### Performance Optimizations
- **Database Indexing**: Additional indexes for common queries
- **Query Optimization**: Optimized SQL queries for large datasets
- **Connection Pooling**: Efficient database connection management
- **Background Jobs**: Async processing for heavy operations

## Troubleshooting

### Common Issues

1. **Circular Reference Error**
   - Ensure parent-child relationships don't create cycles
   - Use validation methods before moving organizations

2. **Nested Set Values Inconsistency**
   - Run database integrity checks
   - Rebuild nested set values if needed

3. **Performance Issues with Large Trees**
   - Add appropriate database indexes
   - Consider pagination for large result sets
   - Use caching for frequently accessed data

### Debugging Tips

1. **Check Nested Set Values**
   ```sql
   SELECT id, name, record_left, record_right, record_depth 
   FROM organizations 
   ORDER BY record_left;
   ```

2. **Verify Hierarchy Integrity**
   ```sql
   SELECT COUNT(*) FROM organizations 
   WHERE record_left >= record_right;
   ```

3. **Check for Orphaned Records**
   ```sql
   SELECT * FROM organizations 
   WHERE parent_id IS NOT NULL 
   AND parent_id NOT IN (SELECT id FROM organizations);
   ```

## Contributing

When contributing to the organization feature:

1. Follow the existing code structure and patterns
2. Add comprehensive tests for new functionality
3. Update documentation for new features
4. Ensure backward compatibility
5. Follow the project's coding standards

## References

- [Nested Set Model Wikipedia](https://en.wikipedia.org/wiki/Nested_set_model)
- [Managing Hierarchical Data in MySQL](https://mikehillyer.com/articles/managing-hierarchical-data-in-mysql/)
- [Go Fiber Framework](https://gofiber.io/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)