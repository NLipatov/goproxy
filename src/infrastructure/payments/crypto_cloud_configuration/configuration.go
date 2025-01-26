package crypto_cloud_configuration

import (
	"log"
	"os"
)

type Configuration struct {
	baseUrl          string
	createInvoiceUrl string
	apiKey           string
	shopId           string
}

func NewConfiguration() Configuration {
	ApiKey := os.Getenv("API_KEY")
	if ApiKey == "" {
		log.Fatalf("API_KEY environment variable not set")
	}

	ShopId := os.Getenv("SHOP_ID")
	if ShopId == "" {
		log.Fatalf("SHOP_ID environment variable not set")
	}

	return Configuration{
		baseUrl:          "https://api.cryptocloud.plus",
		createInvoiceUrl: "/v2/invoice/create",
		apiKey:           ApiKey,
		shopId:           ShopId,
	}
}

func (c Configuration) BaseUrl() string {
	return c.baseUrl
}

func (c Configuration) CreateInvoiceUrl() string {
	return c.createInvoiceUrl
}

func (c Configuration) ApiKey() string {
	return c.apiKey
}

func (c Configuration) ShopId() string {
	return c.shopId
}
