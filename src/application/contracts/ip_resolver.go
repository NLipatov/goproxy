package contracts

import (
	"net"
)

type IPResolver interface {
	GetHostPublicIPs() ([]net.IP, error)
}
