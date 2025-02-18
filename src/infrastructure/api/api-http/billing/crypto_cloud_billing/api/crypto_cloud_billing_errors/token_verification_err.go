package crypto_cloud_billing_errors

type TokenVerificationErr struct {
	err error
}

func NewTokenVerificationErr(err error) *TokenVerificationErr {
	return &TokenVerificationErr{err: err}
}

func (e TokenVerificationErr) Error() string {
	return e.err.Error()
}
