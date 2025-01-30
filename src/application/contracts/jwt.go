package contracts

import "time"

type Jwt interface {
	Generate(secret string, ttl time.Duration, claims map[string]string) (string, error)
	Validate(secret string, token string) (bool, error)
}
