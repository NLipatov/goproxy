package application

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"goproxy/domain/valueobjects"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type UnauthorizedError struct {
}

func (e UnauthorizedError) Error() string {
	return "Unauthorized"
}

type ProxyUseCases struct {
	httpProxyListener HttpProxyListenerService
	proxyService      ProxyService
	authUseCases      AuthUseCases
	readerPool        *sync.Pool
}

func NewProxyUseCases(proxy ProxyService, httpProxyListener HttpProxyListenerService, authUseCases AuthUseCases) *ProxyUseCases {
	return &ProxyUseCases{
		proxyService:      proxy,
		httpProxyListener: httpProxyListener,
		authUseCases:      authUseCases,
		readerPool: &sync.Pool{
			New: func() interface{} {
				return bufio.NewReader(nil)
			},
		},
	}
}

func (p *ProxyUseCases) ServeOnPort(port int) {
	listener, listenerErr := p.httpProxyListener.Listen(port)
	if listenerErr != nil {
		log.Fatal(listenerErr)
	}

	for {
		clientConn, clientConnErr := listener.Accept()
		if clientConnErr != nil {
			log.Printf("failed to accept client connection: %v", clientConnErr)
			continue
		}

		go p.handleConnection(clientConn)
	}
}

func (p *ProxyUseCases) handleConnection(clientConn net.Conn) {
	defer func(clientConn net.Conn) {
		_ = clientConn.Close()
	}(clientConn)

	reader := p.readerPool.Get().(*bufio.Reader)
	reader.Reset(clientConn)
	defer p.readerPool.Put(reader)

	for {
		request, err := http.ReadRequest(reader)
		if err != nil {
			return
		}

		if err := p.HandleAuthorization(clientConn, request); err != nil {
			log.Printf("Authorization failed: %v", err)
			return
		}

		userID, err := strconv.Atoi(request.Header.Get("Proxy-Authorization"))
		if err != nil {
			log.Printf("Error parsing user id: %v", err)
			return
		}

		if request.Method == http.MethodConnect {
			p.proxyService.HandleHttps(clientConn, request, userID)
		} else {
			p.proxyService.HandleHttp(clientConn, request, userID)
		}
	}
}

func (p *ProxyUseCases) HandleAuthorization(clientConn net.Conn, request *http.Request) error {
	if request.Header.Get("Proxy-Authorization") == "" {
		_, _ = clientConn.Write([]byte("HTTP/1.1 407 Proxy Authentication Required\r\nProxy-Authenticate: Basic realm=\"Proxy\"\r\n\r\n"))
		return UnauthorizedError{}
	}

	credentialsHeader := strings.TrimPrefix(request.Header.Get("Proxy-Authorization"), "Basic ")
	credentials, extractCredentialsErr := p.extractCredentialsFromB64(credentialsHeader)
	if extractCredentialsErr != nil {
		log.Printf("Could not extract credentials: %v", extractCredentialsErr)
		return UnauthorizedError{}
	}

	authorized, userId, authorizationErr := p.authUseCases.AuthorizeBasic(credentials)
	if authorizationErr != nil {
		log.Printf("Could not authorize: %v", authorizationErr)
		return UnauthorizedError{}
	}
	if !authorized {
		log.Printf("Not authorized: %s", clientConn.RemoteAddr())
		return UnauthorizedError{}
	}

	request.Header.Set("Proxy-Authorization", fmt.Sprintf("%d", userId))

	return nil
}

func (p *ProxyUseCases) extractCredentialsFromB64(encoded string) (*valueobjects.BasicCredentials, error) {
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
