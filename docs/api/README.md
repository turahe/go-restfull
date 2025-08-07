# API Documentation

This section contains API specifications, Swagger documentation, and API-related resources.

## 📋 Contents

### API Specifications
- **[Swagger YAML](./swagger.yaml)** - OpenAPI specification in YAML format
- **[Swagger JSON](./swagger.json)** - OpenAPI specification in JSON format
- **[API Documentation](./docs.go)** - Generated API documentation

## 🔗 API Access

### Swagger UI
- **URL**: http://localhost:8000/swagger/
- **Description**: Interactive API documentation
- **Features**: Try out API endpoints directly

### API Base URL
- **Development**: http://localhost:8000/api/v1
- **Staging**: https://staging-api.example.com/api/v1
- **Production**: https://api.example.com/api/v1

## 📚 API Overview

### Authentication
All protected endpoints require JWT authentication:

```bash
# Get token
curl -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password"}'

# Use token
curl -X GET http://localhost:8000/api/v1/users \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Response Format
Standard response format:

```json
{
  "status": "success",
  "data": {
    // Response data
  },
  "message": "Operation completed successfully"
}
```

### Error Format
Standard error format:

```json
{
  "status": "error",
  "message": "Error description",
  "code": "ERROR_CODE"
}
```

## 🔐 Security

### Authentication
- **JWT Tokens** - Stateless authentication
- **Token Expiration** - Configurable token lifetime
- **Refresh Tokens** - Automatic token renewal

### Authorization
- **RBAC** - Role-Based Access Control
- **Resource Permissions** - Fine-grained access control
- **Organization Scoping** - Multi-tenant support

## 📊 Endpoints

### Public Endpoints
- `POST /api/v1/auth/login` - User authentication
- `POST /api/v1/auth/register` - User registration
- `GET /api/v1/health` - Health check

### Protected Endpoints
- `GET /api/v1/users` - List users
- `POST /api/v1/users` - Create user
- `GET /api/v1/users/{id}` - Get user details
- `PUT /api/v1/users/{id}` - Update user
- `DELETE /api/v1/users/{id}` - Delete user

### Organization Endpoints
- `GET /api/v1/organizations` - List organizations
- `POST /api/v1/organizations` - Create organization
- `GET /api/v1/organizations/{id}` - Get organization details

## 📄 Pagination

### Query Parameters
- `page` - Page number (default: 1)
- `page_size` - Items per page (default: 10)
- `sort` - Sort field
- `order` - Sort order (asc/desc)

### Response Format
```json
{
  "status": "success",
  "data": [...],
  "pagination": {
    "page": 1,
    "page_size": 10,
    "total": 100,
    "total_pages": 10
  }
}
```

## 🔧 Development

### Local Development
```bash
# Start the server
go run main.go

# Access Swagger UI
open http://localhost:8000/swagger/
```

### Testing APIs
```bash
# Using curl
curl -X GET http://localhost:8000/api/v1/health

# Using Swagger UI
# Visit http://localhost:8000/swagger/
```

## 🔗 Related Documentation

- [Architecture Documentation](../architecture/) - System architecture
- [Security Documentation](../security/) - API security
- [Development Documentation](../development/) - Development practices
