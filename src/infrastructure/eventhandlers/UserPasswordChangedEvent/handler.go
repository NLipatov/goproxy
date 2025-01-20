package UserPasswordChangedEvent

import (
	"encoding/json"
	"goproxy/application"
	"goproxy/domain/events"
	"log"
)

type Handler[T any] struct {
	cache application.Cache[T]
}

func NewUserPasswordChangedEventHandler[T any](cache application.Cache[T]) application.EventHandler {
	return &Handler[T]{
		cache: cache,
	}
}

func (u *Handler[T]) Handle(payload string) error {
	var userPasswordChangedEvent events.UserPasswordChangedEvent
	deserializationErr := json.Unmarshal([]byte(payload), &userPasswordChangedEvent)
	if deserializationErr != nil {
		log.Printf("failed to deserialize user password changed event: %s", deserializationErr)
	}

	deleteErr := u.cache.Delete(userPasswordChangedEvent.Username)
	if deleteErr != nil {
		log.Printf("UserPasswordChangedEvent handling: cache key was not removed: %s", deleteErr)
	}

	return nil
}
