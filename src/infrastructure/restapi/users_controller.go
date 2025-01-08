package restapi

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

type UsersController struct {
	userUseCases application.UserUseCasesContract
}

func NewUsersController(useCases application.UserUseCasesContract) *UsersController {
	return &UsersController{
		userUseCases: useCases,
	}
}

func (l *UsersController) ServePort(port string) error {
	log.Printf("Users REST API is serving port %s", port)

	http.HandleFunc("/users", l.handleUsers)

	return http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}

func (l *UsersController) handleUsers(w http.ResponseWriter, r *http.Request) {
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

func (l *UsersController) PostUser(w http.ResponseWriter, r *http.Request) {
	var cmd dto.PostUserCommand
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 512))
	if err := decoder.Decode(&cmd); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(r.Body)

	command, err := cmd.ToCreateUserCommand()
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

func (l *UsersController) GetUser(w http.ResponseWriter, r *http.Request) {
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

	cmd := dto.FromUser(user)
	serialized, err := json.Marshal(cmd)
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

func (l *UsersController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	var cmd dto.DeleteUserCommand
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 512))
	if err := decoder.Decode(&cmd); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	command, err := cmd.ToDeleteUserCommandDTO()
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	userDeletionErr := l.userUseCases.Delete(command)
	if userDeletionErr != nil {
		if strings.Contains(userDeletionErr.Error(), "not found") {
			respondWithError(w, http.StatusNotFound, "user not found")
			return

		}
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
