package Infrastructure

import (
	"encoding/json"
	"fmt"
	"goproxy/Application"
	"log"
	"net/http"
	"strconv"
)

type HttpRestApiListener struct {
	userUseCases *Application.UserUseCases
}

func NewHttpRestApiListener(useCases *Application.UserUseCases) *HttpRestApiListener {
	return &HttpRestApiListener{
		userUseCases: useCases,
	}
}

func (l *HttpRestApiListener) ServePort(port string) error {
	log.Printf("REST API is serving port %s", port)

	http.HandleFunc("/users", l.handleUsers)

	return http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}

func (l *HttpRestApiListener) handleUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		l.PostUser(w, r)
	case "GET":
		l.GetUser(w, r)
	case "UPDATE":
		l.UpdateUser(w, r)
	case "DELETE":
		l.DeleteUser(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (l *HttpRestApiListener) PostUser(w http.ResponseWriter, r *http.Request) {

}

func (l *HttpRestApiListener) GetUser(w http.ResponseWriter, r *http.Request) {
	strId := r.URL.Query().Get("id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	user, err := l.userUseCases.GetById(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	}

	serialized, err := json.Marshal(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	_, _ = w.Write(serialized)
}

func (l *HttpRestApiListener) UpdateUser(w http.ResponseWriter, r *http.Request) {

}

func (l *HttpRestApiListener) DeleteUser(w http.ResponseWriter, r *http.Request) {

}
