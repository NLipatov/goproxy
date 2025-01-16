package api_ws

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/websocket"
	"goproxy/application"
	"goproxy/infrastructure/api/api-http/google_auth"
)

type WSHandler struct {
	userUseCases     application.UserUseCases
	planInfoUseCases application.UserPlanInfoUseCases
	upgrader         websocket.Upgrader
	pingPeriod       time.Duration
	writeWait        time.Duration
	pongWait         time.Duration
	maxMessageSize   int64
}

func NewWSHandler(userUseCases application.UserUseCases, planInfoUseCases application.UserPlanInfoUseCases) *WSHandler {
	return &WSHandler{
		userUseCases:     userUseCases,
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

	userID, email, err := w.authenticateUser(r)
	if err != nil {
		log.Printf("authentication error: %v", err)
		_ = conn.WriteJSON(map[string]string{"error": err.Error()})
		return
	}

	log.Printf("Usser authenticated: ID=%d, Email=%s", userID, email)

	done := make(chan struct{})

	go w.readPump(conn, done)

	sendUserInfo := func() bool {
		plan, err := w.planInfoUseCases.FetchUserPlan(userID)
		if err != nil {
			log.Printf("could not load user plan: %v", err)
			_ = conn.WriteJSON(map[string]string{"error": "could not load plan info"})
			return false
		}

		traffic, err := w.planInfoUseCases.FetchTrafficUsage(userID)
		if err != nil {
			log.Printf("could not load traffic info: %v", err)
			_ = conn.WriteJSON(map[string]string{"error": "could not load traffic info"})
			return false
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

		log.Printf("sending data to user ID=%d", userID)
		if err := conn.WriteJSON(response); err != nil {
			log.Printf("could not send data to user: %v", err)
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

	updateTicker := time.NewTicker(3 * time.Second)
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

func (w *WSHandler) authenticateUser(r *http.Request) (int, string, error) {
	idToken, err := google_auth.GetIdTokenFromCookie(r)
	if err != nil {
		return 0, "", err
	}

	verifiedToken, err := google_auth.VerifyIDToken(idToken)
	if err != nil {
		return 0, "", err
	}

	claims, ok := verifiedToken.Claims.(jwt.MapClaims)
	if !ok {
		return 0, "", err
	}

	email, ok := claims["email"].(string)
	if !ok || email == "" {
		return 0, "", err
	}

	userIdCookie, err := r.Cookie("user_id")
	if err != nil {
		return 0, "", err
	}

	userId, err := strconv.Atoi(userIdCookie.Value)
	if err != nil {
		return 0, "", err
	}

	return userId, email, nil
}

func (w *WSHandler) readPump(conn *websocket.Conn, done chan struct{}) {
	defer close(done)
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("unexpected closing WebSocket error: %v", err)
			} else {
				log.Printf("WebSocket read err: %v", err)
			}
			return
		}
	}
}
