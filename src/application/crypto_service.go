package application

type CryptoService interface {
	GenerateRandomString(length int) (string, error)
	HashValue(value string) (string, error)
	ValidateHash(fullHash, password string) bool
}
