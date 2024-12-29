package valueobjects

import "errors"

type Password struct {
	Value string
}

func NewPasswordFromString(value string) (Password, error) {
	if value == "" {
		return Password{}, errors.New("password cannot be empty string")
	}

	if len(value) < 8 {
		return Password{}, errors.New("password must be at least 8 characters")
	}

	return Password{value}, nil
}
