package utils

import (
	"regexp"
	"strings"
)

// Slugify converts a string to a URL-friendly slug
// This function transforms any string into a format suitable for use in URLs,
// filenames, or other contexts where special characters and spaces are problematic
//
// The transformation process:
//   1. Converts the string to lowercase
//   2. Replaces all non-alphanumeric characters with hyphens
//   3. Removes leading and trailing hyphens
//
// Parameters:
//   - s: the input string to convert to a slug
//
// Returns:
//   - string: the URL-friendly slug version of the input string
//
// Usage examples:
//   slug := Slugify("Hello World!")           // Returns "hello-world"
//   slug := Slugify("Product Name (2024)")    // Returns "product-name-2024"
//   slug := Slugify("Special@Characters#")    // Returns "special-characters"
//   slug := Slugify("  Multiple   Spaces  ")  // Returns "multiple-spaces"
//
// Common use cases:
//   - URL slugs for blog posts or articles
//   - Filename generation
//   - Database search optimization
//   - SEO-friendly URLs
func Slugify(s string) string {
	// Convert the entire string to lowercase for consistency
	s = strings.ToLower(s)
	
	// Replace all non-alphanumeric characters with hyphens
	// This regex pattern matches any character that is NOT a-z or 0-9
	s = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(s, "-")
	
	// Remove leading and trailing hyphens that may have been created
	// This ensures the slug doesn't start or end with hyphens
	s = strings.Trim(s, "-")
	
	return s
}
