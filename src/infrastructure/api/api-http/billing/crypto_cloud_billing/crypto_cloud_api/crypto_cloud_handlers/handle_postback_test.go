package crypto_cloud_handlers

import (
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"goproxy/application/payments/crypto_cloud/crypto_cloud_commands"
	"goproxy/domain/events"
	"math"
	"math/rand/v2"
	"testing"
	"time"
)

func TestHandlePostBack_Success(t *testing.T) {
	secret := fmt.Sprintf("SECRET_%d", rand.IntN(math.MaxInt32))
	orderID := fmt.Sprintf("ORDER_N_%d", rand.IntN(math.MaxInt32))
	token := generateValidJWT(secret, orderID, time.Now().Add(time.Hour))

	messageBus := &handlePostBackMockMessageBusService{}
	command := crypto_cloud_commands.PostBackCommand{
		OrderID: orderID,
		Token:   token,
		Secret:  secret,
	}

	err := HandlePostBack(command, messageBus)

	assert.NoError(t, err)
	assert.Len(t, messageBus.ProducedEvents, 1)
	assert.Equal(t, "OrderPaidEvent", messageBus.ProducedEvents[0].EventType.Value())

	var payload events.OrderPaidEvent
	err = json.Unmarshal([]byte(messageBus.ProducedEvents[0].Payload), &payload)
	assert.NoError(t, err)
	assert.Equal(t, orderID, payload.OrderId)
}

func TestHandlePostBack_InvalidToken(t *testing.T) {
	secret := "my_secret_key"
	command := crypto_cloud_commands.PostBackCommand{
		OrderID: "test_order_id",
		Token:   "invalid_token",
		Secret:  secret,
	}

	messageBus := &handlePostBackMockMessageBusService{}

	err := HandlePostBack(command, messageBus)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not validate token")
}

func TestHandlePostBack_ExpiredToken(t *testing.T) {
	secret := "my_secret_key"
	orderID := "test_order_id"
	token := generateExpiredJWT(secret, orderID)

	command := crypto_cloud_commands.PostBackCommand{
		OrderID: orderID,
		Token:   token,
		Secret:  secret,
	}

	messageBus := &handlePostBackMockMessageBusService{}

	err := HandlePostBack(command, messageBus)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Token is expired")
}

type handlePostBackMockMessageBusService struct {
	ProducedEvents []events.OutboxEvent
}

func (m *handlePostBackMockMessageBusService) Subscribe(topics []string) error {
	return nil
}

func (m *handlePostBackMockMessageBusService) Consume() (*events.OutboxEvent, error) {
	return nil, nil
}

func (m *handlePostBackMockMessageBusService) Produce(topic string, event events.OutboxEvent) error {
	m.ProducedEvents = append(m.ProducedEvents, event)
	return nil
}

func (m *handlePostBackMockMessageBusService) Close() error {
	return nil
}

func generateValidJWT(secret string, orderID string, expiration time.Time) string {
	claims := jwt.MapClaims{
		"order_id": orderID,
		"exp":      expiration.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func generateExpiredJWT(secret string, orderID string) string {
	expiredTime := time.Now().Add(-time.Hour)
	return generateValidJWT(secret, orderID, expiredTime)
}
