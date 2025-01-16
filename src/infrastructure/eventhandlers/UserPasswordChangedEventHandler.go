package eventhandlers

import (
	"encoding/json"
	"goproxy/application"
	"goproxy/domain/events"
	"log"
)

type UserPasswordChangedEventHandler[T any] struct {
	cache application.Cache[T]
}

func NewUserPasswordChangedEventHandler[T any](cache application.Cache[T]) application.EventHandler {
	return &UserPasswordChangedEventHandler[T]{
		cache: cache,
	}
}

func (u *UserPasswordChangedEventHandler[T]) Handle(payload string) error {
	var userPasswordChangedEvent events.UserPasswordChangedEvent
	deserializationErr := json.Unmarshal([]byte(payload), &userPasswordChangedEvent)
	if deserializationErr != nil {
		log.Printf("failed to deserialize user password changed event: %s", deserializationErr)
	}

	_ = u.cache.Delete(userPasswordChangedEvent.Username)

	return nil
}
