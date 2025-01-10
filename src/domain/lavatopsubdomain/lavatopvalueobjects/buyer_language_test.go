package lavatopvalueobjects

import (
	"testing"
)

func TestBuyerLanguage_String(t *testing.T) {
	tests := []struct {
		input    BuyerLanguage
		expected string
	}{
		{EN, "EN"},
		{RU, "RU"},
		{ES, "ES"},
		{BuyerLanguage(999), ""},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			result := test.input.String()
			if result != test.expected {
				t.Errorf("Unexpected result: got %q, expected %q", result, test.expected)
			}
		})
	}
}

func TestParseBuyerLanguage(t *testing.T) {
	tests := []struct {
		input    string
		expected BuyerLanguage
		err      bool
	}{
		{"EN", EN, false},
		{"RU", RU, false},
		{"ES", ES, false},
		{"en", EN, false},
		{"ru", RU, false},
		{"es", ES, false},
		{"invalid_language", 0, true},
		{"", 0, true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := ParseBuyerLanguage(test.input)
			if (err != nil) != test.err {
				t.Errorf("Unexpected error: got %v, expected error: %v", err, test.err)
			}
			if result != test.expected {
				t.Errorf("Unexpected result: got %v, expected %v", result, test.expected)
			}
		})
	}
}
