package valueobjects

import (
	"errors"
	"regexp"
	"strings"
)

// Phone represents a phone number value object
type Phone struct {
	value          string
	countryCode    string
	nationalNumber string
}

// NewPhone creates a new phone value object with validation
func NewPhone(phone string) (Phone, error) {
	phone = strings.TrimSpace(phone)

	if phone == "" {
		return Phone{}, errors.New("phone cannot be empty")
	}

	// Parse the phone number to extract country code and national number
	countryCode, nationalNumber, err := parsePhoneNumber(phone)
	if err != nil {
		return Phone{}, err
	}

	// Validate the parsed components
	if err := validatePhoneComponents(countryCode, nationalNumber); err != nil {
		return Phone{}, err
	}

	// Normalize the full phone number
	normalized := normalizePhone(phone)

	return Phone{
		value:          normalized,
		countryCode:    countryCode,
		nationalNumber: nationalNumber,
	}, nil
}

// String returns the string representation of the phone
func (p Phone) String() string {
	return p.value
}

// Value returns the phone value
func (p Phone) Value() string {
	return p.value
}

// CountryCode returns the country code
func (p Phone) CountryCode() string {
	return p.countryCode
}

// NationalNumber returns the national number (without country code)
func (p Phone) NationalNumber() string {
	return p.nationalNumber
}

// Equals checks if two phones are equal
func (p Phone) Equals(other Phone) bool {
	return p.value == other.value
}

// parsePhoneNumber parses a phone number to extract country code and national number
func parsePhoneNumber(phone string) (string, string, error) {
	// Remove all non-digit characters except +
	cleaned := regexp.MustCompile(`[^\d+]`).ReplaceAllString(phone, "")

	// Handle different phone number formats
	if strings.HasPrefix(cleaned, "+") {
		// International format with + (e.g., +1234567890)
		return parseInternationalFormat(cleaned)
	} else if len(cleaned) >= 10 {
		// Assume it's a national number, try to detect country code
		return parseNationalFormat(cleaned)
	} else {
		return "", "", errors.New("invalid phone number format")
	}
}

