package services

// PasswordService defines the interface for password-related operations
type PasswordService interface {
	// HashPassword hashes a plain text password
	HashPassword(password string) (string, error)
	
	// ComparePassword compares a plain text password with a hashed password
	ComparePassword(hashedPassword, plainPassword string) bool
	
	// ValidatePassword validates password strength
	ValidatePassword(password string) error
} 
