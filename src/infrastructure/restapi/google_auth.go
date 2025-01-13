package restapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func getPort() int {
	portEnvValue := os.Getenv("SERVE_PORT")
	if portEnvValue == "" {
		log.Fatalf("SERVE_PORT environment variable is not set")
	}

	port, err := strconv.Atoi(portEnvValue)
	if err != nil {
		log.Fatalf("SERVE_PORT environment variable is not a number")
	}

	return port
}

func getGoogleOauthConfig() *oauth2.Config {
	clientSecret := os.Getenv("GOOGLE_AUTH_CLIENT_SECRET")
	if clientSecret == "" {
		log.Fatalf("GOOGLE_AUTH_CLIENT_SECRET environment variable is not set")
	}

	clientId := os.Getenv("GOOGLE_AUTH_CLIENT_ID")
	if clientId == "" {
		log.Fatalf("GOOGLE_AUTH_CLIENT_ID environment variable is not set")
	}

	return &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  fmt.Sprintf("http://localhost:%d/auth/callback", getPort()),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}
}

func HandleGoogleAuth() {
	http.HandleFunc("/auth/login", handleGoogleLogin)
	http.HandleFunc("/auth/callback", handleGoogleCallback)
	log.Println(fmt.Sprintf("Server is running on http://localhost:%d", getPort()))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", getPort()), nil))
}

func handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := getGoogleOauthConfig().AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("state") != "state-token" {
		http.Error(w, "State mismatch", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	token, err := getGoogleOauthConfig().Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	client := getGoogleOauthConfig().Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	userInfo := struct {
		Email         string `json:"email"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		VerifiedEmail bool   `json:"verified_email"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		http.Error(w, "Failed to parse user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	redirectURL := fmt.Sprintf("http://localhost:5173/?email=%s&name=%s", userInfo.Email, userInfo.Name)

	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}
