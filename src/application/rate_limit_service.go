package application

type RateLimiterService interface {
	Allow(userID int, target string, tokens int64) bool
	Done(userID int, target string)
	Stop()
}
