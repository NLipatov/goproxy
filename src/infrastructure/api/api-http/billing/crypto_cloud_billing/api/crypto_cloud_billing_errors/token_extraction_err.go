package crypto_cloud_billing_errors

type TokenExtractionErr struct {
	err error
}

func NewTokenExtractionErr(err error) TokenExtractionErr {
	return TokenExtractionErr{
		err: err,
	}
}

func (e TokenExtractionErr) Error() string {
	return e.err.Error()
}
