package services

import (
	"context"
	"errors"
	"fmt"
	"goproxy/application"
	"goproxy/application/aplication_errors"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"
)

type DialerPool struct {
	ipPool    []net.IP
	userCache application.CacheWithTTL[net.IP]
	mu        sync.RWMutex
	randGen   *rand.Rand

	// used to resolve new public IPs assigned to server
	ipResolver application.IPResolver
}

func NewDialerPool() *DialerPool {
	return &DialerPool{
		ipPool:    []net.IP{},
		userCache: NewMapCacheWithTTL[net.IP](),
		randGen:   rand.New(rand.NewSource(rand.Int63())),
	}
}

func (dp *DialerPool) StartExploringNewPublicIps(ctx context.Context, interval time.Duration) {
	go func() {
		dp.ipResolver = NewIPResolver()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				ips, err := dp.ipResolver.GetHostPublicIPs()
				if err != nil {
					fmt.Printf("Failed to retrieve public IPs: %v\n", err)
					continue
				}

				log.Printf("Discovered %d public IP(s)", len(ips))
				log.Printf("Public IPs: %v", ips)
				dp.mu.Lock()
				dp.ipPool = ips
				dp.mu.Unlock()

				fmt.Printf("Updated public IP pool: %v\n", ips)

			case <-ctx.Done():
				fmt.Println("Stopping public IP exploration")
				return
			}
		}
	}()
}

func (dp *DialerPool) SetPool(ips []net.IP) {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	dp.ipPool = ips
	dp.userCache = NewMapCacheWithTTL[net.IP]()
}

func (dp *DialerPool) BindDialerToUser(userId int, ttl time.Duration) error {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	if len(dp.ipPool) == 0 {
		return errors.New("IP pool is empty")
	}

	ip := dp.randomIP()
	key := dp.toCacheKey(userId)

	_ = dp.userCache.Set(key, ip)
	_ = dp.userCache.Expire(key, ttl)

	return nil
}

func (dp *DialerPool) GetDialer(network string, userId int) (*net.Dialer, error) {
	dp.mu.RLock()
	defer dp.mu.RUnlock()

	if len(dp.ipPool) == 0 {
		return nil, aplication_errors.IpPoolEmptyErr{}
	}

	var localIP net.IP
	key := dp.toCacheKey(userId)

	cachedIP, err := dp.userCache.Get(key)
	if err == nil {
		localIP = cachedIP
	} else {
		localIP = dp.randomIP()
		_ = dp.userCache.Set(key, localIP)
		_ = dp.userCache.Expire(key, 10*time.Minute)
	}

	var localAddr net.Addr
	switch network {
	case "tcp", "tcp4", "tcp6":
		localAddr = &net.TCPAddr{IP: localIP}
	case "udp", "udp4", "udp6":
		localAddr = &net.UDPAddr{IP: localIP}
	default:
		return nil, errors.New("unsupported network protocol")
	}

	return &net.Dialer{LocalAddr: localAddr}, nil
}

func (dp *DialerPool) randomIP() net.IP {
	return dp.ipPool[dp.randGen.Intn(len(dp.ipPool))]
}

func (dp *DialerPool) toCacheKey(userId int) string {
	return fmt.Sprintf("%d", userId)
}
