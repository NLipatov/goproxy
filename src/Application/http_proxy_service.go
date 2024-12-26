package Application

import (
	"net"
	"net/http"
)

type HttpProxyService interface {
	Proxy(clientConn net.Conn, r *http.Request)
}
