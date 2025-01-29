package services

import (
	"testing"
	"time"
)

func TestJWTService_GenerateAndValidate(t *testing.T) {
	jwtService := NewJwt()
	secret := "my_test_secret"
	ttl := 8 * time.Hour
	claims := map[string]string{
		"id":    "42",
		"role":  "admin",
		"email": "test@example.com",
	}

	token, err := jwtService.Generate(secret, ttl, claims)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	isValid, err := jwtService.Validate(secret, token)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	if !isValid {
		t.Fatalf("Expected token to be valid but it was invalid")
	}
}

func TestJWTService_ValidateExpiredToken(t *testing.T) {
	jwtService := NewJwt()
	secret := "my_test_secret"

	ttl := time.Nanosecond //expires immediately
	claims := map[string]string{
		"id": "99",
	}

	token, err := jwtService.Generate(secret, ttl, claims)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	time.Sleep(ttl)

	isValid, err := jwtService.Validate(secret, token)
	if isValid || err == nil {
		t.Fatalf("Expected expired token to be invalid, but validation passed")
	}
}

func TestJWTService_ValidateWithWrongSecret(t *testing.T) {
	jwtService := NewJwt()
	correctSecret := "correct_secret"
	wrongSecret := "wrong_secret"

	ttl := 8 * time.Hour
	claims := map[string]string{
		"id": "100",
	}

	token, err := jwtService.Generate(correctSecret, ttl, claims)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	isValid, err := jwtService.Validate(wrongSecret, token)
	if isValid || err == nil {
		t.Fatalf("Expected token validation to fail with wrong secret, but it passed")
	}
}
