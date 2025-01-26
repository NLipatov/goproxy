package crypto_cloud_billing

import (
	"fmt"
	"goproxy/application/payments/crypto_cloud"
	"goproxy/infrastructure/api/CORS"
	"log"
	"net/http"
)

type Controller struct {
	port        int
	corsManager CORS.CORSManager
	handler     Handler
}

func NewController(cryptoCloudService crypto_cloud.PaymentProvider) *Controller {
	return &Controller{
		corsManager: CORS.NewCORSManager(),
		handler:     NewHandler(cryptoCloudService),
	}
}

func (c *Controller) Listen(port int) {
	c.port = port

	mux := http.NewServeMux()
	mux.HandleFunc("/invoices", c.handler.IssueInvoice)
	mux.HandleFunc("/postback", c.handler.HandlePostBack)

	corsHandler := c.corsManager.AddCORS(mux)

	log.Println(fmt.Sprintf("Server is running on port %d", c.port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", c.port), corsHandler))
}
