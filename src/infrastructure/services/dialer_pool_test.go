package services

import (
	"net"
	"testing"
)

func TestDialerPool_GetDialer_EmptyPool(t *testing.T) {
	resolver := NewIPResolver()
	pool := NewDialerPool(resolver)
	dialer, dialerErr := pool.GetDialer("tcp", 1)
	if dialerErr != nil {
		t.Error(dialerErr)
	}

	if dialer == nil {
		t.Error(dialerErr)
	}
}

func TestDialerPool_GetDialer_NotEmptyPool(t *testing.T) {
	resolver := NewIPResolver()
	pool := NewDialerPool(resolver)

	poolIps := make([]net.IP, 3)
	poolIps[0] = net.ParseIP("127.0.0.1")
	poolIps[1] = net.ParseIP("127.0.0.2")
	poolIps[2] = net.ParseIP("127.0.0.3")
	pool.SetPool(poolIps)

	for i := 0; i < len(poolIps); i++ {
		dialer, dialerErr := pool.GetDialer("tcp", 1)
		if dialerErr != nil {
			t.Error(dialerErr)
		}

		if dialer == nil {
			t.Error(dialerErr)
		}
	}

}
