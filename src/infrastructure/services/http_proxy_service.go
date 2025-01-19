package services

import (
	"bufio"
	"errors"
	"fmt"
	"goproxy/application"
	"goproxy/application/aplication_errors"
	"goproxy/infrastructure/config"
	"goproxy/infrastructure/infraerrs"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type Proxy struct {
	rateLimiter     application.RateLimiterService
	dialerService   application.DialerPool
	trafficReporter *TrafficReporter
}

var bufPool = sync.Pool{
	New: func() any {
		b := make([]byte, 65535)
		return &b
	},
}

func NewProxy(dialerService application.DialerPool) *Proxy {
	rateLimiterConfig := config.LoadRateLimiterConfig()
	trafficReporter, trafficReporterErr := NewTrafficReporter()
	if trafficReporterErr != nil {
		log.Fatalf("failed to create traffic reporterErr: %s", trafficReporterErr)
	}

	return &Proxy{
		rateLimiter:     NewRateLimiter(rateLimiterConfig),
		dialerService:   dialerService,
		trafficReporter: trafficReporter,
	}
}

func (p *Proxy) Proxy(clientConn net.Conn, r *http.Request) {
	defer func(clientConn net.Conn) {
		_ = clientConn.Close()
	}(clientConn)

	userId, err := strconv.Atoi(r.Header.Get("Proxy-Authorization"))
	if err != nil {
		log.Printf("Error parsing user id: %v", err)
		return
	}

	if r.Method == http.MethodConnect {
		p.HandleHttps(clientConn, r, userId)
	} else {
		p.HandleHttp(clientConn, r, userId)
	}
}

func (p *Proxy) HandleHttps(clientConn net.Conn, r *http.Request, userId int) {
	host := r.URL.Host
	if !strings.Contains(host, ":") {
		host += ":443"
	}
	if host == "" {
		log.Println("Target host is empty")
		_, _ = clientConn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
		return
	}

	dialer, dialerErr := p.dialerService.GetDialer("tcp", userId)
	if dialerErr != nil && errors.Is(dialerErr, aplication_errors.ErrIpPoolEmpty{}) {
		dialer = &net.Dialer{}
	} else if dialerErr != nil {
		log.Printf("failed to get dialer: %v", dialerErr)
		_, _ = clientConn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		return
	}

	serverConn, err := dialer.Dial("tcp", host)
	if err != nil {
		log.Println("Could not connect:", err)
		_, _ = clientConn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		return
	}
	defer func(serverConn net.Conn) {
		_ = serverConn.Close()
	}(serverConn)

	_, _ = clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	var wg sync.WaitGroup
	wg.Add(2)

	// client → server
	go func() {
		defer wg.Done()
		_ = p.copyTrafficAndReport(userId, host, serverConn, clientConn, "in")
	}()
	// server → client
	go func() {
		defer wg.Done()
		_ = p.copyTrafficAndReport(userId, host, clientConn, serverConn, "out")
	}()

	wg.Wait()
	go p.trafficReporter.FlushBuckets()
}

func (p *Proxy) copyTrafficAndReport(userId int, host string, dst io.Writer, src io.Reader, direction string) error {
	if direction == "in" {
		defer p.rateLimiter.Done(userId, host)
	}

	bufPtr := bufPool.Get().(*[]byte)
	buf := *bufPtr
	defer bufPool.Put(bufPtr)

	var accumulatedBytes int64
	const threshold = 1_000_000

	for {
		n, readErr := src.Read(buf)
		if n > 0 {
			if direction == "in" {
				accumulatedBytes += int64(n)
				if accumulatedBytes >= threshold {
					if !p.rateLimiter.Allow(userId, host, accumulatedBytes) {
						return infraerrs.RateLimitExceededError{}
					}
					accumulatedBytes = 0
				}
			}

			written, writeErr := dst.Write(buf[:n])
			if writeErr != nil {
				return writeErr
			}
			// direction == "in" → client→server
			// direction == "out" → server→client
			if direction == "in" {
				p.trafficReporter.AddInBytes(userId, int64(written))
			} else {
				p.trafficReporter.AddOutBytes(userId, int64(written))
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				return nil
			}
			return readErr
		}
	}
}

func (p *Proxy) HandleHttp(clientConn net.Conn, firstReq *http.Request, userId int) {
	br := bufio.NewReader(clientConn)
	req := firstReq

	for {
		if req == nil {
			var err error
			req, err = http.ReadRequest(br)
			if err != nil {
				break
			}
		}

		if err := p.handleHttpOnce(clientConn, req, userId); err != nil {
			log.Printf("handleHttpOnce error: %v", err)
			break
		}

		// req.Close indicates whether to close the connection after replying to this request
		if req.Close || strings.EqualFold(req.Header.Get("Connection"), "close") {
			break
		}

		// clear req to handle next request
		req = nil
	}
}

func (p *Proxy) handleHttpOnce(clientConn net.Conn, r *http.Request, userId int) error {
	host := r.URL.Host
	if !strings.Contains(host, ":") {
		host += ":80"
	}

	dialer, dialerErr := p.dialerService.GetDialer("tcp", userId)
	if dialerErr != nil {
		log.Printf("Failed to get dialer: %v", dialerErr)
		_, _ = clientConn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
		return dialerErr
	}

	serverConn, err := dialer.Dial("tcp", host)
	if err != nil {
		log.Println("Could not connect:", err)
		_, _ = clientConn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		return err
	}
	defer func(serverConn net.Conn) {
		_ = serverConn.Close()
	}(serverConn)

	pr, pw := io.Pipe()
	go func() {
		_ = r.Write(pw)
		_ = pw.Close()
	}()

	// Request: PipeReader→serverConn (inBytes)
	copyRequestErr := p.copyTrafficAndReport(userId, host, serverConn, pr, "in")
	if copyRequestErr != nil {
		return fmt.Errorf("copy request error: %w", copyRequestErr)
	}

	// Response: serverConn→clientConn (outBytes)
	copyResponseError := p.copyTrafficAndReport(userId, host, clientConn, serverConn, "out")
	if copyResponseError != nil {
		return fmt.Errorf("copy response error: %w", copyResponseError)
	}

	go p.trafficReporter.FlushBuckets()
	return nil
}
