# Phone Number Parsing Example

This document demonstrates how phone numbers are parsed from registration requests with country codes.

## Registration Request Structure

The registration request now accepts a single `phone` field that contains the full international phone number:

```json
{
  "username": "johndoe",
  "email": "john@example.com",
  "phone": "+1234567890",
  "password": "SecurePass123!",
  "confirm_password": "SecurePass123!"
}
```

## Phone Number Parsing Process

### 1. Input Validation
The system validates the phone number format using the enhanced `valueobjects.Phone` parser:

```go
// In RegisterRequest.Validate()
if _, err := valueobjects.NewPhone(r.Phone); err != nil {
    validator.ValidateCustom("phone", "invalid phone number: "+err.Error())
}
```

### 2. Country Code Detection
The parser automatically detects country codes from the phone number:

```go
// Examples of supported formats:
"+1234567890"     // US/Canada (country code: "1")
"+447911123456"   // UK (country code: "44")
"+49123456789"    // Germany (country code: "49")
"+8612345678901"  // China (country code: "86")
"+919876543210"   // India (country code: "91")
```

### 3. Normalization
The phone number is normalized to international format:

```go
// Input: "+1 234 567 8900" (with spaces)
// Output: "+12345678900" (normalized)

// Input: "+1-234-567-8900" (with dashes)
// Output: "+12345678900" (normalized)

// Input: "+1 (234) 567-8900" (with parentheses)
// Output: "+12345678900" (normalized)
```

### 4. Storage
The normalized phone number is stored in the database:

```go
// In AuthController.Register()
normalizedPhone, err := req.GetNormalizedPhone()
if err != nil {
    return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
        Status:  "error",
        Message: "Invalid phone number format",
    })
}

// Register user with normalized phone number
tokenPair, _, err := c.authService.RegisterUser(ctx.Context(), req.Username, req.Email, normalizedPhone, req.Password)
```

## API Examples

### Valid Registration Requests

#### US Phone Number
```bash
curl -X POST http://localhost:8000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "email": "john@example.com",
    "phone": "+1234567890",
    "password": "SecurePass123!",
    "confirm_password": "SecurePass123!"
  }'
```

#### UK Phone Number
```bash
curl -X POST http://localhost:8000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "janesmith",
    "email": "jane@example.com",
    "phone": "+447911123456",
    "password": "SecurePass123!",
    "confirm_password": "SecurePass123!"
  }'
```

#### German Phone Number with Formatting
```bash
curl -X POST http://localhost:8000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "hansmueller",
    "email": "hans@example.com",
    "phone": "+49 (123) 456-789",
    "password": "SecurePass123!",
    "confirm_password": "SecurePass123!"
  }'
```

### Invalid Registration Requests

#### Invalid Phone Number
```bash
curl -X POST http://localhost:8000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "phone": "123456",  # Too short
    "password": "SecurePass123!",
    "confirm_password": "SecurePass123!"
  }'
```

**Response:**
```json
{
  "status": "error",
  "message": "The given data was invalid.",
  "errors": [
    {
      "field": "phone",
      "message": "invalid phone number: phone number too short"
    }
  ]
}
```

#### Missing Country Code
```bash
curl -X POST http://localhost:8000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "phone": "123456789",  # No country code
    "password": "SecurePass123!",
    "confirm_password": "SecurePass123!"
  }'
```

**Response:**
```json
{
  "status": "error",
  "message": "The given data was invalid.",
  "errors": [
    {
      "field": "phone",
      "message": "invalid phone number: unable to parse country code"
    }
  ]
}
```

## Login with Phone Number

Users can also login using their phone number as the identity:

```bash
curl -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "identity": "+1234567890",
    "password": "SecurePass123!"
  }'
```

## Supported Country Codes

The system supports over 200 country codes including:

- `1` - US/Canada
- `44` - United Kingdom
- `33` - France
- `49` - Germany
- `39` - Italy
- `34` - Spain
- `31` - Netherlands
- `32` - Belgium
- `41` - Switzerland
- `43` - Austria
- `46` - Sweden
- `47` - Norway
- `45` - Denmark
- `48` - Poland
- `86` - China
- `81` - Japan
- `82` - South Korea
- `84` - Vietnam
- `91` - India
- `62` - Indonesia
- `63` - Philippines
- `65` - Singapore
- `66` - Thailand
- `60` - Malaysia

## Error Messages

Common phone number validation errors:

- `"phone cannot be empty"` - Phone number is missing
- `"phone number too short"` - Number is too short
- `"unable to parse country code"` - Country code not recognized
- `"national number must be between 7 and 15 digits"` - Invalid national number length
- `"national number must contain only digits"` - Invalid characters in national number

## Implementation Details

### Request Structure
```go
type RegisterRequest struct {
    Username        string `json:"username" validate:"required,min=3,max=32"`
    Email           string `json:"email" validate:"required,email"`
    Phone           string `json:"phone" validate:"required"` // Full phone number with country code
    Password        string `json:"password" validate:"required,min=8,max=32"`
    ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}
```

### Parsing Methods
```go
// ParsePhone parses the phone number and returns the phone value object
func (r *RegisterRequest) ParsePhone() (*valueobjects.Phone, error)

// GetNormalizedPhone returns the normalized phone number string
func (r *RegisterRequest) GetNormalizedPhone() (string, error)

// GetPhoneCountryCode returns the country code from the phone number
func (r *RegisterRequest) GetPhoneCountryCode() (string, error)

// GetPhoneNationalNumber returns the national number from the phone number
func (r *RegisterRequest) GetPhoneNationalNumber() (string, error)
```

### Controller Usage
```go
// In AuthController.Register()
normalizedPhone, err := req.GetNormalizedPhone()
if err != nil {
    return ctx.Status(http.StatusBadRequest).JSON(responses.ErrorResponse{
        Status:  "error",
        Message: "Invalid phone number format",
    })
}

// Register user with normalized phone number
tokenPair, _, err := c.authService.RegisterUser(ctx.Context(), req.Username, req.Email, normalizedPhone, req.Password)
```

## Best Practices

1. **Always use international format** with `+` prefix
2. **Include country code** for all phone numbers
3. **Validate on client side** before sending to API
4. **Handle errors gracefully** in user interface
5. **Display normalized format** to users after validation
6. **Use consistent formatting** across your application
