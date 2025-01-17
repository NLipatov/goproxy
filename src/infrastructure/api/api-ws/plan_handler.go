package api_ws

import (
	"encoding/json"
	"fmt"
	"goproxy/domain/dataobjects"
	"goproxy/infrastructure/dto"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/websocket"
	"goproxy/application"
	"goproxy/infrastructure/api/api-http/google_auth"
)

const updateInterval = time.Second * 5

type WSHandler struct {
	userApiHost      string
	planInfoUseCases application.UserPlanInfoUseCases
	upgrader         websocket.Upgrader
	pingPeriod       time.Duration
	writeWait        time.Duration
	pongWait         time.Duration
	maxMessageSize   int64
}

func NewWSHandler(planInfoUseCases application.UserPlanInfoUseCases, usersApiHost string) *WSHandler {
	return &WSHandler{
		userApiHost:      usersApiHost,
		planInfoUseCases: planInfoUseCases,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		pingPeriod:     54 * time.Second,
		writeWait:      10 * time.Second,
		pongWait:       60 * time.Second,
		maxMessageSize: 512,
	}
}

func (w *WSHandler) ServeHTTP(wr http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(wr, r.Body, w.maxMessageSize)

	conn, err := w.upgrader.Upgrade(wr, r, nil)
	if err != nil {
		log.Printf("could not upgrade connection to WebSocket: %v", err)
		http.Error(wr, "could not establish WebSocket connection", http.StatusInternalServerError)
		return
	}
	defer func(conn *websocket.Conn) {
		_ = conn.Close()
	}(conn)

	conn.SetReadLimit(w.maxMessageSize)
	_ = conn.SetReadDeadline(time.Now().Add(w.pongWait))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(w.pongWait))
		return nil
	})

	userId, err := w.authenticateUser(r)
	if err != nil {
		log.Printf("authentication error: %v", err)
		_ = conn.WriteJSON(map[string]string{"error": err.Error()})
		return
	}

	done := make(chan struct{})

	go w.readPump(conn, done)

	sendUserInfo := func() bool {
		plan, planErr := w.planInfoUseCases.FetchUserPlan(userId)
		if planErr != nil {
			plan = dataobjects.UserPlan{
				Name:      "N/A",
				Bandwidth: 0,
			}
		}

		traffic, trafficErr := w.planInfoUseCases.FetchTrafficUsage(userId)
		if trafficErr != nil {
			traffic = 0
		}

		response := map[string]interface{}{
			"plan": map[string]interface{}{
				"name": plan.Name,
				"limits": map[string]interface{}{
					"bandwidth": map[string]interface{}{
						"used":  traffic,
						"total": plan.Bandwidth,
					},
					"connections": 25,
					"speed":       "unlimited",
				},
			},
		}

		if writeJsonErr := conn.WriteJSON(response); writeJsonErr != nil {
			return false
		}
		return true
	}

	if !sendUserInfo() {
		log.Printf("WebSocket handling stopped because of error.")
		return
	}

	pingTicker := time.NewTicker(w.pingPeriod)
	defer pingTicker.Stop()

	updateTicker := time.NewTicker(updateInterval)
	defer updateTicker.Stop()

	for {
		select {
		case <-pingTicker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(w.writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("could not send ping: %v", err)
				return
			}

		case <-updateTicker.C:
			if !sendUserInfo() {
				log.Printf("WebSocket handling stopped because of error.")
				return
			}

		case <-done:
			log.Printf("client closed WebSocket connection.")
			return

		case <-r.Context().Done():
			log.Printf("client closed WebSocket connection.")
			return
		}
	}
}

func (w *WSHandler) authenticateUser(r *http.Request) (int, error) {
	idToken, err := google_auth.GetIdTokenFromCookie(r)
	if err != nil {
		return 0, err
	}

	verifiedToken, err := google_auth.VerifyIDToken(idToken)
	if err != nil {
		return 0, err
	}

	claims, ok := verifiedToken.Claims.(jwt.MapClaims)
	if !ok {
		return 0, err
	}

	email, ok := claims["email"].(string)
	if !ok || email == "" {
		return 0, err
	}

	resp, err := http.Get(fmt.Sprintf("%s/users/get?email=%s", w.userApiHost, email))
	if err != nil {
		return 0, fmt.Errorf("failed to fetch user id: %v", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, bodyErr := io.ReadAll(resp.Body)
	if bodyErr != nil {
		return 0, fmt.Errorf("failed to read response body: %v", err)
	}

	var userResult dto.GetUserResult
	deserializationErr := json.Unmarshal(body, &userResult)
	if deserializationErr != nil {
		return 0, fmt.Errorf("failed to deserialize user result: %v", deserializationErr)
	}

	return userResult.Id, nil
}

func (w *WSHandler) readPump(conn *websocket.Conn, done chan struct{}) {
	defer close(done)
	for {
		_, _, _ = conn.ReadMessage()
	}
}
