package services

import (
	"encoding/json"
	"fmt"
	"goproxy/application"
	"goproxy/domain"
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

	kafkaConfig, kafkaConfigErr := config.NewKafkaConfig(domain.PROXY)
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

	go func() {
		// check if another goroutine is checking this user in remote cache
		if _, loaded := remoteCacheLocks.LoadOrStore(key, true); loaded {
			return
		}
		defer remoteCacheLocks.Delete(key)

		restricted, err := u.remoteCache.Get(key)
		if err != nil {
			if !strings.Contains(err.Error(), "not found") {
				log.Printf("Error accessing Redis for user %d: %v", user.Id(), err)
			}

			restricted = false
		}
		_ = u.setToCacheWithTTL(u.localCache, user.Id(), restricted)
	}()

	return false
}

func (u *UserRestrictionService) AddToRestrictionList(user aggregates.User) error {
	setLocalErr := u.setToCacheWithTTL(u.localCache, user.Id(), true)
	if setLocalErr != nil {
		return fmt.Errorf("failed to add to local cache: %v", setLocalErr)
	}
	setRemoteErr := u.setToCacheWithTTL(u.remoteCache, user.Id(), true)
	if setRemoteErr != nil {
		return fmt.Errorf("failed to add to remote cache: %v", setRemoteErr)
	}

	return nil
}

func (u *UserRestrictionService) setToCacheWithTTL(cache application.CacheWithTTL[bool], userId int, restricted bool) error {
	key := u.UserIdToKey(userId)
	setErr := cache.Set(key, restricted)
	if setErr != nil {
		return setErr
	}

	expireErr := cache.Expire(key, userRestrictionServiceTTL)
	if expireErr != nil {
		return expireErr
	}

	return nil
}

func (u *UserRestrictionService) RemoveFromRestrictionList(user aggregates.User) error {
	expireLocalErr := u.localCache.Expire(u.UserIdToKey(user.Id()), time.Nanosecond)
	if expireLocalErr != nil {
		return expireLocalErr
	}
	expireRemoteErr := u.remoteCache.Expire(u.UserIdToKey(user.Id()), time.Nanosecond)
	if expireRemoteErr != nil {
		return expireRemoteErr
	}

	return nil
}

func (u *UserRestrictionService) ProcessEvents() {
	defer func(messageBus application.MessageBusService) {
		_ = messageBus.Close()
	}(u.messageBus)

	topics := []string{fmt.Sprintf("%s", domain.PROXY)}
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

		if event == nil {
			log.Printf("received nil event from message bus")
			continue
		}

		if event.EventType.Value() == "UserConsumedTrafficWithoutPlan" {
			var userConsumedTrafficWithoutPlan events.UserConsumedTrafficWithoutPlan
			deserializationErr := json.Unmarshal([]byte(event.Payload), &userConsumedTrafficWithoutPlan)
			if deserializationErr != nil {
				log.Printf("failed to deserialize user exceeded threshold event: %s", deserializationErr)
			}

			log.Printf("User %d restricted: UserConsumedTrafficWithoutPlan", userConsumedTrafficWithoutPlan.UserId)

			_ = u.setToCacheWithTTL(u.remoteCache, userConsumedTrafficWithoutPlan.UserId, true)
			_ = u.setToCacheWithTTL(u.localCache, userConsumedTrafficWithoutPlan.UserId, true)
		}

		if event.EventType.Value() == "UserExceededTrafficLimitEvent" {
			var userExceededTrafficLimitEvent events.UserExceededTrafficLimitEvent
			deserializationErr := json.Unmarshal([]byte(event.Payload), &userExceededTrafficLimitEvent)
			if deserializationErr != nil {
				log.Printf("failed to deserialize user exceeded threshold event: %s", deserializationErr)
			}

			log.Printf("User %d restricted: UserExceededTrafficLimitEvent", userExceededTrafficLimitEvent.UserId)

			_ = u.setToCacheWithTTL(u.remoteCache, userExceededTrafficLimitEvent.UserId, true)
			_ = u.setToCacheWithTTL(u.localCache, userExceededTrafficLimitEvent.UserId, true)
		}
	}
}

func (u *UserRestrictionService) UserIdToKey(userId int) string {
	return fmt.Sprintf("user:%d:restricted", userId)
}
