// Package tcpParser handles parsing of raw
// tcp byte stream to meaningful http format.
package tcpParser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/devxdh/http-from-scratch/pkg/stream"
)

type Request struct {
	Method  string
	Path    string
	Version string
	Headers map[string]string
	Body    []byte
}

func ParseRequest(reader *stream.Reader) (*Request, error) {
	req := &Request{
		Headers: make(map[string]string),
	}

	line, err := reader.ReadLine()
	if err != nil {
		return nil, err
	}

	fLineArr := strings.Split(line, " ")
	if len(fLineArr) < 3 {
		return nil, fmt.Errorf("[PARSER] Malformed request line: %s", line)
	}

	req.Method = fLineArr[0]
	req.Path = fLineArr[1]
	req.Version = fLineArr[2]

	for {
		line, err := reader.ReadLine()
		if err != nil {
			return nil, err
		}

		if line == "" {
			fmt.Println("[PARSER] End of HTTP request reached.")
			break
		}

		headerLine := strings.SplitN(line, ":", 2)
		if len(headerLine) < 2 {
			return nil, fmt.Errorf("[PARSER] Malformed header line: %s", line)
		}

		key := strings.ToLower(strings.TrimSpace(headerLine[0]))
		val := strings.TrimSpace(headerLine[1])
		req.Headers[key] = val
	}

	if val, ok := req.Headers["content-length"]; ok {
		contentLength, err := strconv.Atoi(val)
		if err != nil {
			return nil, fmt.Errorf("[PARSER] Invalid content-length %s", val)
		}

		bodyByte, err := reader.ReadExact(contentLength)
		if err != nil {
			return nil, fmt.Errorf("[Parser] failed to read content: %v", err)
		}

		req.Body = bodyByte
	}

	return req, nil
}
