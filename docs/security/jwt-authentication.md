# JWT Authentication with Refresh Tokens

This project implements a comprehensive JWT (JSON Web Token) authentication system with refresh tokens for enhanced security.

## Features

- **Access Tokens**: Short-lived tokens (15 minutes) for API access
- **Refresh Tokens**: Long-lived tokens (7 days) for token renewal
- **Token Validation**: Secure token validation with proper error handling
- **Middleware Protection**: JWT middleware for protecting routes
- **Swagger Integration**: JWT authentication documented in Swagger UI

## Configuration

Add the JWT secret to your configuration file:

```yaml
app:
  name: "Your App Name"
  nameSlug: "your-app"
  jwtSecret: "your-super-secret-jwt-key-here-make-it-long-and-secure"
```

## API Endpoints

### Public Endpoints (No Authentication Required)

#### 1. Register User
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "john_doe",
  "email": "john@example.com",
  "phone": "1234567890",
  "password": "securepassword",
  "confirm_password": "securepassword"
}
```

#### 2. Login User
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "john_doe",
  "password": "securepassword"
}
```

**Response:**
```json
{
  "response_code": 200,
  "response_message": "Login successful",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 900
  }
}
```

#### 3. Refresh Access Token
```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### 4. Forget Password (Send OTP)
```http
POST /api/v1/auth/forget-password
Content-Type: application/json

{
  "email": "john@example.com"
}
```

### Protected Endpoints (Authentication Required)

#### 5. Logout User
```http
POST /api/v1/auth/logout
Authorization: Bearer <access_token>
```

## Using Protected Endpoints

For all protected endpoints, include the access token in the Authorization header:

```http
Authorization: Bearer <your_access_token>
```

### Example Protected Request
```http
GET /api/v1/users
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

## Token Structure

### Access Token Claims
```json
{
  "user_id": "uuid",
  "username": "john_doe",
  "email": "john@example.com",
  "type": "access",
  "exp": 1640995200,
  "iat": 1640994300,
  "nbf": 1640994300,
  "iss": "Your App Name",
  "sub": "uuid",
  "jti": "uuid"
}
```

### Refresh Token Claims
```json
{
  "user_id": "uuid",
  "username": "john_doe",
  "email": "john@example.com",
  "type": "refresh",
  "exp": 1641600000,
  "iat": 1640994300,
  "nbf": 1640994300,
  "iss": "Your App Name",
  "sub": "uuid",
  "jti": "uuid"
}
```

## Security Features

1. **Token Type Validation**: Access tokens can only be used for access, refresh tokens only for refresh
2. **Short-lived Access Tokens**: 15-minute expiration reduces risk of token theft
3. **Long-lived Refresh Tokens**: 7-day expiration for user convenience
4. **Secure Signing**: HMAC-SHA256 signing algorithm
5. **Proper Claims**: Standard JWT claims with custom user information
6. **Middleware Protection**: Automatic token validation on protected routes

## Error Responses

### Invalid Token
```json
{
  "message": "Invalid or expired token",
  "error": "token is expired"
}
```

### Missing Authorization Header
```json
{
  "message": "Authorization header is required"
}
```

### Invalid Authorization Format
```json
{
  "message": "Invalid authorization header format"
}
```

## Implementation Details

### Token Generation
- Uses `github.com/golang-jwt/jwt/v5` library
- HMAC-SHA256 signing algorithm
- Configurable expiration times
- UUID-based user identification

### Middleware
- `JWTAuth()`: Required authentication middleware
- `OptionalJWTAuth()`: Optional authentication middleware
- Automatic user context injection
- Proper error handling

### Route Protection
Protected routes are automatically secured with JWT middleware:
- User management endpoints
- Media management endpoints
- Settings management endpoints
- Queue management endpoints

## Best Practices

1. **Store tokens securely**: Use secure storage (not localStorage) for production
2. **Handle token refresh**: Implement automatic token refresh before expiration
3. **Logout properly**: Clear tokens on logout
4. **Use HTTPS**: Always use HTTPS in production
5. **Rotate secrets**: Regularly rotate JWT secrets
6. **Monitor usage**: Log and monitor token usage for security

## Testing

The JWT functionality can be tested using the provided test files:
- `test/jwt_test.go`: Comprehensive JWT tests
- `test/jwt_simple_test.go`: Simple JWT tests

Run tests with:
```bash
go test ./test/ -v
``` 