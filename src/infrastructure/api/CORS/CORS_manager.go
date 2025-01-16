package CORS

import (
	"log"
	"net/http"
	"os"
	"strings"
)

type CORSManager struct {
}

func NewCORSManager() CORSManager {
	return CORSManager{}
}

func (c *CORSManager) AddCORS(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if c.IsOriginAllowed(origin, c.GetAllowedOrigins()) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler.ServeHTTP(w, r)
	})
}

func (c *CORSManager) GetAllowedOrigins() []string {
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		log.Fatalf("ALLOWED_ORIGINS environment variable is not set")
	}

	return strings.Split(allowedOrigins, ",")
}

func (c *CORSManager) IsOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowedOrigin := range allowedOrigins {
		if origin == allowedOrigin {
			return true
		}
	}
	return false
}
