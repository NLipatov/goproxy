package services

import (
	"encoding/json"
	"fmt"
	"goproxy/application"
	"goproxy/domain/aggregates"
	"goproxy/domain/events"
	"goproxy/infrastructure/config"
	"log"
	"strings"
	"sync"
	"time"
)

type UserRestrictionService struct {
	localCache  application.CacheWithTTL[bool]
	messageBus  application.MessageBusService
	remoteCache application.CacheWithTTL[bool]
}

var remoteCacheLocks sync.Map

const userRestrictionServiceTTL = time.Minute * 1

func NewUserRestrictionService() *UserRestrictionService {
	remoteCache, redisCacheErr := NewRedisCache[bool]()
	if redisCacheErr != nil {
		log.Fatalf("failed to initialize redis cache: %v", redisCacheErr)
	}

	kafkaConfig, kafkaConfigErr := config.NewKafkaConfig(config.PROXY)
	if kafkaConfigErr != nil {
		log.Fatal(kafkaConfigErr)
	}

	messageBusService, err := NewKafkaService(kafkaConfig)
	if err != nil {
		log.Fatalf("failed to initialize kafka service: %s", err)
	}

	return &UserRestrictionService{
		localCache:  NewMapCacheWithTTL[bool](),
		messageBus:  messageBusService,
		remoteCache: remoteCache,
	}
}

func (u *UserRestrictionService) IsRestricted(user aggregates.User) bool {
	key := u.UserIdToKey(user.Id())

	// check local cache
	if cached, err := u.localCache.Get(key); err == nil {
		return cached
	}

	// set temporal data while waiting for sync with remote cache
	_ = u.setToLocalCacheWithTTL(user.Id())

	go func() {
		// check if another goroutine is checking this user in remote cache
		if _, loaded := remoteCacheLocks.LoadOrStore(key, true); loaded {
			return
		}
		defer remoteCacheLocks.Delete(key)

		if cached, err := u.remoteCache.Get(key); err == nil && cached {
			_ = u.AddToRestrictionList(user)
		} else if err != nil {
			log.Printf("Error accessing Redis for user %d: %v", user.Id(), err)
			_ = u.setToLocalCacheWithTTL(user.Id())
		}
	}()

	return false
}

func (u *UserRestrictionService) AddToRestrictionList(user aggregates.User) error {
	return u.setToLocalCacheWithTTL(user.Id())
}

func (u *UserRestrictionService) setToLocalCacheWithTTL(userId int) error {
	key := u.UserIdToKey(userId)
	setErr := u.localCache.Set(key, true)
	if setErr != nil {
		return setErr
	}

	expireErr := u.localCache.Expire(key, userRestrictionServiceTTL)
	if expireErr != nil {
		return expireErr
	}

	return nil
}

func (u *UserRestrictionService) RemoveFromRestrictionList(user aggregates.User) error {
	return u.localCache.Expire(u.UserIdToKey(user.Id()), time.Nanosecond)
}

func (u *UserRestrictionService) ProcessEvents() {
	defer func(messageBus application.MessageBusService) {
		_ = messageBus.Close()
	}(u.messageBus)

	topics := []string{fmt.Sprintf("%s", config.PROXY)}
	err := u.messageBus.Subscribe(topics)
	if err != nil {
		log.Fatalf("Failed to subscribe to topics: %s", err)
	}

	log.Printf("Subscribed to topics: %s", strings.Join(topics, ", "))

	for {
		event, consumeErr := u.messageBus.Consume()
		if consumeErr != nil {
			log.Printf("failed to consume from message bus: %s", consumeErr)
		}

		if event.EventType.Value() == "UserConsumedTrafficWithoutPlan" {
			var userConsumedTrafficWithoutPlan events.UserConsumedTrafficWithoutPlan
			deserializationErr := json.Unmarshal([]byte(event.Payload), &userConsumedTrafficWithoutPlan)
			if deserializationErr != nil {
				log.Printf("failed to deserialize user exceeded threshold event: %s", deserializationErr)
			}

			log.Printf("User %d restricted: UserConsumedTrafficWithoutPlan", userConsumedTrafficWithoutPlan.UserId)

			_ = u.setToLocalCacheWithTTL(userConsumedTrafficWithoutPlan.UserId)
		}

		if event.EventType.Value() == "UserExceededTrafficLimitEvent" {
			var userExceededTrafficLimitEvent events.UserExceededTrafficLimitEvent
			deserializationErr := json.Unmarshal([]byte(event.Payload), &userExceededTrafficLimitEvent)
			if deserializationErr != nil {
				log.Printf("failed to deserialize user exceeded threshold event: %s", deserializationErr)
			}

			log.Printf("User %d restricted: UserExceededTrafficLimitEvent", userExceededTrafficLimitEvent.UserId)

			_ = u.setToLocalCacheWithTTL(userExceededTrafficLimitEvent.UserId)
		}
	}
}

func (u *UserRestrictionService) UserIdToKey(userId int) string {
	return fmt.Sprintf("user:%d:restricted", userId)
}
