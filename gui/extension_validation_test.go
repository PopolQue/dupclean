package gui

import (
	"testing"
)

// TestIsValidExtension tests the extension validation function
func TestIsValidExtension(t *testing.T) {
	tests := []struct {
		ext      string
		expected bool
		desc     string
	}{
		// Valid extensions
		{".txt", true, "simple extension"},
		{".TXT", true, "uppercase extension"},
		{".pdf", true, "common extension"},
		{".jpg", true, "image extension"},
		{".tar.gz", true, "double extension"},
		{".config.json", true, "triple extension"},
		{".a", true, "single char extension"},
		{".123", true, "numeric extension"},
		{".txt123", true, "alphanumeric extension"},

		// Invalid - empty/missing dot
		{"", false, "empty string"},
		{"txt", false, "missing dot"},
		{".", false, "only dot"},

		// Invalid - dangerous characters (regex injection)
		{".*", false, "wildcard pattern"},
		{".+", false, "one-or-more pattern"},
		{".?", false, "optional pattern"},
		{".txt*", false, "suffix wildcard"},
		{".{2,4}", false, "regex quantifier"},
		{".[a-z]", false, "character class"},
		{".(txt|pdf)", false, "alternation"},
		{".^txt", false, "start anchor"},
		{".txt$", false, "end anchor"},
		{".\\w+", false, "escape sequence"},
		{".txt|pdf", false, "pipe character"},

		// Invalid - too long
		{".verylongextensionname", false, "too long (>20 chars)"},

		// Edge cases
		{"..txt", true, "double dot (hidden file pattern)"},
		{".test.txt", true, "dot in middle"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			result := isValidExtension(tt.ext)
			if result != tt.expected {
				t.Errorf("isValidExtension(%q) = %v, want %v (%s)", tt.ext, result, tt.expected, tt.desc)
			}
		})
	}
}

// TestIsValidExtension_LengthLimit tests the length limit
func TestIsValidExtension_LengthLimit(t *testing.T) {
	// Exactly 20 chars should be OK
	ext20 := ".abcdefghij123456789" // 20 chars
	if !isValidExtension(ext20) {
		t.Errorf("isValidExtension(%q) should be valid (exactly 20 chars)", ext20)
	}

	// 21 chars should be invalid
	ext21 := ".abcdefghij1234567890" // 21 chars
	if isValidExtension(ext21) {
		t.Errorf("isValidExtension(%q) should be invalid (>20 chars)", ext21)
	}
}

// TestIsValidExtension_NoDangerousChars tests that all dangerous chars are blocked
func TestIsValidExtension_NoDangerousChars(t *testing.T) {
	dangerousChars := []string{"*", "+", "?", "{", "}", "[", "]", "(", ")", "|", "^", "$", "\\"}

	for _, char := range dangerousChars {
		ext := "." + char + "txt"
		if isValidExtension(ext) {
			t.Errorf("isValidExtension(%q) should block dangerous char %q", ext, char)
		}
	}
}

// TestIsValidExtension_AllowedChars tests that all allowed chars work
func TestIsValidExtension_AllowedChars(t *testing.T) {
	// Test all lowercase letters
	for c := 'a'; c <= 'z'; c++ {
		ext := "." + string(c)
		if !isValidExtension(ext) {
			t.Errorf("isValidExtension(%q) should allow lowercase letter", ext)
		}
	}

	// Test all uppercase letters
	for c := 'A'; c <= 'Z'; c++ {
		ext := "." + string(c)
		if !isValidExtension(ext) {
			t.Errorf("isValidExtension(%q) should allow uppercase letter", ext)
		}
	}

	// Test all digits
	for c := '0'; c <= '9'; c++ {
		ext := "." + string(c)
		if !isValidExtension(ext) {
			t.Errorf("isValidExtension(%q) should allow digit", ext)
		}
	}

	// Test dot in middle
	if !isValidExtension(".tar.gz") {
		t.Error("isValidExtension(.tar.gz) should allow dots in middle")
	}
}
