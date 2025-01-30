package generic

import (
	"fmt"
	"goproxy/application/contracts"
	"goproxy/infrastructure/api/CORS"
	"goproxy/infrastructure/api/api-http/billing/generic/handlers"
	"log"
	"net/http"
)

type Controller struct {
	getPlanPricesHandler handlers.GetPlanPricesHandler
	corsManager          CORS.CORSManager
}

func NewBillingController(planPriceRepository contracts.PlanPriceRepository) *Controller {
	return &Controller{
		getPlanPricesHandler: handlers.NewGetPlanPricesHandler(planPriceRepository),
		corsManager:          CORS.NewCORSManager(),
	}
}

func (c *Controller) Listen(port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/plans/prices", c.getPlanPricesHandler.Handle)

	corsHandler := c.corsManager.AddCORS(mux)

	log.Println(fmt.Sprintf("Server is running on port %d", port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), corsHandler))
}
