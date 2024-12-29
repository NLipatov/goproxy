package application

type CryptoService interface {
	GenerateSalt() ([]byte, error)
	HashValue(value string, salt []byte) ([]byte, error)
	ValidateHash(hash []byte, salt []byte, password string) bool
}
