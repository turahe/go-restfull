package utils

import "golang.org/x/crypto/bcrypt"

// GeneratePassword creates a secure hash from a plain text password using bcrypt
// This function uses bcrypt's default cost factor for optimal security vs performance balance
// The generated hash is safe to store in databases and should never be stored as plain text
//
// Parameters:
//   - p: the plain text password to hash
//
// Returns:
//   - string: the bcrypt hash of the password
//
// Note: This function will panic if bcrypt fails to generate the hash, which should
// only happen in extreme cases (e.g., out of memory). In production, consider
// handling this error more gracefully.
//
// Usage example:
//
//	hashedPassword := GeneratePassword("mySecurePassword123")
//	// Store hashedPassword in database
func GeneratePassword(p string) string {
	// Generate bcrypt hash with default cost (10)
	// Higher cost = more secure but slower to generate and verify
	hash, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	if err != nil {
		// Panic if bcrypt fails - this should rarely happen
		panic(err)
	}
	return string(hash)
}

// ComparePassword verifies if a plain text password matches a stored bcrypt hash
// This function safely compares passwords without timing attacks and is the
// recommended way to verify passwords in Go applications
//
// Parameters:
//   - hashedPassword: the bcrypt hash stored in the database
//   - password: the plain text password to verify
//
// Returns:
//   - bool: true if passwords match, false otherwise
//
// Security notes:
//   - This function is safe against timing attacks
//   - Always use this function instead of direct string comparison
//   - The function will return false for invalid bcrypt hashes
//
// Usage example:
//
//	if ComparePassword(storedHash, userInput) {
//	    // Password is correct, proceed with authentication
//	} else {
//	    // Password is incorrect, deny access
//	}
func ComparePassword(hashedPassword, password string) bool {
	// Compare the plain text password with the stored hash
	// bcrypt.CompareHashAndPassword handles all the security aspects
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
