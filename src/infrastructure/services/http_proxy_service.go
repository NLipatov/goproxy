package services

import (
	"bufio"
	"fmt"
	"goproxy/application"
	"goproxy/infrastructure/config"
	"goproxy/infrastructure/infraerrs"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Proxy struct {
	rateLimiter   application.RateLimiterService
	dialerService application.DialerService
}

func NewProxy(dialerService *DialerService) *Proxy {
	rateLimiterConfig := config.LoadRateLimiterConfig()

	return &Proxy{
		rateLimiter:   NewRateLimiter(rateLimiterConfig),
		dialerService: dialerService,
	}
}

func (p *Proxy) Proxy(clientConn net.Conn, r *http.Request) {
	defer func(clientConn net.Conn) {
		_ = clientConn.Close()
	}(clientConn)

	userID, err := strconv.Atoi(r.Header.Get("Proxy-Authorization"))
	if err != nil {
		log.Printf("Error parsing user id: %v", err)
		return
	}

	if r.Method == http.MethodConnect {
		p.HandleHttps(clientConn, r, userID)
	} else {
		p.HandleHttp(clientConn, r, userID)
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
	if dialerErr != nil {
		log.Printf("Failed to get dialer: %v", dialerErr)
		_, _ = clientConn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
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

	rep, repErr := p.newTrafficReporter(userId)
	if repErr != nil {
		log.Fatalf("failed to build traffic reporter: %s", repErr)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// client → server
	go func() {
		defer wg.Done()
		_ = p.copyTrafficAndReport(userId, host, serverConn, clientConn, rep, "in")
	}()
	// server → client
	go func() {
		defer wg.Done()
		_ = p.copyTrafficAndReport(userId, host, clientConn, serverConn, rep, "out")
	}()

	wg.Wait()
	rep.SendFinal()
}

func (p *Proxy) copyTrafficAndReport(userId int, host string, dst io.Writer, src io.Reader, tr *TrafficReporter, direction string) error {
	if direction == "in" {
		defer p.rateLimiter.Done(userId, host)
	}

	buf := make([]byte, 32*1024)
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
				tr.AddInBytes(int64(written))
			} else {
				tr.AddOutBytes(int64(written))
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

	rep, repErr := p.newTrafficReporter(userId)
	if repErr != nil {
		log.Fatalf("failed to build traffic reporter: %s", repErr)
	}

	pr, pw := io.Pipe()
	go func() {
		_ = r.Write(pw)
		_ = pw.Close()
	}()

	// Request: PipeReader→serverConn (inBytes)
	copyRequestErr := p.copyTrafficAndReport(userId, host, serverConn, pr, rep, "in")
	if copyRequestErr != nil {
		return fmt.Errorf("copy request error: %w", copyRequestErr)
	}

	// Response: serverConn→clientConn (outBytes)
	copyResponseError := p.copyTrafficAndReport(userId, host, clientConn, serverConn, rep, "out")
	if copyResponseError != nil {
		return fmt.Errorf("copy response error: %w", copyResponseError)
	}

	rep.SendFinal()
	return nil
}

func (p *Proxy) newTrafficReporter(userId int) (*TrafficReporter, error) {
	return NewTrafficReporter(
		userId,
		1*1024*1024,
		5*time.Second,
	)
}
