package main

import (
	"fmt"
	"log"
	"net"

	"github.com/devxdh/http-from-scratch/pkg/request"
	"github.com/devxdh/http-from-scratch/pkg/stream"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()
	fmt.Printf("[SERVER] Client connected: %s\n", clientAddr)

	reader := stream.NewReader(conn)

	req, err := request.Parse(reader)
	if err != nil {
		fmt.Printf("[SERVER] Error parsing request from %s: %v\n", clientAddr, err)
		return
	}

	fmt.Printf("[SERVER] Successfully parsed request from %s:\n", clientAddr)
	fmt.Printf("  Method:  %s\n", req.Method)
	fmt.Printf("  Path:    %s\n", req.Path)
	fmt.Printf("  Version: %s\n", req.Version)
	fmt.Printf("  Headers: %+v\n", req.Headers)

	if len(req.Body) > 0 {
		fmt.Printf("  Body:    %s\n", string(req.Body))
	} else {
		fmt.Println("  Body:    <emtpy>")
	}
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
