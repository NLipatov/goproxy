package users

import (
	"fmt"
	"goproxy/application"
	"goproxy/infrastructure/api/CORS"
	"log"
	"net/http"
)

type Controller struct {
	handler     Handler
	corsManager CORS.CORSManager
	port        int
}

func NewUsersController(userUseCases application.UserUseCasesContract) *Controller {
	handler := Handler{
		userUseCases: userUseCases,
	}
	return &Controller{
		handler:     handler,
		corsManager: CORS.NewCORSManager(),
	}
}

func (l *Controller) Listen(port int) {
	l.port = port

	mux := http.NewServeMux()
	mux.HandleFunc("/users/get", l.handler.GetUser)
	mux.HandleFunc("/users/post", l.handler.PostUser)
	mux.HandleFunc("/users/delete", l.handler.DeleteUser)

	corsHandler := l.corsManager.AddCORS(mux)

	log.Println(fmt.Sprintf("Server is running on http://localhost:%d", l.port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", l.port), corsHandler))
}
