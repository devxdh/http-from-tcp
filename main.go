package main

import (
	"fmt"
	"log"
	"net"

	"github.com/devxdh/http-from-scratch/pkg/request"
	"github.com/devxdh/http-from-scratch/pkg/response"
	"github.com/devxdh/http-from-scratch/pkg/stream"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()
	reader := stream.NewReader(conn)

	req, err := request.Parse(reader)
	if err != nil {
		fmt.Printf("[SERVER] Error parsing request from %s: %v\n", clientAddr, err)

		res := response.New()
		res.SetStatus(400)
		res.SetBody([]byte("400 Bad Request"), "text/plain")
		_ = res.Send(conn)
		return
	}

	fmt.Printf("[SERVER] %s %s from %s\n", req.Method, req.Path, clientAddr)

	res := response.New()

	switch req.Path {
	case "/":
		res.SetStatus(200)
		res.SetBody([]byte("Welcome to my scratch HTTP Server!"), "text/plain")

	case "/users":
		res.SetStatus(200)
		res.SetBody([]byte(`{"message": "user list endpoint"}`), "application/json")

	default:
		res.SetStatus(404)
		res.SetBody([]byte("404 Page Not Found"), "text/plain")
	}

	err = res.Send(conn)
	if err != nil {
		fmt.Printf("[SERVER] Error sending response to %s: %v\n", clientAddr, err)
		return
	}

	fmt.Printf("[SERVER] Response successfully sent to %s\n", clientAddr)
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
