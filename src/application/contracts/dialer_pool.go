package contracts

import (
	"net"
	"time"
)

type DialerPool interface {
	// GetDialer retrieves a dialer for the given network and userId. If userId is bound, the corresponding IP is used.
	GetDialer(network string, userId int) (*net.Dialer, error)

	// BindDialerToUser binds an IP to a user for the specified TTL without creating a dialer.
	BindDialerToUser(userId int, ttl time.Duration) error

	// SetPool sets ip addresses available to be used with dialer
	SetPool(ips []net.IP)
}
