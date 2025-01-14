package services

import (
	"crypto/rand"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"goproxy/application"
	"math/big"
)

type CryptoServiceImpl struct {
	saltLength int
}

func NewCryptoService(saltLength int) application.CryptoService {
	return &CryptoServiceImpl{
		saltLength: saltLength,
	}
}

func (c *CryptoServiceImpl) GenerateSalt() ([]byte, error) {
	salt := make([]byte, c.saltLength)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, errors.New("failed to generate salt")
	}
	return salt, nil
}

func (c *CryptoServiceImpl) HashValue(value string, salt []byte) ([]byte, error) {
	saltedValue := value + string(salt)

	hash, err := bcrypt.GenerateFromPassword([]byte(saltedValue), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash value")
	}

	return hash, nil
}

func (c *CryptoServiceImpl) ValidateHash(hash []byte, salt []byte, password string) bool {
	saltedValue := password + string(salt)
	err := bcrypt.CompareHashAndPassword(hash, []byte(saltedValue))
	return err == nil
}

func (c *CryptoServiceImpl) GenerateRandomString(length int) (string, error) {
	lowercase := "abcdefghijklmnopqrstuvwxyz"
	uppercase := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits := "0123456789"
	specials := "!@#$%^&*()-_=+[]{}<>?"
	all := lowercase + uppercase + digits + specials

	password := make([]byte, length)
	categories := []string{lowercase, uppercase, digits, specials}
	for i, category := range categories {
		char, err := randomCharFromSet(category)
		if err != nil {
			return "", err
		}
		password[i] = char
	}

	for i := len(categories); i < length; i++ {
		char, err := randomCharFromSet(all)
		if err != nil {
			return "", err
		}
		password[i] = char
	}

	shuffle(password)

	return string(password), nil
}

func randomCharFromSet(set string) (byte, error) {
	index, err := rand.Int(rand.Reader, big.NewInt(int64(len(set))))
	if err != nil {
		return 0, fmt.Errorf("failed to generate random index: %v", err)
	}
	return set[index.Int64()], nil
}

func shuffle(password []byte) {
	for i := len(password) - 1; i > 0; i-- {
		j, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		password[i], password[j.Int64()] = password[j.Int64()], password[i]
	}
}
