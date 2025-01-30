package google_auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"goproxy/application/commands"
	"goproxy/application/contracts"
	"goproxy/application/use_cases"
	"goproxy/domain"
	"goproxy/domain/aggregates"
	"goproxy/domain/events"
	"goproxy/domain/valueobjects"
	"goproxy/infrastructure/api/Cookie"
	"goproxy/infrastructure/config"
	"goproxy/infrastructure/services"
	"io"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

type authData struct {
	idToken string
}

type GoogleAuthService struct {
	userUseCases        use_cases.UserUseCases
	cryptoService       contracts.CryptoService
	cookieBuilder       Cookie.CookieBuilder
	cache               contracts.CacheWithTTL[authData]
	oauthConfigProvider config.GoogleOauthConfigProvider
	messageBus          contracts.MessageBusService
}

func NewGoogleAuthService(userUseCases use_cases.UserUseCases, cryptoService contracts.CryptoService, messageBus contracts.MessageBusService) *GoogleAuthService {
	cache, cacheErr := services.NewRedisCache[authData]()
	if cacheErr != nil {
		log.Fatalf("failed to create cache instance: %s", cacheErr)
	}

	return &GoogleAuthService{
		userUseCases:        userUseCases,
		cryptoService:       cryptoService,
		cookieBuilder:       Cookie.NewCookieBuilder(),
		cache:               cache,
		oauthConfigProvider: config.NewGoogleOauthConfig(),
		messageBus:          messageBus,
	}
}

