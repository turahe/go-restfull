# Casbin RBAC Implementation with Redis Adapter

This document describes the implementation of Role-Based Access Control (RBAC) using Casbin with Redis adapter in the Go RESTful API.

## Overview

The RBAC system provides fine-grained access control for API endpoints based on user roles and permissions. It uses Casbin as the authorization library with Redis for policy storage, ensuring high performance and scalability.

## Architecture

### Components

1. **Casbin RBAC Service** (`internal/infrastructure/adapters/casbin_rbac_service.go`)
   - Implements the RBAC service interface
   - Uses Casbin enforcer with Redis adapter
   - Handles policy management and permission checking

2. **RBAC Middleware** (`internal/router/middleware/rbac.go`)
   - `RBACMiddleware`: Enforces RBAC permissions for authenticated users
   - `OptionalRBACMiddleware`: Optional RBAC check that doesn't fail for unauthenticated users
   - `RequireRole`: Middleware that requires a specific role

3. **RBAC Controller** (`internal/interfaces/http/controllers/rbac_controller.go`)
   - Provides REST API endpoints for RBAC management
   - Allows adding/removing policies and user roles

4. **Configuration Files**
   - `config/rbac_model.conf`: Casbin authorization model
   - `config/rbac_policy.csv`: Initial policy rules
   - Configuration in `config/config.yaml`

## Configuration

### Casbin Configuration

Add the following to your configuration files:

```yaml
casbin:
  model: "config/rbac_model.conf"
  policy: "config/rbac_policy.csv"
  redis:
    host: "localhost"
    port: 6379
    password: ""
    db: 1
    key: "casbin_policy"
```

### RBAC Model (`config/rbac_model.conf`)

```conf
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && keyMatch2(r.obj, p.obj) && (r.act == p.act || p.act == "*")
```

### Initial Policy (`config/rbac_policy.csv`)

```csv
p, admin, /api/*, *
p, admin, /swagger/*, *
p, admin, /healthz, GET

p, user, /api/auth/*, *
p, user, /api/users/profile, GET
p, user, /api/users/profile, PUT
p, user, /api/posts, GET
p, user, /api/posts/*, GET
p, user, /api/comments, GET
p, user, /api/comments/*, GET

p, moderator, /api/*, GET
p, moderator, /api/posts, POST
p, moderator, /api/posts/*, PUT
p, moderator, /api/posts/*, DELETE
p, moderator, /api/comments, POST
p, moderator, /api/comments/*, PUT
p, moderator, /api/comments/*, DELETE
p, moderator, /api/users, GET
p, moderator, /api/users/*, GET

g, admin, moderator
g, moderator, user
```

## Usage

### 1. Adding RBAC Middleware to Routes

```go
// In your router setup
rbacMiddleware := middleware.RBACMiddleware(container.RBACService)

// Apply to specific routes
api.Get("/protected", rbacMiddleware, handler)

// Or apply to route groups
protected := api.Group("/admin", rbacMiddleware)
protected.Get("/users", adminHandler)
```

### 2. Requiring Specific Roles

```go
requireAdmin := middleware.RequireRole(container.RBACService, "admin")
adminRoutes := api.Group("/admin", requireAdmin)
```

### 3. Optional RBAC Check

```go
optionalRBAC := middleware.OptionalRBACMiddleware(container.RBACService)
api.Get("/public", optionalRBAC, handler)
```

### 4. Managing Policies via API

#### Get All Policies
```http
GET /api/rbac/policies
Authorization: Bearer <token>
```

#### Add Policy
```http
POST /api/rbac/policies
Authorization: Bearer <token>
Content-Type: application/json

{
  "subject": "moderator",
  "object": "/api/posts",
  "action": "POST"
}
```

#### Remove Policy
```http
DELETE /api/rbac/policies
Authorization: Bearer <token>
Content-Type: application/json

{
  "subject": "moderator",
  "object": "/api/posts",
  "action": "POST"
}
```

### 5. Managing User Roles

#### Get User Roles
```http
GET /api/rbac/users/{user_id}/roles
Authorization: Bearer <token>
```

#### Add Role to User
```http
POST /api/rbac/users/{user_id}/roles/{role}
Authorization: Bearer <token>
```

#### Remove Role from User
```http
DELETE /api/rbac/users/{user_id}/roles/{role}
Authorization: Bearer <token>
```

#### Get Users for Role
```http
GET /api/rbac/roles/{role}/users
Authorization: Bearer <token>
```

## Role Hierarchy

The system implements a role hierarchy:

- **admin**: Has access to everything
- **moderator**: Has access to most content management features
- **user**: Has basic access to read content and manage profile

Role inheritance: `admin` → `moderator` → `user`

## Permission Format

Permissions follow the format: `subject, object, action`

- **subject**: Role name (e.g., "admin", "moderator", "user")
- **object**: Resource path (e.g., "/api/posts", "/api/users/*")
- **action**: HTTP method (e.g., "GET", "POST", "PUT", "DELETE", "*")

## Path Matching

The system uses `keyMatch2` for path matching, supporting:

- `*`: Matches any sequence of characters
- `{id}`: Matches a parameter (e.g., `/api/posts/{id}` matches `/api/posts/123`)

## Integration with JWT

The RBAC middleware integrates with the existing JWT authentication:

1. JWT middleware extracts user information and stores it in context
2. RBAC middleware retrieves user ID from context
3. RBAC service fetches user roles from Casbin
4. Permission check is performed based on user roles

## Performance Considerations

- **Redis Storage**: Policies are stored in Redis for fast access
- **Caching**: User roles are cached in request context
- **Efficient Matching**: Uses optimized path matching algorithms
- **Connection Pooling**: Redis adapter supports connection pooling

## Security Features

- **Role-based Access**: Access control based on user roles
- **Hierarchical Roles**: Support for role inheritance
- **Fine-grained Permissions**: Control at endpoint level
- **Dynamic Policy Management**: Add/remove policies at runtime
- **Audit Trail**: All policy changes are logged

## Error Handling

The RBAC system provides comprehensive error handling:

- **Authentication Errors**: Returns 401 for unauthenticated requests
- **Authorization Errors**: Returns 403 for insufficient permissions
- **Service Errors**: Returns 500 for internal RBAC service errors
- **Validation Errors**: Returns 400 for invalid requests

## Monitoring and Logging

- All RBAC operations are logged
- Failed permission checks are logged with details
- Policy changes are tracked for audit purposes

## Testing

To test the RBAC implementation:

1. Start the application with Redis
2. Create users with different roles
3. Test API endpoints with different user roles
4. Verify permission enforcement works correctly

## Troubleshooting

### Common Issues

1. **Redis Connection**: Ensure Redis is running and accessible
2. **Policy Loading**: Check that policy files exist and are readable
3. **Role Assignment**: Verify users have roles assigned
4. **Path Matching**: Ensure paths in policies match actual API endpoints

### Debug Mode

Enable debug logging in the configuration to see detailed RBAC operations:

```yaml
log:
  level: "debug"
```

## Future Enhancements

- **Domain-based RBAC**: Support for multi-tenant applications
- **Time-based Policies**: Policies with expiration times
- **Conditional Policies**: Policies based on request attributes
- **Policy Templates**: Reusable policy patterns
- **Graphical Policy Editor**: Web interface for policy management 