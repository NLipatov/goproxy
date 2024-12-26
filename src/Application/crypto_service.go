package Application

type CryptoService interface {
	GenerateSalt() ([]byte, error)
	HashValue(value, salt string) ([]byte, error)
	ValidateHash(value string, salt, hash []byte) bool
}
