// chacha20utils.go
package services

import (
	"crypto/rand"
	"errors"
	"golang.org/x/crypto/chacha20"
	"log"
	"sync"
)

// cipherPool is a pool of reusable ChaCha20 cipher instances.
var (
	cipherPool = sync.Pool{
		New: func() interface{} {
			key, err := generateRandomBytes(chacha20.KeySize)
			if err != nil {
				log.Printf("Error generating ChaCha20 key: %v", err)
				return nil
			}

			nonce, err := generateRandomBytes(chacha20.NonceSize)
			if err != nil {
				log.Printf("Error generating ChaCha20 nonce: %v", err)
				return nil
			}

			cipher, err := chacha20.NewUnauthenticatedCipher(key, nonce)
			if err != nil {
				log.Printf("Error creating ChaCha20 cipher: %v", err)
				return nil
			}

			return cipher
		},
	}
)

// newChaChaCipher retrieves a ChaCha20 cipher from the pool.
// Returns an error if the cipher cannot be created.
func newChaChaCipher() (*chacha20.Cipher, error) {
	cipher, ok := cipherPool.Get().(*chacha20.Cipher)
	if !ok || cipher == nil {
		return nil, errors.New("failed to create ChaCha20 cipher from pool")
	}
	return cipher, nil
}

// generateRandomBytes generates random bytes of the specified length using crypto/rand.
func generateRandomBytes(length int) ([]byte, error) {
	if length <= 0 {
		return nil, errors.New("length must be greater than zero")
	}

	randomBytes := make([]byte, length)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, errors.New("failed to generate random bytes")
	}

	return randomBytes, nil
}

// generateRandomBytesUsingCipher generates random bytes of the specified length using ChaCha20.
func generateRandomBytesUsingCipher(cipher *chacha20.Cipher, length int) ([]byte, error) {
	if cipher == nil {
		return nil, errors.New("cipher cannot be nil")
	}

	if length <= 0 {
		return nil, errors.New("length must be greater than zero")
	}

	randomBytes := make([]byte, length)
	// XORKeyStream transforms zero bytes into random bytes
	cipher.XORKeyStream(randomBytes, make([]byte, length))

	return randomBytes, nil
}

// getCipher retrieves a ChaCha20 cipher from the pool or creates a new one if necessary.
func getCipher() (*chacha20.Cipher, error) {
	cipher, err := newChaChaCipher()
	if err != nil {
		// Attempt to create a new cipher if pool retrieval failed
		key, errKey := generateRandomBytes(chacha20.KeySize)
		if errKey != nil {
			return nil, errKey
		}

		nonce, errNonce := generateRandomBytes(chacha20.NonceSize)
		if errNonce != nil {
			return nil, errNonce
		}

		cipher, err = chacha20.NewUnauthenticatedCipher(key, nonce)
		if err != nil {
			return nil, err
		}

		// Optionally, put the new cipher back into the pool for reuse
		cipherPool.Put(cipher)
	}

	return cipher, nil
}

// GenerateSecureRandomBytes generates secure random bytes using ChaCha20.
func GenerateSecureRandomBytes(length int) ([]byte, error) {
	cipher, err := getCipher()
	if err != nil {
		return nil, err
	}

	randomBytes, err := generateRandomBytesUsingCipher(cipher, length)
	if err != nil {
		return nil, err
	}

	// Return the cipher back to the pool for reuse
	cipherPool.Put(cipher)

	return randomBytes, nil
}
