package application

type Jwt interface {
	Generate(secret string) (string, error)
	Validate(secret string, token string) (bool, error)
}
