package services

import (
	"crypto/rand"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"goproxy/Application"
)

type CryptoService struct {
	saltLength int
}

func NewCryptoService(saltLength int) Application.CryptoService {
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

func (c *CryptoService) HashValue(value string, salt []byte) ([]byte, error) {
	saltedValue := value + string(salt)

	hash, err := bcrypt.GenerateFromPassword([]byte(saltedValue), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash value")
	}

	return hash, nil
}

func (c *CryptoService) ValidateHash(hash []byte, salt []byte, password string) bool {
	saltedValue := password + string(salt)
	err := bcrypt.CompareHashAndPassword(hash, []byte(saltedValue))
	return err == nil
}
