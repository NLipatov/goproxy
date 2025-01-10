package lavatopvalueobjects

import "testing"

func TestLavaTopInvoiceStatus_String(t *testing.T) {
	tests := []struct {
		input    Status
		expected string
	}{
		{NEW, "new"},
		{INPROGRESS, "in-progress"},
		{COMPLETED, "completed"},
		{FAILED, "failed"},
		{CANCELLED, "cancelled"},
		{SUBSCRIPTIONACTIVE, "subscription-active"},
		{SUBSCRIPTIONEXPIRED, "subscription-expired"},
		{SUBSCRIPTIONCANCELLED, "subscription-cancelled"},
		{SUBSCRIPTIONFAILED, "subscription-failed"},
		{Status(999), "invalid-status"},
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

func TestParseInvoiceStatus(t *testing.T) {
	tests := []struct {
		input       string
		expected    Status
		expectError bool
	}{
		{"new", NEW, false},
		{"in-progress", INPROGRESS, false},
		{"completed", COMPLETED, false},
		{"failed", FAILED, false},
		{"cancelled", CANCELLED, false},
		{"subscription-active", SUBSCRIPTIONACTIVE, false},
		{"subscription-expired", SUBSCRIPTIONEXPIRED, false},
		{"subscription-cancelled", SUBSCRIPTIONCANCELLED, false},
		{"subscription-failed", SUBSCRIPTIONFAILED, false},
		{"NEW", NEW, false},
		{"In-Progress", INPROGRESS, false},
		{"invalid-status", 0, true},
		{"", 0, true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := ParseInvoiceStatus(test.input)
			if (err != nil) != test.expectError {
				t.Errorf("Unexpected error state: got %v, expectError: %v", err, test.expectError)
			}
			if result != test.expected {
				t.Errorf("Unexpected result: got %v, expected %v", result, test.expected)
			}
		})
	}
}
