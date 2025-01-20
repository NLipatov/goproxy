package valueobjects

import (
	"testing"
)

func TestNewHash(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"Valid hash", "valid_hash", false},
		{"Empty hash", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewHash(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Value != tt.input {
				t.Errorf("NewHash() = %v, want %v", got.Value, tt.input)
			}
		})
	}
}
