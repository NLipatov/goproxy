package contracts

import (
	"net"
	"net/http"
)

type ProxyService interface {
	HandleHttps(clientConn net.Conn, r *http.Request, userId int)
	HandleHttp(clientConn net.Conn, r *http.Request, userId int)
}
