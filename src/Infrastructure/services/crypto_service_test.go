package services_test

import (
	"goproxy/Infrastructure/services"
	"testing"
)

func TestCryptoService(t *testing.T) {
	var saltLength = 16
	var service = services.NewCryptoService(saltLength)

	t.Run("GenerateSalt", func(t *testing.T) {
		salt, err := service.GenerateSalt()
		if err != nil {
			t.Fatalf("Failed to generate salt: %v", err)
		}

		if len(salt) != saltLength {
			t.Errorf("Expected salt length %d, got %d", saltLength, len(salt))
		}
	})

	t.Run("HashValue and ValidateHash - Valid Password", func(t *testing.T) {
		password := "secure-password"
		salt, err := service.GenerateSalt()
		if err != nil {
			t.Fatalf("Failed to generate salt: %v", err)
		}

		hash, err := service.HashValue(password, salt)
		if err != nil {
			t.Fatalf("Failed to hash value: %v", err)
		}

		if !service.ValidateHash(hash, salt, password) {
			t.Errorf("Password validation failed for correct password")
		}
	})

	t.Run("ValidateHash - Invalid Password", func(t *testing.T) {
		password := "secure-password"
		invalidPassword := "wrong-password"
		salt, err := service.GenerateSalt()
		if err != nil {
			t.Fatalf("Failed to generate salt: %v", err)
		}

		hash, err := service.HashValue(password, salt)
		if err != nil {
			t.Fatalf("Failed to hash value: %v", err)
		}

		if service.ValidateHash(hash, salt, invalidPassword) {
			t.Errorf("Password validation succeeded for incorrect password")
		}
	})

	t.Run("ValidateHash - Different Salt", func(t *testing.T) {
		password := "secure-password"
		salt1, err := service.GenerateSalt()
		if err != nil {
			t.Fatalf("Failed to generate salt: %v", err)
		}

		salt2, err := service.GenerateSalt()
		if err != nil {
			t.Fatalf("Failed to generate second salt: %v", err)
		}

		hash, err := service.HashValue(password, salt1)
		if err != nil {
			t.Fatalf("Failed to hash value: %v", err)
		}

		if service.ValidateHash(hash, salt2, password) {
			t.Errorf("Password validation succeeded with different salt")
		}
	})
}
