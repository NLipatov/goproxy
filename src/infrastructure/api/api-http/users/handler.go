package users

import (
	"encoding/json"
	"errors"
	"fmt"
	"goproxy/application/use_cases"
	"goproxy/domain/aggregates"
	"goproxy/infrastructure/dto"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	userUseCases use_cases.UserUseCasesContract
}

func (l *Handler) PostUser(w http.ResponseWriter, r *http.Request) {
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

func (l *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, idErr := getUserIdFromQuery(r)
	if idErr == nil {
		user, err := l.userUseCases.GetById(id)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				respondWithError(w, http.StatusNotFound, "user not found")
				return
			}
			respondWithError(w, http.StatusInternalServerError, "")
			return
		}

		respondErr := respondWithUserDto(user, w)
		if respondErr != nil {
			respondWithError(w, http.StatusInternalServerError, "")
		}
		return
	}

	email, emailErr := getUserEmailFromQuery(r)
	if emailErr == nil {
		user, err := l.userUseCases.GetByEmail(email)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				respondWithError(w, http.StatusNotFound, "user not found")
				return
			}
			respondWithError(w, http.StatusInternalServerError, "")
			return
		}

		respondErr := respondWithUserDto(user, w)
		if respondErr != nil {
			respondWithError(w, http.StatusInternalServerError, "")
		}
		return
	}

	respondWithError(w, http.StatusBadRequest, "either 'id' or 'email' must be provided")
}

func (l *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
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
		return -1, errors.New("no 'id' query parameter specified")
	}

	id, err := strconv.Atoi(strId)
	return id, err
}

func getUserEmailFromQuery(r *http.Request) (string, error) {
	email := r.URL.Query().Get("email")
	if email == "" {
		return "", errors.New("no 'email' query parameter specified")
	}

	return email, nil
}

func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	http.Error(w, message, statusCode)
	log.Printf("%d: %s", statusCode, message)
}

func respondWithUserDto(user aggregates.User, w http.ResponseWriter) error {
	cmd := dto.FromUser(user)
	serialized, err := json.Marshal(cmd)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "")
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(serialized)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "")
		return err
	}

	return nil
}
