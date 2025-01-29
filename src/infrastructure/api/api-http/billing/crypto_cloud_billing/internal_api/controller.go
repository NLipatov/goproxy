package internal_api

import (
	"fmt"
	"goproxy/application"
	"goproxy/application/payments/crypto_cloud"
	"goproxy/infrastructure/api/CORS"
	handlers2 "goproxy/infrastructure/api/api-http/billing/crypto_cloud_billing/internal_api/handlers"
	services2 "goproxy/infrastructure/api/api-http/billing/crypto_cloud_billing/internal_api/services"
	"log"
	"net/http"
)

type Controller struct {
	port                 int
	corsManager          CORS.CORSManager
	createInvoiceHandler handlers2.CreateInvoiceHandler
	postBackHandler      handlers2.PostbackHandler
}

func NewController(cryptoCloudService crypto_cloud.PaymentProvider,
	planPriceRepository application.PlanPriceRepository, orderRepository application.OrderRepository,
	messageBus application.MessageBusService) *Controller {
	cryptoCloudMessageBusService := services2.NewCryptoCloudMessageBusService(messageBus)
	billingService := services2.NewBillingService(orderRepository, planPriceRepository, cryptoCloudService, cryptoCloudMessageBusService)
	postbackService := services2.NewPostbackService(orderRepository, planPriceRepository, cryptoCloudService, cryptoCloudMessageBusService)
	return &Controller{
		corsManager:          CORS.NewCORSManager(),
		createInvoiceHandler: handlers2.NewCreateInvoiceHandler(billingService),
		postBackHandler:      handlers2.NewPostbackHandler(postbackService),
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
