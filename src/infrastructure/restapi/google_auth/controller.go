package google_auth

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

type GoogleAuthController struct {
	authService *GoogleAuthService
	port        int
}

func NewGoogleAuthController(service *GoogleAuthService) *GoogleAuthController {
	return &GoogleAuthController{
		authService: service,
	}
}

func (g *GoogleAuthController) Listen(port int) {
	g.port = port

	mux := http.NewServeMux()
	mux.HandleFunc("/auth/login", g.authService.handleGoogleLogin)
	mux.HandleFunc("/auth/callback", g.authService.handleGoogleCallback)
	mux.HandleFunc("/auth/user-info", g.authService.GetUserInfo)

	corsHandler := g.AddCORS(mux, getAllowedOrigins())

	log.Println(fmt.Sprintf("Server is running on http://localhost:%d", g.port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", g.port), corsHandler))
}

func (g *GoogleAuthController) AddCORS(handler http.Handler, allowedOrigins []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if isOriginAllowed(origin, allowedOrigins) {
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

func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowedOrigin := range allowedOrigins {
		if origin == allowedOrigin {
			return true
		}
	}
	return false
}

func getAllowedOrigins() []string {
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		log.Fatalf("ALLOWED_ORIGINS environment variable is not set")
	}

	return strings.Split(allowedOrigins, ",")
}
