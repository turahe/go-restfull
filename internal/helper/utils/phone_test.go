package utils

import (
	"testing"
)

func TestFormatPhoneToInternational(t *testing.T) {
	tests := []struct {
		name        string
		phone       string
		countryCode string
		expected    string
		expectError bool
	}{
		{
			name:        "Indonesian number with country code",
			phone:       "08123456789",
			countryCode: "ID",
			expected:    "+628123456789",
			expectError: false,
		},
		{
			name:        "Indonesian number without country code (should default to ID)",
			phone:       "08123456789",
			countryCode: "",
			expected:    "+628123456789",
			expectError: false,
		},
		{
			name:        "US number",
			phone:       "555-123-4567",
			countryCode: "US",
			expected:    "",
			expectError: true,
		},
		{
			name:        "Invalid number",
			phone:       "123",
			countryCode: "ID",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatPhoneToInternational(tt.phone, tt.countryCode)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestValidatePhone(t *testing.T) {
	tests := []struct {
		name        string
		phone       string
		countryCode string
		expected    bool
		expectError bool
	}{
		{
			name:        "Valid Indonesian number",
			phone:       "08123456789",
			countryCode: "ID",
			expected:    true,
			expectError: false,
		},
		{
			name:        "Invalid number",
			phone:       "123",
			countryCode: "ID",
			expected:    false,
			expectError: false,
		},
		{
			name:        "Valid US number",
			phone:       "555-123-4567",
			countryCode: "US",
			expected:    false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidatePhone(tt.phone, tt.countryCode)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetPhoneNumberInfo(t *testing.T) {
	info, err := GetPhoneNumberInfo("08123456789", "ID")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if info == nil {
		t.Error("Expected phone info but got nil")
		return
	}

	if info.CountryCode != 62 {
		t.Errorf("Expected country code 62, got %d", info.CountryCode)
	}

	if info.Region != "ID" {
		t.Errorf("Expected region ID, got %s", info.Region)
	}

	if !info.IsValid {
		t.Error("Expected valid phone number")
	}
}

func TestFormatPhoneToE164(t *testing.T) {
	result, err := FormatPhoneToE164("08123456789", "ID")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	expected := "+628123456789"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}
