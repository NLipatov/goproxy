package accounting

import (
	"fmt"
	"goproxy/application"
	"goproxy/domain/lavatopsubdomain/lavatopaggregates"
	"goproxy/domain/lavatopsubdomain/lavatopvalueobjects"
	"goproxy/infrastructure/api/CORS"
	"goproxy/infrastructure/api/api-http/accounting/lavatop"
	"log"
	"net/http"
)

type Controller struct {
	handler     *lavatop.Handler
	corsManager CORS.CORSManager
	port        int
}

func NewAccountingController(
	billingService application.BillingService[lavatopaggregates.Invoice,
		lavatopvalueobjects.Offer], planRepository application.PlanRepository,
	planOfferRepository application.PlanOfferRepository) *Controller {
	handler := lavatop.NewHandler(billingService, planRepository, planOfferRepository)
	return &Controller{
		handler:     handler,
		corsManager: CORS.NewCORSManager(),
	}
}

func (c *Controller) Listen(port int) {
	c.port = port

	mux := http.NewServeMux()
	mux.HandleFunc("/plans", c.handler.GetPlans)
	mux.HandleFunc("/offers", c.handler.GetOffers)
	mux.HandleFunc("/invoices", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			c.handler.GetInvoices(w, r)
		case http.MethodPost:
			c.handler.PostInvoices(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	corsHandler := c.corsManager.AddCORS(mux)

	log.Println(fmt.Sprintf("Server is running on port %d", c.port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", c.port), corsHandler))
}
