package valueobjects

import "errors"

type Username struct {
	Value string
}

func NewUsernameFromString(username string) (Username, error) {
	if username == "" {
		return Username{}, errors.New("username cannot be empty string")
	}

	return Username{
		Value: username,
	}, nil
}
