package utils

import (
	"fmt"
	"strings"

	"github.com/nyaruka/phonenumbers"
)

// FormatPhoneToInternational formats a phone number to international format
// This function parses a phone number using the Google Phone Numbers library and
// converts it to a standardized international format without spaces or special characters
//
// Parameters:
//   - phone: the phone number to format (can be in various formats)
//   - countryCode: optional country code for parsing (defaults to "ID" for Indonesia if empty)
//
// Returns:
//   - string: the formatted international phone number without spaces or special characters
//   - error: parsing or validation error if the phone number is invalid
//
// Usage example:
//
//	formatted, err := FormatPhoneToInternational("08123456789", "ID")
//	if err != nil {
//	    // Handle error
//	}
//	// formatted = "+628123456789"
func FormatPhoneToInternational(phone, countryCode string) (string, error) {
	// Use Indonesia as default country code if none specified
	if countryCode == "" {
		countryCode = "ID" // Default to Indonesia
	}

	// Parse the phone number using the Google Phone Numbers library
	// This handles various input formats and country-specific parsing rules
	num, err := phonenumbers.Parse(phone, countryCode)
	if err != nil {
		return "", fmt.Errorf("failed to parse phone number: %w", err)
	}

	// Validate that the parsed number is actually a valid phone number
	// This ensures the number follows proper formatting rules for the country
	if !phonenumbers.IsValidNumber(num) {
		return "", fmt.Errorf("invalid phone number: %s", phone)
	}

	// Format to international format (e.g., "+1 555 123 4567")
	formatted := phonenumbers.Format(num, phonenumbers.INTERNATIONAL)

	// Remove all spaces and special characters for consistent, clean format
	// This makes the output suitable for storage and API calls
	formatted = strings.ReplaceAll(formatted, " ", "") // Remove spaces
	formatted = strings.ReplaceAll(formatted, "-", "") // Remove hyphens
	formatted = strings.ReplaceAll(formatted, ".", "") // Remove dots
	formatted = strings.ReplaceAll(formatted, "(", "") // Remove opening parentheses
	formatted = strings.ReplaceAll(formatted, ")", "") // Remove closing parentheses

	return formatted, nil
}

// ValidatePhone validates a phone number using the Google Phone Numbers library
// This function checks if the provided phone number is valid according to
// international phone number standards and country-specific rules
//
// Parameters:
//   - phone: the phone number to validate (can be in various formats)
//   - countryCode: optional country code for validation (defaults to "ID" for Indonesia if empty)
//
// Returns:
//   - bool: true if the phone number is valid, false otherwise
//   - error: parsing error if the phone number cannot be parsed
//
// Usage example:
//
//	isValid, err := ValidatePhone("08123456789", "ID")
//	if err != nil {
//	    // Handle parsing error
//	}
//	if isValid {
//	    // Phone number is valid
//	}
func ValidatePhone(phone, countryCode string) (bool, error) {
	// Use Indonesia as default country code if none specified
	if countryCode == "" {
		countryCode = "ID" // Default to Indonesia
	}

	// Parse the phone number to check its format and structure
	num, err := phonenumbers.Parse(phone, countryCode)
	if err != nil {
		return false, fmt.Errorf("failed to parse phone number: %w", err)
	}

	// Check if the parsed number is valid according to international standards
	// This includes checks for proper length, format, and country-specific rules
	return phonenumbers.IsValidNumber(num), nil
}

