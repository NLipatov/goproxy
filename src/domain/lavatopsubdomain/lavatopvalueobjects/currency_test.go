package lavatopvalueobjects

import (
	"testing"
)

func TestCurrency_String(t *testing.T) {
	tests := []struct {
		input    Currency
		expected string
	}{
		{RUB, "RUB"},
		{USD, "USD"},
		{EUR, "EUR"},
		{Currency(999), ""},
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

func TestParseCurrency(t *testing.T) {
	tests := []struct {
		input    string
		expected Currency
		err      bool
	}{
		{"RUB", RUB, false},
		{"USD", USD, false},
		{"EUR", EUR, false},
		{"rub", RUB, false},
		{"usd", USD, false},
		{"eur", EUR, false},
		{"invalid_currency", 0, true},
		{"", 0, true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := ParseCurrency(test.input)
			if (err != nil) != test.err {
				t.Errorf("Unexpected error: got %v, expected error: %v", err, test.err)
			}
			if result != test.expected {
				t.Errorf("Unexpected result: got %v, expected %v", result, test.expected)
			}
		})
	}
}
