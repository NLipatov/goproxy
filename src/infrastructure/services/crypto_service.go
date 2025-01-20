package services

import (
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"math"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/mem"
	"golang.org/x/crypto/argon2"
	"goproxy/application"
)

// CryptoServiceImpl implements the CryptoService interface.
type CryptoServiceImpl struct {
	// Argon2id parameters
	Memory      uint32 // Memory in KB
	Iterations  uint32 // Number of iterations
	Parallelism uint8  // Parallelism level
	SaltLength  uint32 // Length of the salt in bytes
	KeyLength   uint32 // Length of the final hash
}

var (
	cryptoServiceInstance application.CryptoService
	once                  sync.Once
)

// GetCryptoService returns the singleton instance of CryptoService.
func GetCryptoService() application.CryptoService {
	once.Do(func() {
		memory, iterations, parallelism := determineArgon2idParameters()

		cryptoServiceInstance = &CryptoServiceImpl{
			Memory:      memory,
			Iterations:  iterations,
			Parallelism: parallelism,
			SaltLength:  16, // Recommended salt length
			KeyLength:   32, // Recommended hash length
		}
	})
	return cryptoServiceInstance
}

// determineArgon2idParameters determines the Argon2id parameters based on the system's available resources.
func determineArgon2idParameters() (memory uint32, iterations uint32, parallelism uint8) {
	// Set parallelism based on the number of CPU cores.
	parallelism = uint8(runtime.NumCPU())

	// Detect available system memory using gopsutil.
	vm, err := mem.VirtualMemory()
	if err != nil {
		log.Printf("Failed to get virtual memory info: %v. Using default memory value.", err)
		// Default to 64 MB if memory detection fails.
		memory = 64 * 1024
	} else {
		// Allocate 1% of total memory for hashing, but not less than 64 MB and not more than 1 GB.
		allocatedMemoryKB := vm.Total / 100 / 1024 // 1% of total memory in KB
		memory = uint32(math.Max(float64(64*1024), math.Min(float64(allocatedMemoryKB), float64(1024*1024))))
	}

	// Determine the number of iterations to achieve the target hashing time (e.g., ~100ms).
	iterations = 3
	targetDuration := 100 * time.Millisecond

	for {
		start := time.Now()
		_ = argon2.IDKey([]byte("test-password"), make([]byte, memory), iterations, memory, parallelism, 32)
		duration := time.Since(start)
		if duration >= targetDuration || iterations >= 10 {
			break
		}
		iterations++
	}

	return
}

// HashValue hashes the provided value using Argon2id.
func (c *CryptoServiceImpl) HashValue(value string) (string, error) {
	if value == "" {
		return "", errors.New("value cannot be empty")
	}

	// Generate salt
	salt, err := GenerateSecureRandomBytes(int(c.SaltLength))
	if err != nil {
		return "", err
	}

	// Create hash using Argon2id
	hash := argon2.IDKey([]byte(value), salt, c.Iterations, c.Memory, c.Parallelism, c.KeyLength)

	// Encode parameters, salt, and hash into a string
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Format: $argon2id$v=19$m=65536,t=3,p=2$<salt>$<hash>
	fullHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, c.Memory, c.Iterations, c.Parallelism, b64Salt, b64Hash)

	return fullHash, nil
}

// ValidateHash verifies the correspondence between the hash and the password.
func (c *CryptoServiceImpl) ValidateHash(fullHash string, password string) bool {
	// Parse the hash
	params, salt, hash, err := decodeFullHash(fullHash)
	if err != nil {
		return false
	}

	// Generate hash for the entered password
	comparisonHash := argon2.IDKey([]byte(password), salt, params.Iterations, params.Memory, params.Parallelism, uint32(len(hash)))

	// Compare hashes
	return subtle.ConstantTimeCompare(hash, comparisonHash) == 1
}

// GenerateRandomString generates a random string of the specified length from a set of characters using ChaCha20.
func (c *CryptoServiceImpl) GenerateRandomString(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("length must be greater than zero")
	}

	const charSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}<>?|~`"
	charSetLen := int64(len(charSet))

	// Calculate how many random bytes are needed
	randomBytesNeeded := length

	// Generate random bytes using ChaCha20
	randomBytes, err := GenerateSecureRandomBytes(randomBytesNeeded)
	if err != nil {
		return "", err
	}

	result := make([]byte, length)
	for i := 0; i < length; i++ {
		idx := int(randomBytes[i] % byte(charSetLen))
		result[i] = charSet[idx]
	}

	return string(result), nil
}

// decodeFullHash parses the full Argon2id hash and extracts the parameters, salt, and hash.
func decodeFullHash(fullHash string) (params *Argon2Params, salt, hash []byte, err error) {
	// Expected format: $argon2id$v=19$m=65536,t=3,p=2$<salt>$<hash>
	parts := strings.Split(fullHash, "$")
	if len(parts) != 6 {
		return nil, nil, nil, errors.New("invalid hash format")
	}

	if parts[1] != "argon2id" {
		return nil, nil, nil, errors.New("unsupported algorithm")
	}

	var version int
	_, err = fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	}

	if version != argon2.Version {
		return nil, nil, nil, errors.New("incompatible Argon2 version")
	}

	var memory, iterations uint32
	var parallelism uint8
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, err
	}

	hash, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, err
	}

	params = &Argon2Params{
		Memory:      memory,
		Iterations:  iterations,
		Parallelism: parallelism,
	}

	return params, salt, hash, nil
}

// Argon2Params contains the parameters used for Argon2id hashing.
type Argon2Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
}
