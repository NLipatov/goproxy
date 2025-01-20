package Cookie

import (
	"log"
	"net/http"
	"os"
	"time"
)

type CookieBuilder struct {
	env string
}

func NewCookieBuilder() CookieBuilder {
	env := os.Getenv("ENVIRONMENT")

	return CookieBuilder{
		env: env,
	}
}

func (g *CookieBuilder) BuildCookie(path, name, value string, ttl time.Duration) *http.Cookie {
	if path == "" {
		path = "/"
	}

	expiration := time.Now().Add(ttl)
	return &http.Cookie{
		Name:     name,
		Value:    value,
		HttpOnly: true,                 // Accessible only by HTTP, prevents XSS
		Secure:   g.cookieSecureFlag(), // Should use HTTPS only
		Path:     path,                 // Any path
		SameSite: g.cookieSameSite(),   // CSRF protection
		MaxAge:   int(ttl.Seconds()),   // Set expiration
		Expires:  expiration,           // Concrete expiration time
	}
}

func (g *CookieBuilder) cookieSecureFlag() bool {
	if g.env == "development" {
		log.Printf("DANGER: using unsecure cookie mode as env is development")
	}

	return g.env != "development"
}

func (g *CookieBuilder) cookieSameSite() http.SameSite {
	if g.env == "development" {
		log.Printf("DANGER: using unsecure same site LAX mode as env is development")
		return http.SameSiteLaxMode
	}

	return http.SameSiteStrictMode
}