// GetPhoneNumberInfo returns detailed information about a phone number
// This function provides comprehensive metadata about a phone number including
// country code, region, and various formatting options
//
// Parameters:
//   - phone: the phone number to analyze (can be in various formats)
//   - countryCode: optional country code for parsing (defaults to "ID" for Indonesia if empty)
//
// Returns:
//   - *PhoneNumberInfo: detailed phone number information, or nil if invalid
//   - error: parsing or validation error if the phone number is invalid
//
// Usage example:
//
//	info, err := GetPhoneNumberInfo("08123456789", "ID")
//	if err != nil {
//	    // Handle error
//	}
//	fmt.Printf("Country: %s, Code: +%d\n", info.Region, info.CountryCode)
func GetPhoneNumberInfo(phone, countryCode string) (*PhoneNumberInfo, error) {
	// Use Indonesia as default country code if none specified
	if countryCode == "" {
		countryCode = "ID" // Default to Indonesia
	}

	// Parse the phone number to extract detailed information
	num, err := phonenumbers.Parse(phone, countryCode)
	if err != nil {
		return nil, fmt.Errorf("failed to parse phone number: %w", err)
	}

	// Validate that the parsed number is actually valid
	if !phonenumbers.IsValidNumber(num) {
		return nil, fmt.Errorf("invalid phone number: %s", phone)
	}

	// Extract country code (e.g., 62 for Indonesia, 1 for US/Canada)
	countryCodeNum := num.GetCountryCode()

	// Extract national number (the part after country code)
	nationalNumber := num.GetNationalNumber()

	// Get the region/country code (e.g., "ID" for Indonesia, "US" for United States)
	region := phonenumbers.GetRegionCodeForNumber(num)

	// Format the number in different standard formats for flexibility
	international := phonenumbers.Format(num, phonenumbers.INTERNATIONAL) // +1 555 123 4567
	national := phonenumbers.Format(num, phonenumbers.NATIONAL)           // (555) 123-4567
	e164 := phonenumbers.Format(num, phonenumbers.E164)                   // +15551234567

	// Create and return comprehensive phone number information
	return &PhoneNumberInfo{
		CountryCode:    int(countryCodeNum), // Convert to int for JSON compatibility
		NationalNumber: nationalNumber,      // National number without country code
		Region:         region,              // ISO country code
		NumberType:     "UNKNOWN",           // Simplified for now (could be enhanced)
		International:  international,       // Human-readable international format
		National:       national,            // Human-readable national format
		E164:           e164,                // E.164 format (standard for APIs)
		IsValid:        true,                // Mark as valid since we passed validation
	}, nil
}

// PhoneNumberInfo contains comprehensive information about a parsed phone number
// This struct provides all the metadata extracted from a phone number for
// display, validation, and API integration purposes
type PhoneNumberInfo struct {
	CountryCode    int    `json:"countryCode"`    // Country calling code (e.g., 62 for Indonesia)
	NationalNumber uint64 `json:"nationalNumber"` // National number without country code
	Region         string `json:"region"`         // ISO country code (e.g., "ID", "US")
	NumberType     string `json:"numberType"`     // Type of number (e.g., "MOBILE", "FIXED_LINE")
	International  string `json:"international"`  // Human-readable international format
	National       string `json:"national"`       // Human-readable national format
	E164           string `json:"e164"`           // E.164 format (standard for APIs)
	IsValid        bool   `json:"isValid"`        // Whether the number is valid
}

// FormatPhoneToE164 formats a phone number to E.164 format
// E.164 is the international standard format for phone numbers, commonly used
// in APIs and telecommunications systems (e.g., "+628123456789")
//
// Parameters:
//   - phone: the phone number to format (can be in various formats)
//   - countryCode: optional country code for parsing (defaults to "ID" for Indonesia if empty)
//
// Returns:
//   - string: the E.164 formatted phone number
//   - error: parsing or validation error if the phone number is invalid
//
// Usage example:
//
//	e164, err := FormatPhoneToE164("08123456789", "ID")
//	if err != nil {
//	    // Handle error
//	}
//	// e164 = "+628123456789"
//
// Note: E.164 format is the standard for:
//   - SMS APIs
//   - Voice APIs
//   - Database storage
//   - International communications
func FormatPhoneToE164(phone, countryCode string) (string, error) {
	// Use Indonesia as default country code if none specified
	if countryCode == "" {
		countryCode = "ID" // Default to Indonesia
	}

	// Parse the phone number to ensure it's valid
	num, err := phonenumbers.Parse(phone, countryCode)
	if err != nil {
		return "", fmt.Errorf("failed to parse phone number: %w", err)
	}

	// Validate that the parsed number is actually valid
	if !phonenumbers.IsValidNumber(num) {
		return "", fmt.Errorf("invalid phone number: %s", phone)
	}

	// Return the E.164 format (e.g., "+628123456789")
	// This format is ideal for APIs and telecommunications systems
	return phonenumbers.Format(num, phonenumbers.E164), nil
}
