package infrastructure

import (
	"encoding/json"
	"errors"
	"fmt"
	"goproxy/application"
	"goproxy/infrastructure/dto"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type HttpRestApiListener struct {
	userUseCases application.UserUseCasesContract
}

func NewHttpRestApiListener(useCases application.UserUseCasesContract) *HttpRestApiListener {
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
	case "DELETE":
		l.DeleteUser(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (l *HttpRestApiListener) PostUser(w http.ResponseWriter, r *http.Request) {
	var dto dto.PostUserCommand
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 512))
	if err := decoder.Decode(&dto); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(r.Body)

	command, err := dto.ToCreateUserCommand()
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	id, err := l.userUseCases.Create(command)
	if err != nil {
		if strings.Contains(err.Error(), "violates unique constraint") {
			respondWithError(w, http.StatusBadRequest, "User with such username already exists")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "")
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(fmt.Sprintf("%d", id)))
}

func (l *HttpRestApiListener) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := getUserIdFromQuery(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := l.userUseCases.GetById(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondWithError(w, http.StatusNotFound, "user not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "")
		return
	}

	dto := dto.FromUser(user)
	serialized, err := json.Marshal(dto)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(serialized)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "")
		return
	}
}

func (l *HttpRestApiListener) DeleteUser(w http.ResponseWriter, r *http.Request) {
	var dto dto.DeleteUserCommand
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 512))
	if err := decoder.Decode(&dto); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	command, err := dto.ToDeleteUserCommandDTO()
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	userDeletionErr := l.userUseCases.Delete(command)
	if userDeletionErr != nil {
		respondWithError(w, http.StatusInternalServerError, "")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func getUserIdFromQuery(r *http.Request) (int, error) {
	strId := r.URL.Query().Get("id")
	if strId == "" {
		return -1, errors.New("'id' query parameter is required")
	}

	id, err := strconv.Atoi(strId)
	return id, err
}

func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	http.Error(w, message, statusCode)
	log.Printf("%d: %s", statusCode, message)
}
