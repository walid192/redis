package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"syscall"
)

var (
	listen = flag.String("listen", ":6379", "adress to listen to")
)

func main() {
	flag.Parse()

	err := runServer(*listen)
	if err != nil {
		fmt.Println("Error starting server: ", err.Error())
		os.Exit(1)
	}

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	_, err = l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}


	defer l.Close()

	fmt.Println("Server is listening on", *listen)

	clients:=make(map[int]net.Conn)
	fdToConn:=make(map[int]int)


	readFds:=&syscall.FdSet{}

	for {
		clearFds(readFds)

		fd := int(listener.(*net.TCPListener).File().Fd())
		syscall.FD_SET(fd, readFds)

		for clientFd := range clients {
			syscall.FD_SET(clientFd, readFds)
		}

		_, err := syscall.Select(fd+1, readFds, nil, nil, nil)
		if err != nil {
			fmt.Println("Error with select:", err)
			continue
		}

		if syscall.FD_ISSET(fd, readFds) {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println("Error accepting connection:", err)
				continue
			}

			clientFd := int(conn.(*net.TCPConn).File().Fd())
			clients[clientFd] = conn
			fdToConn[clientFd] = clientFd
			fmt.Println("New client connected:", clientFd)
		}

		for clientFd := range clients {
			if syscall.FD_ISSET(clientFd, readFds) {
				handleRequest(clients[clientFd])
			}
		}

	}
}


func clearFds(fds *syscall.FdSet) {
	for i := 0; i < len(fds.Bits); i++ {
		fds.Bits[i] = 0
	}
}


func runServer(listen string) error {
	l, err := net.Listen("tcp", listen)
	if err != nil {
		return err
	}
	defer l.Close()
	fmt.Printf("Listening on %s...\n", listen)
	return nil
}


func handleRequest(conn net.Conn) {
	defer conn.Close()

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
