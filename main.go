package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

func handleConnection(clientConn net.Conn) {
	defer func(clientConn net.Conn) {
		err := clientConn.Close()
		if err != nil {
			log.Printf("Error closing client connection: %v", err)
		}
	}(clientConn)

	reader := bufio.NewReader(clientConn)
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		log.Println("Could not read request:", err)
		return
	}

	parts := strings.Fields(requestLine)
	if len(parts) < 2 {
		log.Println("Invalid request")
		return
	}

	method, target := parts[0], parts[1]
	log.Printf("Method: %s, Target: %s\n", method, target)

	if method == "CONNECT" {
		handleHttps(clientConn, target)
	} else {
		handleHttp(clientConn, reader)
	}
}

func handleHttps(clientConn net.Conn, target string) {
	serverConn, err := net.Dial("tcp", target)
	if err != nil {
		log.Println("Could not connect:", err)
		_, _ = clientConn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		return
	}
	defer func(serverConn net.Conn) {
		_ = serverConn.Close()
	}(serverConn)

	_, _ = clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	go func() {
		_, err = io.Copy(serverConn, clientConn)
		if err != nil {
			log.Println("Error copying client connection to server:", err)
		}
	}()

	_, err = io.Copy(clientConn, serverConn)
	if err != nil {
		log.Println("Error copying server connection to client:", err)
	}
}

func handleHttp(clientConn net.Conn, reader *bufio.Reader) {
	req, err := http.ReadRequest(reader)
	if err != nil {
		log.Println("Could not read HTTP-request:", err)
		return
	}

	serverConn, err := net.Dial("tcp", req.Host)
	if err != nil {
		log.Println("Could not connect:", err)
		_, _ = clientConn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		return
	}
	defer func(serverConn net.Conn) {
		_ = serverConn.Close()
	}(serverConn)

	_ = req.Write(serverConn)

	resp, err := http.ReadResponse(bufio.NewReader(serverConn), req)
	if err != nil {
		log.Println("Could not read response:", err)
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	_ = resp.Write(clientConn)
}

func main() {
	port := ":8888"
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("Could not start server:", err)
	}
	defer func(listener net.Listener) {
		_ = listener.Close()
	}(listener)

	log.Printf("Proxy listening on poort %s", port)

	for {
		clientConn, acceptErr := listener.Accept()
		if acceptErr != nil {
			log.Printf("Failed serving client: %s", acceptErr)
			continue
		}
		go handleConnection(clientConn)
	}
}
