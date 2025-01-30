package services

import (
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"goproxy/application/contracts"
	"time"
)

type HS256Jwt struct {
}

func NewHS256Jwt() contracts.Jwt {
	return HS256Jwt{}
}

func (j HS256Jwt) Generate(secret string, ttl time.Duration, claims map[string]string) (string, error) {
	tokenClaims := jwt.MapClaims{
		"exp": time.Now().Add(ttl).Unix(),
	}

	for k, v := range claims {
		tokenClaims[k] = v
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func (j HS256Jwt) Validate(secret string, jwtToken string) (bool, error) {
	if len(secret) > 2 && secret[:2] == "ey" {
		return false, fmt.Errorf(
			"expected secret key but received JWT token. Did you mix up token and secret",
		)
	}

	token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
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
