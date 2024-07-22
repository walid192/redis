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
	for {
		req, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleRequest(req)

	}
}

func runServer(listen string) error {
	l, err := net.Listen("tcp", listen)
	if err != nil {
		return err
	}
	defer l.Close()
	fmt.Printf("Listening on %s...\n", listen)
	for {
		req, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleRequest(req)

	}
}

func handleRequest(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		requestLine, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading request line:", err.Error())
			return
		}

		requestLine = strings.TrimSpace(requestLine)
		if requestLine == "" {
			break
		}

		if requestLine == "PING" {
			response := "+PONG\r\n"
			_, err := conn.Write([]byte(response))
			if err != nil {
				fmt.Println("Error writing response:", err.Error())
				return
			}
		} else {
			response := fmt.Sprintf("+%s\r\n", requestLine)
			_, err := conn.Write([]byte(response))
			if err != nil {
				fmt.Println("Error writing response:", err.Error())
				return
			}
		}
	}
}
