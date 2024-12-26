package services

import (
	"crypto/rand"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

type CryptoService struct {
	saltLength int
}

func NewCryptoService(saltLength int) *CryptoService {
	return &CryptoService{
		saltLength: saltLength,
	}
}

func (c *CryptoService) GenerateSalt() ([]byte, error) {
	salt := make([]byte, c.saltLength)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, errors.New("failed to generate salt")
	}
	return salt, nil
}

func (c *CryptoService) HashValue(value, salt string) ([]byte, error) {
	saltedValue := value + salt

	hash, err := bcrypt.GenerateFromPassword([]byte(saltedValue), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash value")
	}

	return hash, nil
}

func (c *CryptoService) ValidateHash(value string, salt, hash []byte) bool {
	saltedValue := append([]byte(value), salt...)
	err := bcrypt.CompareHashAndPassword(hash, saltedValue)
	return err == nil
}
