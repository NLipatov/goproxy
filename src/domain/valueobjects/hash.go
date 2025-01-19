package valueobjects

import (
	"errors"
)

type Argon2idHash struct {
	Value string
}

func NewHash(hash string) (Argon2idHash, error) {
	if len(hash) == 0 {
		return Argon2idHash{}, errors.New("hash cannot be empty")
	}

	return Argon2idHash{
		Value: hash,
	}, nil
}
