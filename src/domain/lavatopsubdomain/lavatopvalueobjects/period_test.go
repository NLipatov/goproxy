package lavatopvalueobjects

import (
	"testing"
)

func TestPeriodicity_String(t *testing.T) {
	tests := []struct {
		input    Periodicity
		expected string
	}{
		{ONE_TIME, "ONE_TIME"},
		{MONTHLY, "MONTHLY"},
		{PERIOD_90_DAYS, "PERIOD_90_DAYS"},
		{PERIOD_180_DAYS, "PERIOD_180_DAYS"},
		{PERIOD_YEAR, "PERIOD_YEAR"},
		{Periodicity(999), ""},
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

func TestParsePeriodicity(t *testing.T) {
	tests := []struct {
		input    string
		expected Periodicity
		err      bool
	}{
		{"ONE_TIME", ONE_TIME, false},
		{"MONTHLY", MONTHLY, false},
		{"PERIOD_90_DAYS", PERIOD_90_DAYS, false},
		{"PERIOD_180_DAYS", PERIOD_180_DAYS, false},
		{"PERIOD_YEAR", PERIOD_YEAR, false},
		{"one_time", ONE_TIME, false}, // case-insensitive
		{"monthly", MONTHLY, false},
		{"invalid_period", 0, true},
		{"", 0, true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := ParsePeriodicity(test.input)
			if (err != nil) != test.err {
				t.Errorf("Unexpected error state: got %v, expected error: %v", err, test.err)
			}
			if result != test.expected {
				t.Errorf("Unexpected result: got %v, expected %v", result, test.expected)
			}
		})
	}
}
