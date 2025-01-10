package lavatopvalueobjects

import (
	"testing"
)

func TestPaymentMethod_String(t *testing.T) {
	tests := []struct {
		input    PaymentMethod
		expected string
	}{
		{BANK131, "BANK131"},
		{UNLIMINT, "UNLIMINT"},
		{PAYPAL, "PAYPAL"},
		{STRIPE, "STRIPE"},
		{PaymentMethod(999), ""},
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

func TestParsePaymentMethod(t *testing.T) {
	tests := []struct {
		input    string
		expected PaymentMethod
		err      bool
	}{
		{"BANK131", BANK131, false},
		{"UNLIMINT", UNLIMINT, false},
		{"PAYPAL", PAYPAL, false},
		{"STRIPE", STRIPE, false},
		{"bank131", BANK131, false},
		{"unlimint", UNLIMINT, false},
		{"stripe", STRIPE, false},
		{"invalid_method", 0, true},
		{"", 0, true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := ParsePaymentMethod(test.input)
			if (err != nil) != test.err {
				t.Errorf("Unexpected error state: got %v, expected error: %v", err, test.err)
			}
			if result != test.expected {
				t.Errorf("Unexpected result: got %v, expected %v", result, test.expected)
			}
		})
	}
}
