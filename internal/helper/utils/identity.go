package utils

import "strings"

// ParseIdentity detects if identity is email or phone
func ParseIdentity(identity string) (email, phone string) {
	if strings.Contains(identity, "@") {
		return identity, ""
	}
	return "", identity
}

// IsEmail checks if the given string is an email
func IsEmail(identity string) bool {
	return strings.Contains(identity, "@")
}

// IsPhone checks if the given string is a phone number
func IsPhone(identity string) bool {
	return !strings.Contains(identity, "@")
}
