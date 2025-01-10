package valueobjects

import (
	"testing"
)

func TestParseEmailFromString_Extended(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
	}{
		// valid
		{"Simple valid email", "user@example.com", false},
		{"Email with subdomain", "user@sub.example.com", false},
		{"Email with numbers", "user123@example.com", false},
		{"Email with allowed special chars", "user.name+tag@example.com", false},
		{"Email with dash in domain", "user-name@example-domain.com", false},
		{"Email with numeric domain", "user@123.123.123.123", false},
		{"Email with IPv6", "user@[2001:db8::1234:5678]", false},
		{"IPv6 loopback", "user@[::1]", false},
		{"Short IPv6", "user@[::]", false},
		{"Email with long TLD", "user@example.technology", false},
		{"Email with uppercase", "USER@EXAMPLE.COM", false},
		{"Email with single-character local part", "u@example.com", false},
		{"Email with valid international domain", "user@xn--d1acufc.xn--p1ai", false}, // Internationalized domain name (IDN)

		// invalid
		{"Missing @ symbol", "userexample.com", true},
		{"Double @ symbols", "user@@example.com", true},
		{"No domain part", "user@", true},
		{"No local part", "@example.com", true},
		{"Space in local part", "us er@example.com", true},
		{"Space in domain", "user@exa mple.com", true},
		{"Special char not allowed in local part", "user!name@example.com", true},
		{"Unescaped quotes in local part", `"user"example.com`, true},
		{"Domain missing TLD", "user@example", true},
		{"TLD too short", "user@example.c", true},
		{"TLD too long", "user@example.abcdefghijklm", true},
		{"IP without brackets", "user@192.168.1.1.1", true}, // Invalid IPv4 address with 5 octets
		{"Email with invalid IDN", "user@xn--example-.com", true},

		// invalid. edge cases
		{"Empty string", "", true},
		{"Only @ symbol", "@", true},
		{"Double dots in domain", "user@example..com", true},
		{"Trailing dot in domain", "user@example.com.", true},
		{"Leading dot in local part", ".user@example.com", true},
		{"Trailing dot in local part", "user.@example.com", true},
		{"Only special chars in local part", "#@example.com", true},
		{"Email with new line", "user@example.com\n", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseEmailFromString(test.input)
			if (err != nil) != test.expectErr {
				t.Errorf("Test failed for input: %q, got error: %v, want error: %v", test.input, err, test.expectErr)
			}
		})
	}
}
