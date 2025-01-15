package valueobjects

import (
	"testing"
)

func TestNewUsernameFromString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantValue string
		wantErr   bool
	}{
		{
			name:      "valid username",
			input:     "valid_username",
			wantValue: "valid_username",
			wantErr:   false,
		},
		{
			name:      "empty username",
			input:     "",
			wantValue: "",
			wantErr:   true,
		},
		{
			name:      "username with uppercase letters",
			input:     "InvalidUsername",
			wantValue: "",
			wantErr:   true,
		},
		{
			name:      "username with spaces",
			input:     "user name",
			wantValue: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewUsernameFromString(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("expected error = %v, got error = %v", tt.wantErr, err)
			}

			if got.Value != tt.wantValue {
				t.Errorf("expected value = %q, got value = %q", tt.wantValue, got.Value)
			}
		})
	}
}

func TestNormalizeUsername(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "lowercase username",
			input: "username",
			want:  "username",
		},
		{
			name:  "username with spaces",
			input: "user name",
			want:  "user_name",
		},
		{
			name:  "username with uppercase letters",
			input: "UserName",
			want:  "username",
		},
		{
			name:  "username with mixed cases and spaces",
			input: "User Name",
			want:  "user_name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeUsername(tt.input)

			if got != tt.want {
				t.Errorf("expected = %q, got = %q", tt.want, got)
			}
		})
	}
}
