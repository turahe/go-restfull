package utils

import (
	"testing"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple word",
			input:    "Hello",
			expected: "hello",
		},
		{
			name:     "multiple words",
			input:    "Hello World",
			expected: "hello-world",
		},
		{
			name:     "special characters",
			input:    "Hello@World!",
			expected: "hello-world",
		},
		{
			name:     "numbers and letters",
			input:    "Post 123",
			expected: "post-123",
		},
		{
			name:     "multiple spaces",
			input:    "Hello   World",
			expected: "hello-world",
		},
		{
			name:     "leading and trailing spaces",
			input:    "  Hello World  ",
			expected: "hello-world",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only special characters",
			input:    "@#$%",
			expected: "",
		},
		{
			name:     "mixed case with numbers",
			input:    "My Post 2024",
			expected: "my-post-2024",
		},
		{
			name:     "underscores and hyphens",
			input:    "my_post-title",
			expected: "my-post-title",
		},
		{
			name:     "consecutive special characters",
			input:    "Hello!!!World",
			expected: "hello-world",
		},
		{
			name:     "unicode characters",
			input:    "Café & Résumé",
			expected: "caf-r-sum",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Slugify(tt.input)
			if result != tt.expected {
				t.Errorf("Slugify(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSlugify_EdgeCases(t *testing.T) {
	// Test with very long string
	longInput := "This is a very long string that should be properly slugified with many words and special characters @#$%"
	expected := "this-is-a-very-long-string-that-should-be-properly-slugified-with-many-words-and-special-characters"
	result := Slugify(longInput)
	if result != expected {
		t.Errorf("Slugify(long string) = %q, want %q", result, expected)
	}

	// Test with only numbers
	result = Slugify("123456")
	if result != "123456" {
		t.Errorf("Slugify('123456') = %q, want '123456'", result)
	}

	// Test with single character
	result = Slugify("A")
	if result != "a" {
		t.Errorf("Slugify('A') = %q, want 'a'", result)
	}
}

func BenchmarkSlugify(b *testing.B) {
	input := "This is a benchmark test string with special characters @#$%"
	for i := 0; i < b.N; i++ {
		Slugify(input)
	}
}
