package services

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDialerService_UpdateIPPool(t *testing.T) {
	dialerService := NewDialerService()

	ips := []net.IP{
		net.ParseIP("192.0.2.1"),
		net.ParseIP("192.0.2.2"),
	}
	dialerService.SetPool(ips)

	dialerService.mu.Lock()
	assert.Equal(t, ips, dialerService.ipPool, "IP pool should match updated values")
	assert.Equal(t, 0, dialerService.ipIndex, "IP index should reset to 0")
	dialerService.mu.Unlock()
}

func TestDialerService_GetDialer_Rotation(t *testing.T) {
	dialerService := NewDialerService()

	ips := []net.IP{
		net.ParseIP("192.0.2.1"),
		net.ParseIP("192.0.2.2"),
		net.ParseIP("192.0.2.3"),
	}
	dialerService.SetPool(ips)

	expectedIPs := []string{"192.0.2.1", "192.0.2.2", "192.0.2.3", "192.0.2.1"}
	for i, expected := range expectedIPs {
		dialer, err := dialerService.GetDialer("tcp", 1)
		assert.NoError(t, err, "GetDialer should not return an error")
		assert.Equal(t, expected, dialer.LocalAddr.(*net.TCPAddr).IP.String(), "Dialer should use the correct IP in rotation")
		assert.Equal(t, (i+1)%len(ips), dialerService.ipIndex, "IP index should increment correctly")
	}
}

func TestDialerService_GetDialer_EmptyPool(t *testing.T) {
	dialerService := NewDialerService()

	dialer, err := dialerService.GetDialer("tcp", 1)
	assert.NoError(t, err, "GetDialer should not return an error with empty pool")
	assert.Nil(t, dialer.LocalAddr, "Dialer.LocalAddr should be nil when IP pool is empty")
}

func TestDialerService_GetDialer_UnsupportedNetwork(t *testing.T) {
	dialerService := NewDialerService()

	ips := []net.IP{
		net.ParseIP("192.0.2.1"),
	}
	dialerService.SetPool(ips)

	_, err := dialerService.GetDialer("unsupported", 1)
	assert.Error(t, err, "GetDialer should return an error for unsupported network")
	assert.Equal(t, "unsupported network: unsupported", err.Error(), "Error message should match")
}
