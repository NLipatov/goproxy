package eventhandlers

import (
	"encoding/json"
	"goproxy/application"
	"goproxy/dal/repositories"
	"goproxy/domain/events"
	"log"
)

type UserPasswordChangedEventHandler struct {
	cache repositories.BigCacheUserRepositoryCache
}

func NewUserPasswordChangedEventHandler(cache repositories.BigCacheUserRepositoryCache) application.EventHandler {
	return &UserPasswordChangedEventHandler{
		cache: cache,
	}
}

func (u *UserPasswordChangedEventHandler) Handle(payload string) error {
	var userPasswordChangedEvent events.UserPasswordChangedEvent
	deserializationErr := json.Unmarshal([]byte(payload), &userPasswordChangedEvent)
	if deserializationErr != nil {
		log.Printf("failed to deserialize user password changed event: %s", deserializationErr)
	}

	err := u.cache.Delete(userPasswordChangedEvent.Username)
	if err == nil {
		log.Printf("user %s removed from user repository cache", userPasswordChangedEvent.Username)
	} else {
		log.Printf("user %s was not removed from user repository cache: %s", userPasswordChangedEvent.Username, err)
	}

	return nil
}
