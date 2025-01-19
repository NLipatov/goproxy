package services

import (
	"bytes"
	"net"
)

type HostIPResolver struct{}

func NewIPResolver() *HostIPResolver {
	return &HostIPResolver{}
}

func (r *HostIPResolver) GetHostPublicIPs() ([]net.IP, error) {
	var publicIPs []net.IP

	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, addrsErr := iface.Addrs()
		if addrsErr != nil {
			continue
		}

		for _, addr := range addrs {
			ip, _, parseErr := net.ParseCIDR(addr.String())
			if parseErr != nil {
				continue
			}

			if isPublicIPv4(ip) {
				publicIPs = append(publicIPs, ip)
			}
		}
	}

	return publicIPs, nil
}

func isPublicIPv4(ip net.IP) bool {
	if ip == nil || ip.To4() == nil {
		return false
	}

	privateRanges := []struct {
		start net.IP
		end   net.IP
	}{
		{net.IPv4(10, 0, 0, 0), net.IPv4(10, 255, 255, 255)},
		{net.IPv4(172, 16, 0, 0), net.IPv4(172, 31, 255, 255)},
		{net.IPv4(192, 168, 0, 0), net.IPv4(192, 168, 255, 255)},
	}

	for _, r := range privateRanges {
		if bytesCompare(ip, r.start) >= 0 && bytesCompare(ip, r.end) <= 0 {
			return false
		}
	}

	return !ip.IsLoopback() && !ip.IsUnspecified()
}

func bytesCompare(ip1, ip2 net.IP) int {
	return bytes.Compare(ip1.To4(), ip2.To4())
}
