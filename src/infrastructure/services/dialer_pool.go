package services

import (
	"context"
	"errors"
	"goproxy/application/aplication_errors"
	"goproxy/application/contracts"
	"log"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"
)

const defaultRotationRecordTTL = time.Minute

type DialerPool struct {
	mu        sync.RWMutex
	randGen   *rand.Rand
	ipPool    []net.IP
	dialers   map[string]*net.Dialer
	userCache contracts.CacheWithTTL[net.IP]

	// used to resolve new public IPs assigned to server
	ipResolver contracts.IPResolver
}

func NewDialerPool(ipResolver contracts.IPResolver) *DialerPool {
	return &DialerPool{
		ipPool:     []net.IP{},
		dialers:    make(map[string]*net.Dialer),
		userCache:  NewMapCacheWithTTL[net.IP](),
		randGen:    rand.New(rand.NewSource(time.Now().UnixNano())),
		ipResolver: ipResolver,
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
					log.Printf("Failed to retrieve public IPs: %v\n", err)
					continue
				}
				dp.SetPool(ips)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (dp *DialerPool) SetPool(ips []net.IP) {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	dp.ipPool = ips
	dp.dialers = make(map[string]*net.Dialer, len(ips))

	for _, ip := range ips {
		ipStr := ip.String()
		dp.dialers[ipStr] = &net.Dialer{
			LocalAddr: &net.TCPAddr{IP: ip},
		}
	}
}

func (dp *DialerPool) GetDialer(_ string, userId int) (*net.Dialer, error) {
	dp.mu.RLock()
	empty := len(dp.ipPool) == 0
	dp.mu.RUnlock()
	if empty {
		return &net.Dialer{
			LocalAddr: &net.TCPAddr{},
		}, nil
	}

	key := strconv.Itoa(userId)

	cachedIP, err := dp.userCache.Get(key)
	if err != nil {
		dp.mu.RLock()
		ip := dp.randomIPLocked()
		dp.mu.RUnlock()

		_ = dp.userCache.Set(key, ip)
		_ = dp.userCache.Expire(key, defaultRotationRecordTTL)
		cachedIP = ip
	}

	dp.mu.RLock()
	dialer, ok := dp.dialers[cachedIP.String()]
	dp.mu.RUnlock()
	if !ok {
		return nil, errors.New("failed to retrieve dialer for IP " + cachedIP.String())
	}
	return dialer, nil
}

func (dp *DialerPool) BindDialerToUser(userId int, ttl time.Duration) error {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	if len(dp.ipPool) == 0 {
		return aplication_errors.ErrIpPoolEmpty{}
	}
	ip := dp.randomIPLocked()
	key := strconv.Itoa(userId)
	_ = dp.userCache.Set(key, ip)
	_ = dp.userCache.Expire(key, ttl)
	return nil
}

func (dp *DialerPool) randomIPLocked() net.IP {
	return dp.ipPool[dp.randGen.Intn(len(dp.ipPool))]
}
