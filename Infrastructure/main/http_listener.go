package main

import (
	"0trace/Application"
	"0trace/Domain/ValueObjects"
	"0trace/Infrastructure/services"
	"bufio"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
)

type httpListener struct {
	authService      Application.AuthService
	httpProxyService Application.HttpProxyService
}

func newHttpListener() *httpListener {
	return &httpListener{
		authService:      services.NewAuthService(),
		httpProxyService: services.NewProxy(),
	}
}

func (l *httpListener) servePort(port string) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatal("Could not start server:", err)
	}
	defer func(listener net.Listener) {
		_ = listener.Close()
	}(listener)

	log.Printf("Serving port %s", port)

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

		isCredentialsValid, authorizationErr := l.authService.Authorize(credentials)
		if authorizationErr != nil {
			log.Printf("Could not authorize: %v", authorizationErr)
			continue
		}
		if !isCredentialsValid {
			log.Printf("Invalid credentials. Client: %s", clientConn.RemoteAddr())
			continue
		}

		go l.httpProxyService.Proxy(clientConn, request)
	}
}

func (l *httpListener) ExtractCredentialsFromB64(encoded string) (*ValueObjects.BasicCredentials, error) {
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

	return &ValueObjects.BasicCredentials{
		Username: parts[0],
		Password: parts[1],
	}, nil
}
