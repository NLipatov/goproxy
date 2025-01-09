package infraerrs

type RateLimitExceededError struct {
}

func (rle RateLimitExceededError) Error() string {
	return "rate limit exceeded"
}
