package services

import (
	"encoding/json"
	"fmt"
	"goproxy/application"
	"goproxy/domain"
	"goproxy/domain/aggregates"
	"goproxy/domain/events"
	"goproxy/domain/valueobjects"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const defaultValidateCacheTTL = time.Second * 10

type validateResult struct {
	result bool
	err    error
}

type AuthService struct {
	cryptoService    application.CryptoService
	validateCache    application.CacheWithTTL[validateResult]
	validateCacheTTL time.Duration
	messageBus       application.MessageBusService
	once             sync.Once
}

func NewAuthService(cryptoService application.CryptoService, messageBusService application.MessageBusService) *AuthService {
	validateCacheTTL := defaultValidateCacheTTL
	validateTtlEnv := os.Getenv("AUTH_SERVICE_VALIDATE_TTL_MS")
	if validateTtlEnv != "" {
		ttlMillis, err := strconv.Atoi(validateTtlEnv)
		if err == nil {
			validateCacheTTL = time.Duration(ttlMillis) * time.Millisecond
		}
	}

	service := AuthService{
		cryptoService:    cryptoService,
		validateCache:    NewMapCacheWithTTL[validateResult](),
		validateCacheTTL: validateCacheTTL,
		messageBus:       messageBusService,
	}

	service.startProcessingEvents()

	return &service
}

func (a *AuthService) AuthorizeBasic(user aggregates.User, credentials valueobjects.BasicCredentials) (bool, error) {
	cacheKey := fmt.Sprintf("%s:%x", credentials.Username, user.PasswordHash())
	cached, cachedErr := a.validateCache.Get(cacheKey)
	if cachedErr == nil {
		return cached.result, cached.err
	}

	isPasswordValid := a.cryptoService.ValidateHash(user.PasswordHash(), credentials.Password)
	if !isPasswordValid {
		return false, fmt.Errorf("invalid credentials")
	}

	_ = a.validateCache.Set(cacheKey, validateResult{true, nil})
	_ = a.validateCache.Expire(cacheKey, a.validateCacheTTL)

	return true, nil
}

func (a *AuthService) startProcessingEvents() {
	a.once.Do(func() {

		go a.processEvents()
	})
}

func (a *AuthService) processEvents() {
	defer func(messageBus application.MessageBusService) {
		_ = messageBus.Close()
	}(a.messageBus)

	topics := []string{fmt.Sprintf("%s", domain.PROXY)}
	err := a.messageBus.Subscribe(topics)
	if err != nil {
		log.Fatalf("Failed to subscribe to topics: %s", err)
	}

	log.Printf("Subscribed to topics: %s", strings.Join(topics, ", "))

	for {
		event, consumeErr := a.messageBus.Consume()
		if consumeErr != nil {
			log.Printf("failed to consume from message bus: %s", consumeErr)
		}

		if event == nil {
			log.Printf("received nil event from message bus")
			continue
		}

		if event.EventType.Value() == "UserPasswordChangedEvent" {
			var userPasswordChangedEvent events.UserPasswordChangedEvent
			deserializationErr := json.Unmarshal([]byte(event.Payload), &userPasswordChangedEvent)
			if deserializationErr != nil {
				log.Printf("failed to deserialize user password changed event: %s", deserializationErr)
			}

			err = a.validateCache.Delete(userPasswordChangedEvent.Username)
			if err == nil {
				log.Printf("user %s removed from validation cache", userPasswordChangedEvent.Username)
			} else {
				log.Printf("user %s was not removed from validation cache: %s", userPasswordChangedEvent.Username, err)
			}
		}
	}
}
