package crypto_cloud_billing

import (
	"fmt"
	"goproxy/application"
	"goproxy/application/payments/crypto_cloud"
	"goproxy/infrastructure/api/CORS"
	"goproxy/infrastructure/api/api-http/billing/crypto_cloud_billing/handlers"
	"goproxy/infrastructure/api/api-http/billing/crypto_cloud_billing/services"
	"log"
	"net/http"
)

type Controller struct {
	port                 int
	corsManager          CORS.CORSManager
	createInvoiceHandler handlers.CreateInvoiceHandler
	postBackHandler      handlers.PostbackHandler
}

func NewController(cryptoCloudService crypto_cloud.PaymentProvider,
	planPriceRepository application.PlanPriceRepository, orderRepository application.OrderRepository,
	messageBus application.MessageBusService) *Controller {
	cryptoCloudMessageBusService := services.NewCryptoCloudMessageBusService(messageBus)
	billingService := services.NewBillingService(orderRepository, planPriceRepository, cryptoCloudService, cryptoCloudMessageBusService)
	postbackService := services.NewPostbackService(orderRepository, planPriceRepository, cryptoCloudService, cryptoCloudMessageBusService)
	return &Controller{
		corsManager:          CORS.NewCORSManager(),
		createInvoiceHandler: handlers.NewCreateInvoiceHandler(billingService),
		postBackHandler:      handlers.NewPostbackHandler(postbackService),
	}
}

func (c *Controller) Listen(port int) {
	c.port = port

	mux := http.NewServeMux()
	mux.HandleFunc("/invoices", c.createInvoiceHandler.Handle)
	mux.HandleFunc("/postback", c.postBackHandler.Handle)

	corsHandler := c.corsManager.AddCORS(mux)

	log.Println(fmt.Sprintf("Server is running on port %d", c.port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", c.port), corsHandler))
}
