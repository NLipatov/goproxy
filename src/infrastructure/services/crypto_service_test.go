package services_test

import (
	"goproxy/infrastructure/services"
	"math"
	"strings"
	"testing"
)

func TestCryptoService(t *testing.T) {
	service := services.GetCryptoService()

	t.Run("HashValue and ValidateHash - Valid Password", func(t *testing.T) {
		password := "secure-password"

		hash, err := service.HashValue(password)
		if err != nil {
			t.Fatalf("Failed to hash value: %v", err)
		}

		if hash == "" {
			t.Fatal("HashValue returned an empty string")
		}

		if !strings.HasPrefix(hash, "$argon2id$") {
			t.Errorf("Hash does not have expected prefix $argon2id$: %s", hash)
		}

		if !service.ValidateHash(hash, password) {
			t.Errorf("Password validation failed for correct password")
		}
	})

	t.Run("ValidateHash - Invalid Password", func(t *testing.T) {
		password := "secure-password"
		invalidPassword := "wrong-password"

		hash, err := service.HashValue(password)
		if err != nil {
			t.Fatalf("Failed to hash value: %v", err)
		}

		if service.ValidateHash(hash, invalidPassword) {
			t.Errorf("Password validation succeeded for incorrect password")
		}
	})

	t.Run("ValidateHash - Malformed Hash", func(t *testing.T) {
		password := "secure-password"
		malformedHash := "invalid-hash-format"

		if service.ValidateHash(malformedHash, password) {
			t.Errorf("Password validation succeeded for malformed hash")
		}
	})

	t.Run("GenerateRandomString - Valid Length", func(t *testing.T) {
		stringLength := 32
		randomString, err := service.GenerateRandomString(stringLength)
		if err != nil {
			t.Fatalf("Failed to generate random string: %v", err)
		}

		if len(randomString) != stringLength {
			t.Errorf("Expected random string length %d, got %d", stringLength, len(randomString))
		}
	})

	t.Run("GenerateRandomString - Valid Characters", func(t *testing.T) {
		allValidChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}<>?|~`"
		stringLength := 32
		randomString, err := service.GenerateRandomString(stringLength)
		if err != nil {
			t.Fatalf("Failed to generate random string: %v", err)
		}

		for _, char := range randomString {
			if !strings.Contains(allValidChars, string(char)) {
				t.Errorf("Random string contains invalid character: %q", char)
			}
		}
	})

	t.Run("GenerateRandomString - Zero Length", func(t *testing.T) {
		_, err := service.GenerateRandomString(0)
		if err == nil {
			t.Fatal("GenerateRandomString should have returned an error as length is zero")
		}
	})

	t.Run("GenerateRandomString - Uniqueness", func(t *testing.T) {
		stringLength := 32
		generatedStrings := make([]string, 1000)

		for i := 0; i < len(generatedStrings); i++ {
			randomString, err := service.GenerateRandomString(stringLength)
			if err != nil {
				t.Fatalf("Failed to generate random string: %v", err)
			}
			generatedStrings[i] = randomString
		}

		charCounts := make(map[rune]int)
		for _, str := range generatedStrings {
			for _, char := range str {
				charCounts[char]++
			}
		}

		entropy := calculateEntropy(charCounts, stringLength*len(generatedStrings))
		t.Logf("Calculated entropy: %.2f", entropy)

		expectedEntropy := math.Log2(float64(len("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}<>?|~`")))
		if entropy < expectedEntropy*0.75 {
			t.Errorf("Entropy too low: %.2f. Randomness insufficient.", entropy)
		}
	})
}

func calculateEntropy(charCounts map[rune]int, totalChars int) float64 {
	var entropy float64
	for _, count := range charCounts {
		probability := float64(count) / float64(totalChars)
		if probability > 0 {
			entropy -= probability * math.Log2(probability)
		}
	}
	return entropy
}

func BenchmarkGenerateRandomString(b *testing.B) {
	service := services.GetCryptoService()

	for i := 0; i < b.N; i++ {
		_, _ = service.GenerateRandomString(32)
	}
}