// parseInternationalFormat parses international format phone numbers
func parseInternationalFormat(phone string) (string, string, error) {
	// Remove the + sign
	phone = strings.TrimPrefix(phone, "+")

	if len(phone) < 10 {
		return "", "", errors.New("phone number too short")
	}

	// Common country code patterns
	countryCodes := []string{
		"1",   // US/Canada
		"44",  // UK
		"33",  // France
		"49",  // Germany
		"39",  // Italy
		"34",  // Spain
		"31",  // Netherlands
		"32",  // Belgium
		"41",  // Switzerland
		"43",  // Austria
		"46",  // Sweden
		"47",  // Norway
		"45",  // Denmark
		"358", // Finland
		"48",  // Poland
		"420", // Czech Republic
		"36",  // Hungary
		"40",  // Romania
		"421", // Slovakia
		"386", // Slovenia
		"385", // Croatia
		"387", // Bosnia
		"389", // Macedonia
		"381", // Serbia
		"382", // Montenegro
		"355", // Albania
		"371", // Latvia
		"372", // Estonia
		"370", // Lithuania
		"375", // Belarus
		"380", // Ukraine
		"373", // Moldova
		"374", // Armenia
		"995", // Georgia
		"994", // Azerbaijan
		"7",   // Russia/Kazakhstan
		"86",  // China
		"81",  // Japan
		"82",  // South Korea
		"84",  // Vietnam
		"66",  // Thailand
		"65",  // Singapore
		"60",  // Malaysia
		"63",  // Philippines
		"62",  // Indonesia
		"91",  // India
		"880", // Bangladesh
		"94",  // Sri Lanka
		"95",  // Myanmar
		"856", // Laos
		"855", // Cambodia
		"977", // Nepal
		"975", // Bhutan
		"880", // Bangladesh
		"960", // Maldives
		"93",  // Afghanistan
		"98",  // Iran
		"964", // Iraq
		"962", // Jordan
		"961", // Lebanon
		"963", // Syria
		"966", // Saudi Arabia
		"968", // Oman
		"971", // UAE
		"973", // Bahrain
		"974", // Qatar
		"965", // Kuwait
		"967", // Yemen
		"972", // Israel
		"970", // Palestine
		"20",  // Egypt
		"212", // Morocco
		"213", // Algeria
		"216", // Tunisia
		"218", // Libya
		"220", // Gambia
		"221", // Senegal
		"222", // Mauritania
		"223", // Mali
		"224", // Guinea
		"225", // Ivory Coast
		"226", // Burkina Faso
		"227", // Niger
		"228", // Togo
		"229", // Benin
		"230", // Mauritius
		"231", // Liberia
		"232", // Sierra Leone
		"233", // Ghana
		"234", // Nigeria
		"235", // Chad
		"236", // Central African Republic
		"237", // Cameroon
		"238", // Cape Verde
		"239", // Sao Tome and Principe
		"240", // Equatorial Guinea
		"241", // Gabon
		"242", // Republic of the Congo
		"243", // Democratic Republic of the Congo
		"244", // Angola
		"245", // Guinea-Bissau
		"246", // British Indian Ocean Territory
		"247", // Ascension Island
		"248", // Seychelles
		"249", // Sudan
		"250", // Rwanda
		"251", // Ethiopia
		"252", // Somalia
		"253", // Djibouti
		"254", // Kenya
		"255", // Tanzania
		"256", // Uganda
		"257", // Burundi
		"258", // Mozambique
		"260", // Zambia
		"261", // Madagascar
		"262", // Reunion
		"263", // Zimbabwe
		"264", // Namibia
		"265", // Malawi
		"266", // Lesotho
		"267", // Botswana
		"268", // Swaziland
		"269", // Comoros
		"27",  // South Africa
		"290", // Saint Helena
		"291", // Eritrea
		"297", // Aruba
		"298", // Faroe Islands
		"299", // Greenland
		"30",  // Greece
		"31",  // Netherlands
		"32",  // Belgium
		"33",  // France
		"34",  // Spain
		"350", // Gibraltar
		"351", // Portugal
		"352", // Luxembourg
		"353", // Ireland
		"354", // Iceland
		"355", // Albania
		"356", // Malta
		"357", // Cyprus
		"358", // Finland
		"359", // Bulgaria
		"36",  // Hungary
		"37",  // Romania
		"380", // Ukraine
		"381", // Serbia
		"382", // Montenegro
		"383", // Kosovo
		"385", // Croatia
		"386", // Slovenia
		"387", // Bosnia and Herzegovina
		"389", // Macedonia
		"39",  // Italy
		"40",  // Romania
		"41",  // Switzerland
		"420", // Czech Republic
		"421", // Slovakia
		"423", // Liechtenstein
		"43",  // Austria
		"44",  // United Kingdom
		"45",  // Denmark
		"46",  // Sweden
		"47",  // Norway
		"48",  // Poland
		"49",  // Germany
		"500", // Falkland Islands
		"501", // Belize
		"502", // Guatemala
		"503", // El Salvador
		"504", // Honduras
		"505", // Nicaragua
		"506", // Costa Rica
		"507", // Panama
		"508", // Saint Pierre and Miquelon
		"509", // Haiti
		"51",  // Peru
		"52",  // Mexico
		"53",  // Cuba
		"54",  // Argentina
		"55",  // Brazil
		"56",  // Chile
		"57",  // Colombia
		"58",  // Venezuela
		"590", // Guadeloupe
		"591", // Bolivia
		"592", // Guyana
		"593", // Ecuador
		"594", // French Guiana
		"595", // Paraguay
		"596", // Martinique
		"597", // Suriname
		"598", // Uruguay
		"599", // Netherlands Antilles
		"60",  // Malaysia
		"61",  // Australia
		"62",  // Indonesia
		"63",  // Philippines
		"64",  // New Zealand
		"65",  // Singapore
		"66",  // Thailand
		"670", // East Timor
		"672", // Australian External Territories
		"673", // Brunei
		"674", // Nauru
		"675", // Papua New Guinea
		"676", // Tonga
		"677", // Solomon Islands
		"678", // Vanuatu
		"679", // Fiji
		"680", // Palau
		"681", // Wallis and Futuna
		"682", // Cook Islands
		"683", // Niue
		"685", // Samoa
		"686", // Kiribati
		"687", // New Caledonia
		"688", // Tuvalu
		"689", // French Polynesia
		"690", // Tokelau
		"691", // Micronesia
		"692", // Marshall Islands
		"7",   // Russia
		"81",  // Japan
		"82",  // South Korea
		"84",  // Vietnam
		"850", // North Korea
		"852", // Hong Kong
		"853", // Macau
		"855", // Cambodia
		"856", // Laos
		"86",  // China
		"880", // Bangladesh
		"886", // Taiwan
		"90",  // Turkey
		"91",  // India
		"92",  // Pakistan
		"93",  // Afghanistan
		"94",  // Sri Lanka
		"95",  // Myanmar
		"960", // Maldives
		"961", // Lebanon
		"962", // Jordan
		"963", // Syria
		"964", // Iraq
		"965", // Kuwait
		"966", // Saudi Arabia
		"967", // Yemen
		"968", // Oman
		"970", // Palestine
		"971", // UAE
		"972", // Israel
		"973", // Bahrain
		"974", // Qatar
		"975", // Bhutan
		"976", // Mongolia
		"977", // Nepal
		"98",  // Iran
		"992", // Tajikistan
		"993", // Turkmenistan
		"994", // Azerbaijan
		"995", // Georgia
		"996", // Kyrgyzstan
		"998", // Uzbekistan
	}

	// Try to match country codes from longest to shortest
	for _, code := range countryCodes {
		if strings.HasPrefix(phone, code) {
			nationalNum := strings.TrimPrefix(phone, code)
			if len(nationalNum) >= 7 && len(nationalNum) <= 15 {
				return code, nationalNum, nil
			}
		}
	}

	return "", "", errors.New("unable to parse country code")
}

