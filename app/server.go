package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

var (
	listen = flag.String("listen", ":6379", "address to listen to")
)

func main() {
	flag.Parse()

	l, err := net.Listen("tcp", *listen)
	if err != nil {
		fmt.Println("Error starting server:", err)
		os.Exit(1)
	}
	defer l.Close()

	fmt.Println("Server is listening on", *listen)

	connChan := make(chan net.Conn)
	closeChan := make(chan net.Conn)

	go acceptConnections(l, connChan)

	activeConnections := make(map[net.Conn]struct{})

	for {
		select {
		case conn := <-connChan:
			activeConnections[conn] = struct{}{}
			go handleConnection(conn, closeChan)
		case conn := <-closeChan:
			delete(activeConnections, conn)
			conn.Close()
		}
	}
}

func acceptConnections(l net.Listener, connChan chan net.Conn) {
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		connChan <- conn
	}
}

func handleConnection(conn net.Conn, closeChan chan net.Conn) {
	defer func() {
		closeChan <- conn
	}()

	reader := bufio.NewReader(conn)
	for {
		requestLine, err := reader.ReadString('\n')
		if err != nil {
			if err.Error() != "EOF" {
				fmt.Println("Error reading request line:", err)
			}
			return
		}

		requestLine = strings.TrimSpace(requestLine)
		if requestLine == "" {
			continue
		}

		fmt.Printf("Received command: %s\n", requestLine)

		var response string
		if requestLine == "PING" {
			response = "+PONG\r\n"
		} else {
			response = fmt.Sprintf("+%s\r\n", requestLine)
		}

		_, err = conn.Write([]byte(response))
		if err != nil {
			fmt.Println("Error writing response:", err)
			return
		}
	}
}