func (g *GoogleAuthService) handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	stateToken, stateTokenErr := g.cryptoService.GenerateRandomString(32)
	if stateTokenErr != nil {
		log.Printf("Failed to generate random string: %s", stateTokenErr)
		http.Error(w, "server failed generate stateToken id", http.StatusInternalServerError)
		return
	}

	err := g.cache.Set(stateToken, authData{})
	if err != nil {
		log.Printf("Failed to save state in cache: %s", err)
		http.Error(w, "server failed to save state token", http.StatusInternalServerError)
		return
	}
	_ = g.cache.Expire(stateToken, time.Minute*5)

	url := g.oauthConfigProvider.Config.AuthCodeURL(stateToken, oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (g *GoogleAuthService) handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	if _, err := g.cache.Get(state); err != nil {
		log.Printf("Invalid or expired state: %s", err)
		http.Error(w, "Invalid or expired state", http.StatusUnauthorized)
		return
	}

	if err := g.cache.Expire(state, time.Nanosecond); err != nil {
		log.Printf("Failed to delete state from cache: %s", err)
	}

	code := r.URL.Query().Get("code")
	token, err := g.oauthConfigProvider.Config.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, g.cookieBuilder.BuildCookie("/", "state-token-id", "", time.Nanosecond))

	client := g.oauthConfigProvider.Config.Client(context.Background(), token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	userInfo := struct {
		Email         string `json:"email"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		VerifiedEmail bool   `json:"verified_email"`
	}{}

	userInfoErr := json.NewDecoder(resp.Body).Decode(&userInfo)
	if userInfoErr != nil {
		http.Error(w, fmt.Sprintf("Failed to parse user info: %s", err), http.StatusInternalServerError)
		return
	}

	_, userErr := g.userUseCases.GetByEmail(userInfo.Email)
	if userErr != nil {
		if strings.Contains(userErr.Error(), "user not found") {
			normalizedUsername := valueobjects.NormalizeUsername(userInfo.Name)
			_, createUserErr := g.createNewUser(normalizedUsername, userInfo.Email)
			if createUserErr != nil {
				log.Printf("Failed to create new user: %s", createUserErr.Error())
				http.Error(w, "failed to generate new user", http.StatusInternalServerError)
				return
			}
		} else {
			log.Printf("Failed to fetch user: %s", userErr)
			http.Error(w, "failed to fetch user", http.StatusInternalServerError)
			return
		}
	}

	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No ID token found", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, g.cookieBuilder.BuildCookie("/", "id_token", idToken, time.Hour))
	http.Redirect(w, r, "http://localhost:5173/dashboard", http.StatusTemporaryRedirect)
}

type googleCerts struct {
	Keys []struct {
		Kid string `json:"kid"`
		N   string `json:"n"`
		E   string `json:"e"`
	} `json:"keys"`
}

func fetchGoogleCerts() (*googleCerts, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v3/certs")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Google certs: %v", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, bodyErr := io.ReadAll(resp.Body)
	if bodyErr != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var certs googleCerts
	unmarshalErr := json.Unmarshal(body, &certs)
	if unmarshalErr != nil {
		return nil, fmt.Errorf("failed to parse certs: %v", err)
	}

	return &certs, nil
}

func getPublicKey(certs *googleCerts, kid string) (*rsa.PublicKey, error) {
	for _, key := range certs.Keys {
		if key.Kid == kid {
			n, err := base64.RawURLEncoding.DecodeString(key.N)
			if err != nil {
				return nil, fmt.Errorf("failed to decode key N: %v", err)
			}

			e, err := base64.RawURLEncoding.DecodeString(key.E)
			if err != nil {
				return nil, fmt.Errorf("failed to decode key E: %v", err)
			}

			pubKey := &rsa.PublicKey{
				N: new(big.Int).SetBytes(n),
				E: int(new(big.Int).SetBytes(e).Uint64()),
			}
			return pubKey, nil
		}
	}
	return nil, errors.New("no matching public key found")
}

func VerifyIDToken(idToken string) (*jwt.Token, error) {
	certs, err := fetchGoogleCerts()
	if err != nil {
		return nil, err
	}

	token, err := jwt.Parse(idToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		kid, ok := t.Header["kid"].(string)
		if !ok {
			return nil, errors.New("missing kid in token header")
		}

		return getPublicKey(certs, kid)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to verify ID token: %v", err)
	}

	return token, nil
}

func (g *GoogleAuthService) createNewUser(name, email string) (proxyPassword string, err error) {
	password, passwordErr := g.createPassword(32)
	if passwordErr != nil {
		return "", fmt.Errorf("failed to register user - failed to create password: %s", passwordErr)
	}

	postUserCommand := commands.PostUser{
		Username: name,
		Password: password,
		Email:    email,
	}
	_, createErr := g.userUseCases.Create(postUserCommand)
	if createErr != nil {
		return "", fmt.Errorf("failed to register user - failed to create user: %s", createErr)
	}

	return password.Value, nil
}

func (g *GoogleAuthService) CheckAuthStatus(w http.ResponseWriter, r *http.Request) {
	idToken, err := GetIdTokenFromCookie(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	_, err = VerifyIDToken(idToken)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	response := map[string]bool{"authenticated": true}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func GetIdTokenFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie("id_token")
	if err != nil || cookie.Value == "" {
		return "", errors.New("not authenticated")
	}
	return cookie.Value, nil
}

func (g *GoogleAuthService) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	idToken, err := GetIdTokenFromCookie(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	verifiedToken, err := VerifyIDToken(idToken)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	claims, ok := verifiedToken.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Failed to parse token claims", http.StatusInternalServerError)
		return
	}

	name, _ := claims["name"].(string)
	picture, _ := claims["picture"].(string)

	if name == "" || picture == "" {
		http.Error(w, "Invalid token data", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"name":    name,
		"picture": picture,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

type BasicCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (g *GoogleAuthService) ResetPassword(w http.ResponseWriter, r *http.Request) {
	idToken, err := GetIdTokenFromCookie(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	verifiedToken, err := VerifyIDToken(idToken)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	claims, ok := verifiedToken.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Failed to parse token claims", http.StatusInternalServerError)
		return
	}

	email := claims["email"].(string)
	if email == "" {
		log.Printf("failed to reset proxy password: email claim empty")
		http.Error(w, "Failed to reset password", http.StatusInternalServerError)
		return
	}

	user, userErr := g.userUseCases.GetByEmail(email)
	if userErr != nil {
		log.Printf("failed to reset proxy user - failed to fetch user: %s", userErr)
		http.Error(w, "Failed to reset password", http.StatusInternalServerError)
		return
	}

	password, passwordErr := g.createPassword(32)
	if passwordErr != nil {
		fmt.Printf("failed to reset password - failed to create password: %s", passwordErr)
		http.Error(w, "failed to reset password", http.StatusInternalServerError)
		return
	}

	hash, err := g.cryptoService.HashValue(password.Value)
	if err != nil {
		log.Printf("failed to reset password - failed to hash password: %s", err)
		http.Error(w, "failed to reset password", http.StatusInternalServerError)
		return
	}

	updatedUser, updatedUserErr := aggregates.NewUser(user.Id(), user.Username(), user.Email(), hash)
	if updatedUserErr != nil {
		log.Printf("failed to reset password - failed to update user: %s", updatedUserErr)
		http.Error(w, "failed to reset password", http.StatusInternalServerError)
		return
	}

	updateErr := g.userUseCases.Update(updatedUser)
	if updateErr != nil {
		log.Printf("failed to reset password - failed to update user: %s", updateErr)
		http.Error(w, "failed to reset password", http.StatusInternalServerError)
		return
	}

	g.ProduceUserChangePasswordEvent(user.Username())

	updatedBasicCredentials := BasicCredentials{
		Username: user.Username(),
		Password: password.Value,
	}
	serializedCredentials, serializedCredentialsErr := json.Marshal(updatedBasicCredentials)
	if serializedCredentialsErr != nil {
		log.Printf("failed to reset password - failed to serialize credentials: %s", serializedCredentialsErr)
		http.Error(w, "failed to reset password", http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(serializedCredentials)
}

func (g *GoogleAuthService) createPassword(length int) (valueobjects.Password, error) {
	randomString, randomStringErr := g.cryptoService.GenerateRandomString(length)
	if randomStringErr != nil {
		return valueobjects.Password{}, fmt.Errorf("failed to register user - failed to generate password: %s", randomStringErr)
	}

	password, passwordErr := valueobjects.NewPasswordFromString(randomString)
	if passwordErr != nil {
		return valueobjects.Password{}, fmt.Errorf("failed to register user - failed to create password: %s", passwordErr)
	}

	return password, nil
}

func (g *GoogleAuthService) ProduceUserChangePasswordEvent(username string) {
	userPasswordChangedEvent := events.UserPasswordChangedEvent{
		Username: username,
	}

	serializedEvent, serializationErr := json.Marshal(userPasswordChangedEvent)
	if serializationErr != nil {
		log.Printf("failed to produce user changed password event - failed to serialize event: %s", serializationErr)
	}

	outboxEvent, outboxEventErr := events.NewOutboxEvent(-1, string(serializedEvent), false, "UserPasswordChangedEvent")
	if outboxEventErr != nil {
		log.Printf("failed to produce user changed password event - failed to create outbox event: %s", outboxEventErr)
	}

	produceErr := g.messageBus.Produce(fmt.Sprintf("%s", domain.PROXY), outboxEvent)
	if produceErr != nil {
		log.Printf("failed to reset password - failed to produce event: %s", produceErr)
	}
}
