package valueobjects

import (
	"testing"
)

func TestParseGuidFromString(t *testing.T) {
	tests := []struct {
		input       string
		expectError bool
	}{
		{"6c0cf730-3432-4755-941b-ca23b419d6df", false},
		{"6C0CF730-3432-4755-941B-CA23B419D6DF", true},
		{"12345678-1234-1234-1234-123456789abc", false},
		{"invalid-uuid", true},
		{"6c0cf730-3432-4755-941b", true},
		{"", true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			_, err := ParseGuidFromString(test.input)
			if (err != nil) != test.expectError {
				t.Errorf("Unexpected error: got %v, expectError: %v", err, test.expectError)
			}
		})
	}
}

func TestGuid_String(t *testing.T) {
	tests := []struct {
		offerId     Guid
		expectValue string
	}{
		{Guid{"f6860bad-4fdc-4984-8579-a21418760194"}, "f6860bad-4fdc-4984-8579-a21418760194"},
		{Guid{"d71c8e31-c758-42b8-b43e-69d2a670c0c2"}, "d71c8e31-c758-42b8-b43e-69d2a670c0c2"},
		{Guid{"6c0cf730-3432-4755-941b-ca23b419d6df"}, "6c0cf730-3432-4755-941b-ca23b419d6df"},
		{Guid{"12345678-1234-1234-1234-123456789abc"}, "12345678-1234-1234-1234-123456789abc"},
	}

	for _, test := range tests {
		t.Run(test.offerId.value, func(t *testing.T) {
			id := test.offerId.String()
			if id != test.expectValue {
				t.Errorf("Unexpected id: got %q, expected %q", id, test.expectValue)
			}
		})
	}
}

func TestGuid_InvalidCases(t *testing.T) {
	tests := []struct {
		offerId Guid
	}{
		{Guid{"6C0CF730-3432-4755-941B-CA23B419D6DF"}},
		{Guid{"invalid-uuid"}},
		{Guid{"6c0cf730-3432-4755-941b"}},
		{Guid{""}},
	}

	for _, test := range tests {
		t.Run(test.offerId.value, func(t *testing.T) {
			id := test.offerId.String()
			if id != "" {
				t.Errorf("Expected empty string for invalid GUID, got %q", id)
			}
		})
	}
}
