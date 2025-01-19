package services

import (
	"fmt"
	"net"
	"sync"
)

type DialerService struct {
	ipPool  []net.IP
	mu      sync.Mutex
	ipIndex int
}

func NewDialerService() *DialerService {
	return &DialerService{
		ipPool:  []net.IP{},
		ipIndex: 0,
	}
}

func (d *DialerService) SetPool(ips []net.IP) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.ipPool = ips
	d.ipIndex = 0
}

func (d *DialerService) GetDialer(network string, userId int) (*net.Dialer, error) {
	// ToDo: implement sticky sessions based on userId

	d.mu.Lock()
	defer d.mu.Unlock()

	if len(d.ipPool) == 0 {
		return &net.Dialer{}, nil
	}

	localIP := d.ipPool[d.ipIndex]
	d.ipIndex = (d.ipIndex + 1) % len(d.ipPool)

	var localAddr net.Addr
	switch network {
	case "tcp", "tcp4", "tcp6":
		localAddr = &net.TCPAddr{IP: localIP}
	case "udp", "udp4", "udp6":
		localAddr = &net.UDPAddr{IP: localIP}
	default:
		return nil, fmt.Errorf("unsupported network: %s", network)
	}

	return &net.Dialer{LocalAddr: localAddr}, nil
}
