package application

import (
	"net"
)

type HttpProxyListenerService interface {
	Listen(port int) (net.Listener, error)
}
