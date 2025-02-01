package free_plan_billing

import (
	"fmt"
	"goproxy/infrastructure/api/CORS"
	"log"
	"net/http"
)

type Controller struct {
	corsManager          CORS.CORSManager
	createInvoiceHandler Handler
}

func NewController(service *Service) Controller {
	return Controller{
		corsManager:          CORS.NewCORSManager(),
		createInvoiceHandler: NewHandler(service),
	}
}

func (c *Controller) Listen(port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/invoices", c.createInvoiceHandler.Handle)

	corsHandler := c.corsManager.AddCORS(mux)

	log.Println(fmt.Sprintf("free plan controller is listening on port %d", port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), corsHandler))
}
