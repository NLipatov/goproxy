package crypto_cloud_handlers

import (
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"goproxy/application"
	"goproxy/application/payments/crypto_cloud/crypto_cloud_commands"
	"goproxy/domain"
	"goproxy/domain/events"
	"time"
)

func HandlePostBack(command crypto_cloud_commands.PostBackCommand, messageBus application.MessageBusService) error {
	isTokenValid, isTokenValidErr := verifyJwt(command.Token, command.Secret)
	if isTokenValidErr != nil {
		return fmt.Errorf("could not validate token: %s", isTokenValidErr)
	}

	if !isTokenValid {
		return fmt.Errorf("invalid token")
	}

	orderPaidEvent := events.NewOrderPaidEvent(command.OrderID)
	serializedEvent, serializedEventErr := json.Marshal(orderPaidEvent)
	if serializedEventErr != nil {
		return serializedEventErr
	}

	event, eventErr := events.NewOutboxEvent(-1, string(serializedEvent), false, "OrderPaidEvent")
	if eventErr != nil {
		return eventErr
	}

	produceErr := messageBus.Produce(fmt.Sprintf("%s", domain.BILLING), event)
	if produceErr != nil {
		return produceErr
	}

	return nil
}

func verifyJwt(tokenString, secret string) (bool, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return false, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return false, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return false, fmt.Errorf("invalid token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if exp, expPresent := claims["exp"].(float64); expPresent {
			expTime := time.Unix(int64(exp), 0)
			if time.Now().After(expTime) {
				return false, fmt.Errorf("token has expired at %v", expTime)
			}
		}
	}

	return true, nil
}
