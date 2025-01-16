package google_auth

import (
	"fmt"
	"goproxy/infrastructure/api/CORS"
	"log"
	"net/http"
)

type GoogleAuthController struct {
	corsManager CORS.CORSManager
	authService *GoogleAuthService
	port        int
}

func NewGoogleAuthController(service *GoogleAuthService) *GoogleAuthController {
	return &GoogleAuthController{
		authService: service,
		corsManager: CORS.NewCORSManager(),
	}
}

func (g *GoogleAuthController) Listen(port int) {
	g.port = port

	mux := http.NewServeMux()
	mux.HandleFunc("/auth/login", g.authService.handleGoogleLogin)
	mux.HandleFunc("/auth/callback", g.authService.handleGoogleCallback)
	mux.HandleFunc("/auth/user-info", g.authService.GetUserInfo)
	mux.HandleFunc("/auth/reset-password", g.authService.ResetPassword)

	corsHandler := g.corsManager.AddCORS(mux)

	log.Println(fmt.Sprintf("Server is running on http://localhost:%d", g.port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", g.port), corsHandler))
}
