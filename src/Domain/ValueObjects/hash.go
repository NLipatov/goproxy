package ValueObjects

import "errors"

type Hash struct {
	Value []byte
}

func NewHash(hash []byte) (Hash, error) {
	if len(hash) == 0 {
		return Hash{}, errors.New("hash cannot be empty")
	}

	return Hash{
		Value: hash,
	}, nil
}
