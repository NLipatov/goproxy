package services

import (
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"goproxy/domain/aggregates"
	"goproxy/domain/events"
	"testing"
	"time"
)

func TestUserRestrictionService(t *testing.T) {
	localCache := new(MockCacheWithTTL[bool])
	remoteCache := new(MockCacheWithTTL[bool])
	messageBus := new(MockMessageBusService)

	service := &UserRestrictionService{
		localCache:  localCache,
		remoteCache: remoteCache,
		messageBus:  messageBus,
	}

	user, validationErr := aggregates.NewUser(1, "test_user", make([]byte, 32), make([]byte, 32))
	assert.NoError(t, validationErr)

	t.Run("IsRestricted - Found in local cache", func(t *testing.T) {
		localCache.On("Get", "user:1:restricted").Return(true, nil)
		assert.True(t, service.IsRestricted(user))
		localCache.AssertCalled(t, "Get", "user:1:restricted")
	})

	t.Run("IsRestricted - Found in remote cache", func(t *testing.T) {
		localCache.On("Get", "user:1:restricted").Return(false, errors.New("not found"))
		remoteCache.On("Get", "user:1:restricted").Return(true, nil)
		localCache.On("Set", "user:1:restricted", true).Return(nil)
		localCache.On("Expire", "user:1:restricted", userRestrictionServiceTTL).Return(nil)

		assert.False(t, service.IsRestricted(user)) // Возвращает false сразу
		time.Sleep(50 * time.Millisecond)           // Ждём завершения асинхронной горутины

		remoteCache.AssertCalled(t, "Get", "user:1:restricted")
		localCache.AssertCalled(t, "Set", "user:1:restricted", true)
		localCache.AssertCalled(t, "Expire", "user:1:restricted", userRestrictionServiceTTL)
	})

	t.Run("AddToRestrictionList", func(t *testing.T) {
		localCache.On("Set", "user:1:restricted", true).Return(nil).Once()
		localCache.On("Expire", "user:1:restricted", userRestrictionServiceTTL).Return(nil).Once()

		remoteCache.On("Set", "user:1:restricted", true).Return(nil).Once()
		remoteCache.On("Expire", "user:1:restricted", userRestrictionServiceTTL).Return(nil).Once()

		err := service.AddToRestrictionList(user)
		assert.NoError(t, err)

		localCache.AssertExpectations(t)
		remoteCache.AssertExpectations(t)
	})

	t.Run("RemoveFromRestrictionList", func(t *testing.T) {
		localCache.On("Expire", "user:1:restricted", time.Nanosecond).Return(nil)
		remoteCache.On("Expire", "user:1:restricted", time.Nanosecond).Return(nil)

		err := service.RemoveFromRestrictionList(user)
		assert.NoError(t, err)

		localCache.AssertExpectations(t)
		remoteCache.AssertExpectations(t)
	})

	t.Run("ProcessEvents", func(t *testing.T) {
		userExceededTrafficLimitEvent := events.NewUserExceededTrafficLimitEvent(1)
		serializedEvent, serializationErr := json.Marshal(userExceededTrafficLimitEvent)
		assert.NoError(t, serializationErr)

		outboxEvent, outboxEventValidationErr := events.NewOutboxEvent(1, string(serializedEvent), false, "UserExceededTrafficLimitEvent")
		assert.NoError(t, outboxEventValidationErr)

		messageBus.On("Subscribe", mock.Anything).Return(nil)
		messageBus.On("Consume").Return(&outboxEvent, nil).Once()
		messageBus.On("Consume").Return((*events.OutboxEvent)(nil), errors.New("no more messages")).Once()
		messageBus.On("Consume").Return((*events.OutboxEvent)(nil), errors.New("no more messages")).Maybe() // Дополнительные вызовы

		messageBus.On("Close").Return(nil)

		localCache.On("Set", "user:1:restricted", true).Return(nil)
		localCache.On("Expire", "user:1:restricted", userRestrictionServiceTTL).Return(nil)
		remoteCache.On("Set", "user:1:restricted", true).Return(nil)
		remoteCache.On("Expire", "user:1:restricted", userRestrictionServiceTTL).Return(nil)

		go service.ProcessEvents()
		time.Sleep(50 * time.Millisecond)

		messageBus.AssertCalled(t, "Subscribe", mock.Anything)
		messageBus.AssertCalled(t, "Consume")
		localCache.AssertCalled(t, "Set", "user:1:restricted", true)
		remoteCache.AssertCalled(t, "Set", "user:1:restricted", true)
	})
}
