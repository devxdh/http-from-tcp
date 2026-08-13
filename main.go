package main

import (
	"fmt"
	"log"
	"net"

	"github.com/devxdh/http-from-scratch/pkg/stream"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()
	fmt.Printf("[SERVER] Client connected: %s\n", clientAddr)

	reader := stream.NewReader(conn)

	for {
		line, err := reader.ReadLine()
		if err != nil {
			fmt.Printf("[SERVER] Client disconnected (%s): %v\n", clientAddr, err)
			return
		}

		if line == "" {
			fmt.Println("[SERVER] End of HTTP headers reached.")
			break
		}
	}

	fmt.Printf("[SERVER] finished reading from %s\n", clientAddr)
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("[SERVER] Error", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("failed to accept connection: %v", err)
			continue
		}

		go handleConnection(conn)
	}
}
