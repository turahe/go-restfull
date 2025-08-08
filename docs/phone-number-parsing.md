# Phone Number Parsing with Country Codes

This document describes the enhanced phone number parsing functionality that supports international phone numbers with country codes.

## Overview

The phone number parsing system now supports:
- International format with `+` prefix
- Country code detection and validation
- National number extraction
- Various input formats (spaces, dashes, parentheses)

## Supported Formats

### International Format (Recommended)
```
+[country code][national number]
```

Examples:
- `+1234567890` (US/Canada)
- `+447911123456` (UK)
- `+49123456789` (Germany)
- `+8612345678901` (China)
- `+919876543210` (India)

### National Format (Fallback)
```
[country code][national number]
```

Examples:
- `1234567890` (US/Canada)
- `447911123456` (UK)
- `49123456789` (Germany)

### Formatted Input
The system also accepts formatted phone numbers:
- `+1 234 567 8900` (with spaces)
- `+1-234-567-8900` (with dashes)
- `+1 (234) 567-8900` (with parentheses)

## API Usage

### Registration Request
```json
{
  "username": "johndoe",
  "email": "john@example.com",
  "phone": "+1234567890",
  "password": "SecurePass123!",
  "confirm_password": "SecurePass123!"
}
```

### Login Request (with Identity)
```json
{
  "identity": "john@example.com",
  "password": "SecurePass123!"
}
```

## Supported Country Codes

The system supports a comprehensive list of country codes including:

### Major Countries
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

### Complete List
The system supports over 200 country codes including all major countries and territories.

## Validation Rules

1. **Country Code**: Must be a valid international country code
2. **National Number**: Must be 7-15 digits
3. **Total Length**: Must be 10-15 digits (including country code)
4. **Format**: Must contain only digits and the `+` prefix

## Error Messages

- `"phone cannot be empty"` - Phone number is missing
- `"invalid phone number format"` - Invalid format
- `"phone number too short"` - Number is too short
- `"unable to parse country code"` - Country code not recognized
- `"national number must be between 7 and 15 digits"` - Invalid national number length
- `"national number must contain only digits"` - Invalid characters in national number

## Implementation Details

### Phone Value Object
```go
type Phone struct {
    value         string
    countryCode   string
    nationalNumber string
}
```

### Methods
- `String()` - Returns normalized international format
- `Value()` - Returns the full phone number
- `CountryCode()` - Returns the country code
- `NationalNumber()` - Returns the national number
- `Equals(other Phone)` - Compares two phone numbers

### Parsing Logic
1. **Clean Input**: Remove all non-digit characters except `+`
2. **Detect Format**: Check if number starts with `+`
3. **Parse Country Code**: Match against known country codes
4. **Validate Components**: Ensure proper lengths and formats
5. **Normalize**: Return standardized international format

## Testing

Run the phone number tests:
```bash
go test ./internal/domain/valueobjects -v
```

## Examples

### Valid Phone Numbers
```go
// US number
phone, err := NewPhone("+1234567890")
// Result: countryCode="1", nationalNumber="234567890"

// UK number
phone, err := NewPhone("+447911123456")
// Result: countryCode="44", nationalNumber="7911123456"

// German number with formatting
phone, err := NewPhone("+49 (123) 456-789")
// Result: countryCode="49", nationalNumber="123456789"
```

### Invalid Phone Numbers
```go
// Too short
phone, err := NewPhone("+123456")
// Error: "phone number too short"

// Invalid format
phone, err := NewPhone("123456789")
// Error: "unable to parse country code"

// Invalid characters
phone, err := NewPhone("+123456789a")
// Error: "national number must contain only digits"
```

## Integration with Auth System

The enhanced phone parsing is fully integrated with the authentication system:

1. **Registration**: Users can register with international phone numbers
2. **Login**: Users can login using phone number as identity
3. **Validation**: Phone numbers are validated during registration
4. **Storage**: Phone numbers are stored in normalized format

## Best Practices

1. **Always use international format** with `+` prefix
2. **Include country code** for all phone numbers
3. **Validate on client side** before sending to API
4. **Handle errors gracefully** in user interface
5. **Display normalized format** to users after validation
