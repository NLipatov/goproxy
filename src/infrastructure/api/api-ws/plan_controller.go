package api_ws

import (
	"fmt"
	"goproxy/application"
	"goproxy/infrastructure/api/CORS"
	"log"
	"net/http"
)

type PlanController struct {
	wsHandler   *WSHandler
	corsManager CORS.CORSManager
}

func NewPlanController(userPlanInfoUseCases application.UserPlanInfoUseCases, usersApiHost string) *PlanController {
	return &PlanController{
		wsHandler:   NewWSHandler(userPlanInfoUseCases, usersApiHost),
		corsManager: CORS.NewCORSManager(),
	}
}

func (pc PlanController) Listen(port int) {
	http.Handle("/plans", pc.corsManager.AddCORS(pc.wsHandler))

	log.Printf("WebSocket server is running on port %d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
