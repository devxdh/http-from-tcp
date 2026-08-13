package main

import (
	"fmt"
	"log"
	"net"

	"github.com/devxdh/http-from-scratch/pkg/stream"
)

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

		fmt.Println("[SERVER] Client Connected!")
		streamedChan := stream.ReadRawStream(conn)

		for line := range streamedChan {
			fmt.Println("read: ", line)
		}

		fmt.Println("[SERVER] Client session finished. Waiting for next connection...")
	}
}
