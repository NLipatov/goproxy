package crypto_cloud_billing

import (
	"fmt"
	"goproxy/application"
	"goproxy/application/payments/crypto_cloud"
	"goproxy/infrastructure/api/CORS"
	"goproxy/infrastructure/api/api-http/crypto_cloud_billing/handlers"
	"goproxy/infrastructure/api/api-http/crypto_cloud_billing/services"
	"log"
	"net/http"
)

type Controller struct {
	port                 int
	corsManager          CORS.CORSManager
	handler              Handler
	createInvoiceHandler handlers.CreateInvoiceHandler
}

func NewController(cryptoCloudService crypto_cloud.PaymentProvider,
	planPriceRepository application.PlanPriceRepository, orderRepository application.OrderRepository) *Controller {
	billingService := services.NewBillingService(orderRepository, planPriceRepository, cryptoCloudService)
	return &Controller{
		corsManager:          CORS.NewCORSManager(),
		handler:              NewHandler(cryptoCloudService, planPriceRepository, orderRepository),
		createInvoiceHandler: handlers.NewCreateInvoiceHandler(billingService),
	}
}

func (c *Controller) Listen(port int) {
	c.port = port

	mux := http.NewServeMux()
	mux.HandleFunc("/invoices", c.createInvoiceHandler.Handle)
	mux.HandleFunc("/postback", c.handler.HandlePostBack)

	corsHandler := c.corsManager.AddCORS(mux)

	log.Println(fmt.Sprintf("Server is running on port %d", c.port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", c.port), corsHandler))
}
