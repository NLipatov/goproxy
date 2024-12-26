package ValueObjects

import "errors"

type Salt struct {
	Value []byte
}

func NewSalt(salt []byte) (Salt, error) {
	if len(salt) == 0 {
		return Salt{}, errors.New("salt is empty")
	}

	return Salt{
		Value: salt,
	}, nil
}
