package plans

import (
	"fmt"
	"goproxy/application"
	"goproxy/infrastructure/api/CORS"
	"goproxy/infrastructure/api/api-http/plans/plans_handlers"
	"log"
	"net/http"
)

type Controller struct {
	getPlansHandler plans_handlers.GetPlansHandler
	corsManager     CORS.CORSManager
}

func NewPlansController(planRepository application.PlanRepository) *Controller {
	return &Controller{
		corsManager:     CORS.NewCORSManager(),
		getPlansHandler: plans_handlers.NewHandler(planRepository),
	}
}

func (c *Controller) Listen(port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/plans", c.getPlansHandler.Handle)

	corsHandler := c.corsManager.AddCORS(mux)

	log.Println(fmt.Sprintf("Server is running on port %d", port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), corsHandler))
}