// parseNationalFormat parses national format phone numbers
func parseNationalFormat(phone string) (string, string, error) {
	// For national numbers, we'll assume common country codes
	// This is a simplified approach - in production, you might want to use a library like libphonenumber

	// Try common country codes
	countryCodes := []string{"1", "44", "33", "49", "39", "34", "31", "32", "41", "43", "46", "47", "45", "48", "86", "81", "82", "84", "91", "62", "63", "65", "66", "60"}

	for _, code := range countryCodes {
		if strings.HasPrefix(phone, code) {
			nationalNum := strings.TrimPrefix(phone, code)
			if len(nationalNum) >= 7 && len(nationalNum) <= 15 {
				return code, nationalNum, nil
			}
		}
	}

	// If no country code found, assume it's a local number
	// This is a fallback - in production, you should require explicit country codes
	if len(phone) >= 10 && len(phone) <= 15 {
		return "1", phone, nil // Default to US/Canada
	}

	return "", "", errors.New("unable to parse national phone number")
}

// validatePhoneComponents validates the country code and national number
func validatePhoneComponents(countryCode, nationalNumber string) error {
	if countryCode == "" {
		return errors.New("country code is required")
	}

	if nationalNumber == "" {
		return errors.New("national number is required")
	}

	// Validate national number length (typically 7-15 digits)
	if len(nationalNumber) < 7 || len(nationalNumber) > 15 {
		return errors.New("national number must be between 7 and 15 digits")
	}

	// Validate that national number contains only digits
	if !regexp.MustCompile(`^\d+$`).MatchString(nationalNumber) {
		return errors.New("national number must contain only digits")
	}

	return nil
}

// isValidPhone validates phone format (legacy method for backward compatibility)
func isValidPhone(phone string) bool {
	// Accept various international phone formats
	phoneRegex := regexp.MustCompile(`^\+?[\d\s\-\(\)]{10,15}$`)
	return phoneRegex.MatchString(phone)
}

// normalizePhone normalizes the phone number
func normalizePhone(phone string) string {
	// Remove all non-digit characters except +
	reg := regexp.MustCompile(`[^\d+]`)
	return reg.ReplaceAllString(phone, "")
}
