// Package response handles writing and sending
// a http response over raw tcp adhering to HTTP/1.1
package response

import (
	"fmt"
	"io"
	"strings"
)

var StatusRegister = map[int]string{
	200: "OK",
	201: "Created",
	400: "Bad Request",
	404: "Not Found",
	500: "Internal Server Error",
}

type Response struct {
	StatusCode int
	StatusText string
	Headers    map[string]string
	Body       []byte
}

func New() *Response {
	return &Response{
		StatusCode: 200,
		StatusText: StatusRegister[200],
		Headers:    make(map[string]string),
	}
}

func (res *Response) SetHeader(key, value string) {
	res.Headers[key] = value
}

func (res *Response) SetBody(body []byte, contentType string) {
	res.Body = body
	if contentType != "" {
		res.SetHeader("Content-Type", contentType)
	}
}

func (res *Response) SetStatus(code int) {
	text, ok := StatusRegister[code]
	if !ok {
		text = "Unknown"
	}

	res.StatusCode = code
	res.StatusText = text
}

func (res *Response) Send(w io.Writer) error {
	var builder strings.Builder

	fmt.Fprintf(&builder, "HTTP/1.1 %d %s\r\n", res.StatusCode, res.StatusText)

	res.SetHeader("Content-Length", fmt.Sprintf("%d", len(res.Body)))

	for key, val := range res.Headers {
		fmt.Fprintf(&builder, "%s: %s\r\n", key, val)
	}

	fmt.Fprint(&builder, "\r\n")

	_, err := w.Write([]byte(builder.String()))
	if err != nil {
		return fmt.Errorf("failed to write response headers: %w", err)
	}

	if len(res.Body) > 0 {
		_, err := w.Write(res.Body)
		if err != nil {
			return fmt.Errorf("failed to write response body: %w", err)
		}
	}

	return nil
}
