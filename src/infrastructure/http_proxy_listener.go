package infrastructure

import (
	"fmt"
	"goproxy/application"
	"goproxy/infrastructure/services"
	"log"
	"net"
)

type HttpListener struct {
	httpProxyService application.HttpProxyService
}

func NewHttpListener() *HttpListener {
	return &HttpListener{
		httpProxyService: services.NewProxy(services.NewDialerService()),
	}
}

func (l *HttpListener) Listen(port int) (net.Listener, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("could not start server: %v", err)
	}
	log.Printf("Proxy is serving port %d", port)
	return listener, nil
}
