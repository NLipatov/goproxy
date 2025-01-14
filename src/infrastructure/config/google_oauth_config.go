package config

import (
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"log"
	"os"
	"strconv"
)

type GoogleOauthConfigProvider struct {
	Config oauth2.Config
	Port   int
}

func NewGoogleOauthConfig() GoogleOauthConfigProvider {
	clientSecret := os.Getenv("GOOGLE_AUTH_CLIENT_SECRET")
	if clientSecret == "" {
		log.Fatalf("GOOGLE_AUTH_CLIENT_SECRET environment variable is not set")
	}

	clientId := os.Getenv("GOOGLE_AUTH_CLIENT_ID")
	if clientId == "" {
		log.Fatalf("GOOGLE_AUTH_CLIENT_ID environment variable is not set")
	}

	hostEnvVar := os.Getenv("GOOGLE_AUTH_HOST")
	if hostEnvVar == "" {
		log.Fatalf("GOOGLE_AUTH_HOST environment variable is not set")
	}

	host := hostEnvVar

	portEnvValue := os.Getenv("GOOGLE_AUTH_PORT")
	if portEnvValue == "" {
		log.Fatalf("GOOGLE_AUTH_PORT environment variable is not set")
	}

	port, err := strconv.Atoi(portEnvValue)
	if err != nil {
		log.Fatalf("GOOGLE_AUTH_PORT environment variable is not a number")
	}

	config := oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  fmt.Sprintf("%s:%d/auth/callback", host, port),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}

	return GoogleOauthConfigProvider{
		Config: config,
		Port:   port,
	}
}
