package valueobjects

import (
	"testing"
)

func TestNewHash(t *testing.T) {
	tests := []struct {
		name      string
		hash      string
		expectErr bool
	}{
		{
			name:      "Valid Argon2id Hash",
			hash:      "$argon2id$v=19$m=65536,t=3,p=2$VGVzdFNhbHQ$VGVzdEhhc2g",
			expectErr: false,
		},
		{
			name:      "Empty Hash",
			hash:      "",
			expectErr: true,
		},
		{
			name:      "Invalid Prefix",
			hash:      "$argon2i$v=19$m=65536,t=3,p=2$VGVzdFNhbHQ$VGVzdEhhc2g",
			expectErr: true,
		},
		{
			name:      "Incorrect Version",
			hash:      "$argon2id$v=18$m=65536,t=3,p=2$VGVzdFNhbHQ$VGVzdEhhc2g",
			expectErr: true,
		},
		{
			name:      "Malformed Parameters",
			hash:      "$argon2id$v=19$m=65536,t=3$VGVzdFNhbHQ$VGVzdEhhc2g",
			expectErr: true,
		},
		{
			name:      "Non-numeric Parameters",
			hash:      "$argon2id$v=19$m=sixtyFive,t=three,p=two$VGVzdFNhbHQ$VGVzdEhhc2g",
			expectErr: true,
		},
		{
			name:      "Empty Salt",
			hash:      "$argon2id$v=19$m=65536,t=3,p=2$$VGVzdEhhc2g",
			expectErr: true,
		},
		{
			name:      "Empty Hash Part",
			hash:      "$argon2id$v=19$m=65536,t=3,p=2$VGVzdFNhbHQ$",
			expectErr: true,
		},
		{
			name:      "Invalid Base64 Salt",
			hash:      "$argon2id$v=19$m=65536,t=3,p=2$InvalidBase64$VGVzdEhhc2g",
			expectErr: true,
		},
		{
			name:      "Invalid Base64 Hash",
			hash:      "$argon2id$v=19$m=65536,t=3,p=2$VGVzdFNhbHQ$InvalidBase64",
			expectErr: true,
		},
		{
			name:      "Extra Parts in Hash",
			hash:      "$argon2id$v=19$m=65536,t=3,p=2$VGVzdFNhbHQ$VGVzdEhhc2g$Extra",
			expectErr: true,
		},
		{
			name:      "Missing Parts in Hash",
			hash:      "$argon2id$v=19$m=65536,t=3,p=2$VGVzdFNhbHQ",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewHash(tt.hash)
			if (err != nil) != tt.expectErr {
				t.Errorf("NewHash() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}
