package services

import (
	"bufio"
	"fmt"
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
}

func NewProxy() *Proxy {
	return &Proxy{}
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
		p.handleHttps(clientConn, r, userID)
	} else {
		p.handleHttpConnection(clientConn, r, userID)
	}
}

func (p *Proxy) handleHttps(clientConn net.Conn, r *http.Request, userId int) {
	host := r.URL.Host
	if !strings.Contains(host, ":") {
		host += ":443"
	}
	if host == "" {
		log.Println("Target host is empty")
		_, _ = clientConn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
		return
	}

	serverConn, err := net.Dial("tcp", host)
	if err != nil {
		log.Println("Could not connect:", err)
		_, _ = clientConn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		return
	}
	defer func(serverConn net.Conn) {
		_ = serverConn.Close()
	}(serverConn)

	_, _ = clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	rep := p.newTrafficReporter(userId)

	var wg sync.WaitGroup
	wg.Add(2)

	// client → server
	go func() {
		defer wg.Done()
		_ = copyTrafficAndReport(serverConn, clientConn, rep, "in")
	}()
	// server → client
	go func() {
		defer wg.Done()
		_ = copyTrafficAndReport(clientConn, serverConn, rep, "out")
	}()

	wg.Wait()
	rep.SendFinal()
}

func copyTrafficAndReport(dst io.Writer, src io.Reader, tr *TrafficReporter, direction string) error {
	buf := make([]byte, 32*1024)

	for {
		n, readErr := src.Read(buf)
		if n > 0 {
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

func (p *Proxy) handleHttpConnection(clientConn net.Conn, firstReq *http.Request, userId int) {
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

	serverConn, err := net.Dial("tcp", host)
	if err != nil {
		log.Println("Could not connect:", err)
		_, _ = clientConn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		return err
	}
	defer func(serverConn net.Conn) {
		_ = serverConn.Close()
	}(serverConn)

	rep := p.newTrafficReporter(userId)

	pr, pw := io.Pipe()
	go func() {
		_ = r.Write(pw)
		_ = pw.Close()
	}()

	// Request: PipeReader→serverConn (inBytes)
	copyRequestErr := copyTrafficAndReport(serverConn, pr, rep, "in")
	if copyRequestErr != nil {
		return fmt.Errorf("copy request error: %w", copyRequestErr)
	}

	// Response: serverConn→clientConn (outBytes)
	copyResponseError := copyTrafficAndReport(clientConn, serverConn, rep, "out")
	if copyResponseError != nil {
		return fmt.Errorf("copy response error: %w", copyResponseError)
	}

	rep.SendFinal()
	return nil
}

func (p *Proxy) newTrafficReporter(userId int) *TrafficReporter {
	return NewTrafficReporter(
		userId,
		1*1024*1024,
		5*time.Second,
	)
}
