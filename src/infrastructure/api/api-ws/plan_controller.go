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

func NewPlanController(userUseCases application.UserUseCases, userPlanInfoUseCases application.UserPlanInfoUseCases) *PlanController {
	return &PlanController{
		wsHandler:   NewWSHandler(userUseCases, userPlanInfoUseCases),
		corsManager: CORS.NewCORSManager(),
	}
}

func (pc PlanController) Listen(port int) {
	http.Handle("/ws", pc.corsManager.AddCORS(pc.wsHandler))

	log.Printf("WebSocket server is running on port %d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
