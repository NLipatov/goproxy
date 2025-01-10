package valueobjects

import (
	"fmt"
	"regexp"
)

type Email struct {
	value string
}

var emailRegex = regexp.MustCompile(`^(?i)` +
	`[a-z0-9](?:[a-z0-9._%+\-]*[a-z0-9])?` + // local part
	`@` +
	`(?:` +
	// regular domain labels + TLD
	`(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+(?:[a-z]{2,12}|xn--[a-z0-9]+)` +
	`|` +
	// or plain IPv4 address
	`(?:\d{1,3}\.){3}\d{1,3}` +
	`|` +
	// or plain IPv6 address
	`\[[0-9A-Fa-f:]+\]` +
	`)$`)

func ParseEmailFromString(value string) (Email, error) {
	if !emailRegex.MatchString(value) {
		return Email{}, fmt.Errorf("invalid email: %s", value)
	}

	return Email{
		value: value,
	}, nil
}

func (email Email) String() string {
	return email.value
}
