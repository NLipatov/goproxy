package application

import "net"

type DialerService interface {
	GetDialer(network string, userId int) (*net.Dialer, error)
}

type DialerPool interface {
	SetPool(ips []net.IP)
}
