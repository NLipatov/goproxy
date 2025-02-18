package crypto_cloud_billing_errors

type TokenEmailExtractionErr struct {
	err error
}

func NewTokenEmailExtractionErr(err error) *TokenEmailExtractionErr {
	return &TokenEmailExtractionErr{
		err: err,
	}
}

func (e *TokenEmailExtractionErr) Error() string {
	return e.err.Error()
}
