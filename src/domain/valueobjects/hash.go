package valueobjects

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

type Argon2idHash struct {
	Value string
}

func NewHash(hash string) (Argon2idHash, error) {
	if len(hash) == 0 {
		return Argon2idHash{}, errors.New("hash cannot be empty")
	}

	if err := validateArgon2idHash(hash); err != nil {
		return Argon2idHash{}, err
	}

	return Argon2idHash{
		Value: hash,
	}, nil
}

func validateArgon2idHash(hash string) error {
	// expected format: $argon2id$v=19$m=65536,t=3,p=2$<salt_base64>$<hash_base64>
	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		return fmt.Errorf("invalid hash format: expected 6 parts, got %d", len(parts))
	}

	if parts[1] != "argon2id" {
		return fmt.Errorf("invalid algorithm: expected 'argon2id', got '%s'", parts[1])
	}

	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return fmt.Errorf("invalid version format: %v", err)
	}
	if version != argon2Version {
		return fmt.Errorf("unsupported Argon2 version: expected %d, got %d", argon2Version, version)
	}

	// m, t, p parameter check
	var memory, iterations uint32
	var parallelism uint8
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
	if err != nil {
		return fmt.Errorf("invalid parameters format: %v", err)
	}

	if memory == 0 || iterations == 0 || parallelism == 0 {
		return errors.New("memory, iterations, and parallelism must be greater than zero")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return fmt.Errorf("invalid base64 salt: %v", err)
	}
	if len(salt) == 0 {
		return errors.New("salt cannot be empty")
	}

	hashBytes, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return fmt.Errorf("invalid base64 hash: %v", err)
	}
	if len(hashBytes) == 0 {
		return errors.New("hash cannot be empty")
	}

	return nil
}

const argon2Version = 19
