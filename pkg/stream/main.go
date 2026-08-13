// Package stream is responsible for handling
// all the operations pertaining to raw tcp data stream
package stream

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// ReadRawStream takes a connection as an parameter
// and returns a byte channel `<-chan []byte`.
func ReadRawStream(conn io.ReadCloser) <-chan string {
	streamChan := make(chan string, 1)

	go func() {
		// Closes both the connection and the channel
		// after their purpose is served
		defer conn.Close()
		defer close(streamChan)

		// Buffer which reads only 8 bytes at a time
		recvBuf := make([]byte, 8)
		var accumulator strings.Builder

		for {
			bytesRead, err := conn.Read(recvBuf)

			if bytesRead > 0 {
				currentChunk := string(recvBuf[:bytesRead])
				for _, char := range currentChunk {
					if char == '\n' {
						streamChan <- accumulator.String()
						accumulator.Reset()
					} else {
						accumulator.WriteRune(char)
					}
				}
			}

			if err != nil {
				if errors.Is(err, io.EOF) {
					fmt.Println("Client disconnected!")
					if accumulator.Len() > 0 {
						streamChan <- accumulator.String()
					}

					break
				}
				fmt.Printf("Read error: %v\n", err)
				break
			}
		}
	}()

	return streamChan
}
