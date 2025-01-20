package application

import (
	"net"
)

type IPResolver interface {
	GetHostPublicIPs() ([]net.IP, error)
}
