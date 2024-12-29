package valueobjects

type CredentialsType int

const (
	Basic CredentialsType = iota
)

type Credentials interface {
	Type() CredentialsType
}

type BasicCredentials struct {
	Username string
	Password string
}

func (c BasicCredentials) Type() CredentialsType {
	return Basic
}
