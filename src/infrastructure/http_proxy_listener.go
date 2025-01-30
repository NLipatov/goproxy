package infrastructure

import (
	"fmt"
	"goproxy/application/contracts"
	"log"
	"net"
)

type HttpListener struct {
	httpProxyService contracts.HttpProxyService
}

func NewHttpListener(proxy contracts.HttpProxyService) *HttpListener {
	return &HttpListener{
		httpProxyService: proxy,
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
