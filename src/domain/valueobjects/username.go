package valueobjects

import (
	"errors"
	"strings"
	"unicode"
)

type Username struct {
	Value string
}

func NewUsernameFromString(username string) (Username, error) {
	chars := []rune(username)

	if len(chars) == 0 {
		return Username{}, errors.New("username cannot be empty string")
	}

	for _, v := range chars {
		if unicode.IsUpper(v) {
			return Username{}, errors.New("username must not contain uppercase letters")
		}
		if unicode.IsSpace(v) {
			return Username{}, errors.New("username must not contain spaces")
		}
	}

	return Username{
		Value: username,
	}, nil
}

func NormalizeUsername(username string) string {
	username = strings.ToLower(username)
	return strings.ReplaceAll(username, " ", "_")
}
