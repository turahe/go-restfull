package utils

import (
	"fmt"
	"strings"

	"github.com/nyaruka/phonenumbers"
)

// FormatPhoneToInternational formats a phone number to international format
// phone: the phone number to format
// countryCode: optional country code (defaults to "ID" if empty)
func FormatPhoneToInternational(phone, countryCode string) (string, error) {
	if countryCode == "" {
		countryCode = "ID" // Default to Indonesia
	}

	// Parse the phone number
	num, err := phonenumbers.Parse(phone, countryCode)
	if err != nil {
		return "", fmt.Errorf("failed to parse phone number: %w", err)
	}

	// Check if the number is valid
	if !phonenumbers.IsValidNumber(num) {
		return "", fmt.Errorf("invalid phone number: %s", phone)
	}

	// Format to international format
	formatted := phonenumbers.Format(num, phonenumbers.INTERNATIONAL)

	// Remove spaces and special characters for consistent format
	formatted = strings.ReplaceAll(formatted, " ", "")
	formatted = strings.ReplaceAll(formatted, "-", "")
	formatted = strings.ReplaceAll(formatted, ".", "")
	formatted = strings.ReplaceAll(formatted, "(", "")
	formatted = strings.ReplaceAll(formatted, ")", "")

	return formatted, nil
}

// ValidatePhone validates a phone number
// phone: the phone number to validate
// countryCode: optional country code (defaults to "ID" if empty)
func ValidatePhone(phone, countryCode string) (bool, error) {
	if countryCode == "" {
		countryCode = "ID" // Default to Indonesia
	}

	// Parse the phone number
	num, err := phonenumbers.Parse(phone, countryCode)
	if err != nil {
		return false, fmt.Errorf("failed to parse phone number: %w", err)
	}

	// Check if the number is valid
	return phonenumbers.IsValidNumber(num), nil
}

// GetPhoneNumberInfo returns detailed information about a phone number
// phone: the phone number to analyze
// countryCode: optional country code (defaults to "ID" if empty)
func GetPhoneNumberInfo(phone, countryCode string) (*PhoneNumberInfo, error) {
	if countryCode == "" {
		countryCode = "ID" // Default to Indonesia
	}

	// Parse the phone number
	num, err := phonenumbers.Parse(phone, countryCode)
	if err != nil {
		return nil, fmt.Errorf("failed to parse phone number: %w", err)
	}

	// Check if the number is valid
	if !phonenumbers.IsValidNumber(num) {
		return nil, fmt.Errorf("invalid phone number: %s", phone)
	}

	// Get country code
	countryCodeNum := num.GetCountryCode()

	// Get national number
	nationalNumber := num.GetNationalNumber()

	// Get region
	region := phonenumbers.GetRegionCodeForNumber(num)

	// Format in different formats
	international := phonenumbers.Format(num, phonenumbers.INTERNATIONAL)
	national := phonenumbers.Format(num, phonenumbers.NATIONAL)
	e164 := phonenumbers.Format(num, phonenumbers.E164)

	return &PhoneNumberInfo{
		CountryCode:    int(countryCodeNum),
		NationalNumber: nationalNumber,
		Region:         region,
		NumberType:     "UNKNOWN", // Simplified for now
		International:  international,
		National:       national,
		E164:           e164,
		IsValid:        true,
	}, nil
}

// PhoneNumberInfo contains detailed information about a phone number
type PhoneNumberInfo struct {
	CountryCode    int    `json:"countryCode"`
	NationalNumber uint64 `json:"nationalNumber"`
	Region         string `json:"region"`
	NumberType     string `json:"numberType"`
	International  string `json:"international"`
	National       string `json:"national"`
	E164           string `json:"e164"`
	IsValid        bool   `json:"isValid"`
}

// FormatPhoneToE164 formats a phone number to E.164 format (international format without spaces)
// phone: the phone number to format
// countryCode: optional country code (defaults to "ID" if empty)
func FormatPhoneToE164(phone, countryCode string) (string, error) {
	if countryCode == "" {
		countryCode = "ID" // Default to Indonesia
	}

	// Parse the phone number
	num, err := phonenumbers.Parse(phone, countryCode)
	if err != nil {
		return "", fmt.Errorf("failed to parse phone number: %w", err)
	}

	// Check if the number is valid
	if !phonenumbers.IsValidNumber(num) {
		return "", fmt.Errorf("invalid phone number: %s", phone)
	}

	// Format to E.164 format
	return phonenumbers.Format(num, phonenumbers.E164), nil
}
