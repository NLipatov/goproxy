package infrastructure

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"goproxy/application"
	"goproxy/domain/valueobjects"
	"goproxy/infrastructure/services"
	"log"
	"net"
	"net/http"
	"strings"
)

type HttpListener struct {
	authUseCases     application.AuthUseCases
	httpProxyService application.HttpProxyService
}

func NewHttpListener(authUseCases application.AuthUseCases) *HttpListener {
	return &HttpListener{
		httpProxyService: services.NewProxy(),
		authUseCases:     authUseCases,
	}
}

func (l *HttpListener) ServePort(port string) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatal("Could not start server:", err)
	}
	defer func(listener net.Listener) {
		_ = listener.Close()
	}(listener)

	log.Printf("Proxy is serving port %s", port)

	for {
		clientConn, acceptErr := listener.Accept()
		if acceptErr != nil {
			log.Printf("Failed serving client: %s", acceptErr)
			continue
		}

		reader := bufio.NewReader(clientConn)
		request, readRequestErr := http.ReadRequest(reader)
		if readRequestErr != nil {
			log.Printf("Could not read request: %v", readRequestErr)
			continue
		}

		if request.Header.Get("Proxy-Authorization") == "" {
			_, _ = clientConn.Write([]byte("HTTP/1.1 407 Proxy Authentication Required\r\nProxy-Authenticate: Basic realm=\"Proxy\"\r\n\r\n"))
			continue
		}

		credentialsHeader := strings.TrimPrefix(request.Header.Get("Proxy-Authorization"), "Basic ")
		credentials, extractCredentialsErr := l.ExtractCredentialsFromB64(credentialsHeader)
		if extractCredentialsErr != nil {
			log.Printf("Could not extract credentials: %v", extractCredentialsErr)
			continue
		}

		authorized, userId, authorizationErr := l.authUseCases.Authorize(credentials)
		if authorizationErr != nil {
			log.Printf("Could not authorize: %v", authorizationErr)
			continue
		}
		if !authorized {
			log.Printf("Not authorized: %s", clientConn.RemoteAddr())
			continue
		}

		request.Header.Set("Proxy-Authorization", fmt.Sprintf("%d", userId))

		go l.httpProxyService.Proxy(clientConn, request)
	}
}

func (l *HttpListener) ExtractCredentialsFromB64(encoded string) (*valueobjects.BasicCredentials, error) {
	if encoded == "" {
		return nil, fmt.Errorf("empty Base64 string")
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %v", err)
	}

	decoded := string(decodedBytes)
	parts := strings.SplitN(decoded, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid format, missing ':'")
	}

	return &valueobjects.BasicCredentials{
		Username: parts[0],
		Password: parts[1],
	}, nil
}
