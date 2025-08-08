package valueobjects

import (
	"testing"
)

func TestNewPhone_InternationalFormat(t *testing.T) {
	tests := []struct {
		name        string
		phone       string
		expectError bool
		countryCode string
		nationalNum string
	}{
		{
			name:        "US phone with +",
			phone:       "+1234567890",
			expectError: false,
			countryCode: "1",
			nationalNum: "234567890",
		},
		{
			name:        "UK phone with +",
			phone:       "+447911123456",
			expectError: false,
			countryCode: "44",
			nationalNum: "7911123456",
		},
		{
			name:        "German phone with +",
			phone:       "+49123456789",
			expectError: false,
			countryCode: "49",
			nationalNum: "123456789",
		},
		{
			name:        "Chinese phone with +",
			phone:       "+8612345678901",
			expectError: false,
			countryCode: "86",
			nationalNum: "12345678901",
		},
		{
			name:        "Indian phone with +",
			phone:       "+919876543210",
			expectError: false,
			countryCode: "91",
			nationalNum: "9876543210",
		},
		{
			name:        "Phone with spaces",
			phone:       "+1 234 567 8900",
			expectError: false,
			countryCode: "1",
			nationalNum: "2345678900",
		},
		{
			name:        "Phone with dashes",
			phone:       "+1-234-567-8900",
			expectError: false,
			countryCode: "1",
			nationalNum: "2345678900",
		},
		{
			name:        "Phone with parentheses",
			phone:       "+1 (234) 567-8900",
			expectError: false,
			countryCode: "1",
			nationalNum: "2345678900",
		},
		{
			name:        "Invalid phone too short",
			phone:       "+123456",
			expectError: true,
		},
		{
			name:        "Invalid phone no country code",
			phone:       "123456789",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			phone, err := NewPhone(tt.phone)

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

			if phone.CountryCode() != tt.countryCode {
				t.Errorf("Country code mismatch: got %s, want %s", phone.CountryCode(), tt.countryCode)
			}

			if phone.NationalNumber() != tt.nationalNum {
				t.Errorf("National number mismatch: got %s, want %s", phone.NationalNumber(), tt.nationalNum)
			}

			// Test String() method returns normalized format
			expectedNormalized := "+" + tt.countryCode + tt.nationalNum
			if phone.String() != expectedNormalized {
				t.Errorf("String() mismatch: got %s, want %s", phone.String(), expectedNormalized)
			}
		})
	}
}

func TestNewPhone_NationalFormat(t *testing.T) {
	tests := []struct {
		name        string
		phone       string
		expectError bool
		countryCode string
		nationalNum string
	}{
		{
			name:        "US phone without +",
			phone:       "1234567890",
			expectError: false,
			countryCode: "1",
			nationalNum: "234567890",
		},
		{
			name:        "UK phone without +",
			phone:       "447911123456",
			expectError: false,
			countryCode: "44",
			nationalNum: "7911123456",
		},
		{
			name:        "German phone without +",
			phone:       "49123456789",
			expectError: false,
			countryCode: "49",
			nationalNum: "123456789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			phone, err := NewPhone(tt.phone)

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

			if phone.CountryCode() != tt.countryCode {
				t.Errorf("Country code mismatch: got %s, want %s", phone.CountryCode(), tt.countryCode)
			}

			if phone.NationalNumber() != tt.nationalNum {
				t.Errorf("National number mismatch: got %s, want %s", phone.NationalNumber(), tt.nationalNum)
			}
		})
	}
}

func TestPhone_Equals(t *testing.T) {
	phone1, _ := NewPhone("+1234567890")
	phone2, _ := NewPhone("+1234567890")
	phone3, _ := NewPhone("+447911123456")

	if !phone1.Equals(phone2) {
		t.Error("Expected phones to be equal")
	}

	if phone1.Equals(phone3) {
		t.Error("Expected phones to be different")
	}
}

func TestPhone_Value(t *testing.T) {
	phone, _ := NewPhone("+1234567890")
	expected := "+1234567890"

	if phone.Value() != expected {
		t.Errorf("Value() mismatch: got %s, want %s", phone.Value(), expected)
	}
}
