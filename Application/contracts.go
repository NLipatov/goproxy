package Application

import (
	"0trace/Domain/ValueObjects"
	"net"
	"net/http"
)

type AuthService interface {
	Authorize(credentials ValueObjects.Credentials) (bool, error)
}

type HttpProxyService interface {
	Proxy(clientConn net.Conn, r *http.Request)
}
