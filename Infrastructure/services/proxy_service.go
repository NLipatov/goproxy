package services

import (
	"bufio"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
)

type Proxy struct {
}

func NewProxy() *Proxy {
	return &Proxy{}
}

func (p *Proxy) Proxy(clientConn net.Conn, r *http.Request) {
	defer func(clientConn net.Conn) {
		err := clientConn.Close()
		if err != nil {
			log.Printf("Error closing client connection: %v", err)
		}
	}(clientConn)

	if r.Method == "CONNECT" {
		handleHttps(clientConn, r)
	} else {
		handleHttp(clientConn, r)
	}
}

func handleHttps(clientConn net.Conn, r *http.Request) {
	if !strings.Contains(r.URL.Host, ":") {
		r.URL.Host += ":443"
	}

	if r.URL.Host == "" {
		log.Println("Target host is empty")
		_, _ = clientConn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
		return
	}

	serverConn, err := net.Dial("tcp", r.URL.Host)
	if err != nil {
		log.Println("Could not connect:", err)
		_, _ = clientConn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		return
	}
	defer func(serverConn net.Conn) {
		_ = serverConn.Close()
	}(serverConn)

	_, _ = clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	var once sync.Once
	closeConn := func() {
		_ = clientConn.Close()
		_ = serverConn.Close()
	}

	go func() {
		_, err = io.Copy(serverConn, clientConn)
		if err != nil {
			log.Printf("Critical error copying client to server: %v", err)
		}
		once.Do(closeConn)
	}()

	_, err = io.Copy(clientConn, serverConn)
	if err != nil {
		log.Printf("Critical error copying server to client: %v", err)
	}
	once.Do(closeConn)
}

func handleHttp(clientConn net.Conn, r *http.Request) {
	if !strings.Contains(r.URL.Host, ":") {
		r.URL.Host += ":80"
	}

	serverConn, err := net.Dial("tcp", r.URL.Host)
	if err != nil {
		log.Println("Could not connect:", err)
		_, _ = clientConn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		return
	}
	defer func(serverConn net.Conn) {
		_ = serverConn.Close()
	}(serverConn)

	writeReqErr := r.Write(serverConn)
	if writeReqErr != nil {
		log.Printf("Could not write request to target: %v", writeReqErr)
		_, _ = clientConn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		return
	}

	resp, err := http.ReadResponse(bufio.NewReader(serverConn), r)
	if err != nil {
		log.Println("Could not read response:", err)
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	writeRespErr := resp.Write(clientConn)
	if writeRespErr != nil {
		log.Printf("Could not write response to client: %v", writeRespErr)
	}
}
