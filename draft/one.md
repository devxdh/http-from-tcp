### 2. The Request Lifecycle (`handleConnection`)

Before we dive into the low-level parsing mechanics, let's look at how our server coordinates a single connection from start to finish.

The `handleConnection` function executes four concrete steps in sequence: it wraps the raw TCP socket in a stream reader, parses the incoming HTTP request, routes it based on the URL path, and writes back an HTTP response.

Don't worry if some of these function calls look unfamiliar right now; we are going to build every single one of them. Here is how the complete flow looks:

```go
func handleConnection(conn net.Conn) {
	// Close the TCP connection when this function exits
	defer conn.Close()

	// Fetches the client's IP and port (e.g., "127.0.0.1:54321")
	clientAddr := conn.RemoteAddr().String()
	
	// Wrap the raw socket in a buffered stream reader
	reader := stream.NewReader(conn)

	// Read raw bytes from the socket and parse them into an HTTP Request struct
	req, err := request.Parse(reader)
	if err != nil {
		fmt.Printf("[SERVER] Error parsing request from %s: %v\n", clientAddr, err)

		// Send back a standard 400 Bad Request if the incoming bytes are malformed
		res := response.New()
		res.SetStatus(400)
		res.SetBody([]byte("400 Bad Request"), "text/plain")
		_ = res.Send(conn)
		return
	}

	fmt.Printf("[SERVER] %s %s from %s\n", req.Method, req.Path, clientAddr)

	// Create an in-memory HTTP response
	res := response.New()

	// Route matching based on the requested URL path
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

	// Format our response into raw HTTP/1.1 bytes and write them back down the socket
	err = res.Send(conn)
	if err != nil {
		fmt.Printf("[SERVER] Error sending response to %s: %v\n", clientAddr, err)
		return
	}

	fmt.Printf("[SERVER] Response successfully sent to %s\n", clientAddr)
}

```

Notice the three custom packages we are referencing here:

1. **`stream`**: Handles reading raw, continuous bytes off the TCP connection without losing data.
2. **`request`**: Parses those raw bytes into structured HTTP data (methods, paths, and headers).
3. **`response`**: Formats our status codes and payload into valid HTTP/1.1 wire text and pushes it back over the socket.

Now that we see how the lifecycle fits together, let's build these components step by step; starting with the stream reader.

---

#### Reading the Raw Byte-Stream (`pkg/stream`)

When we read from a raw TCP socket, we cannot simply ask the network for "one HTTP line"; the kernel only gives us raw chunks of bytes. If we tried to read byte-by-byte directly from the network connection just to find newline characters, we would trigger syscall on every single byte read.

To solve this, we wrap our raw TCP connection inside Go's `bufio.Reader`. This pulls a large chunk of bytes into an in-memory buffer in user space all at once. From there, we can scan for lines and extract exact byte counts without spamming the kernel.

Here is our `stream.Reader` implementation:

```go
package stream

import (
	"bufio"
	"io"
	"strings"
)

type Reader struct {
	buffered *bufio.Reader
}

// NewReader wraps a raw TCP connection with a buffered reader
func NewReader(r io.Reader) *Reader {
	return &Reader{
		buffered: bufio.NewReader(r),
	}
}

// ReadLine reads until a newline character and strips trailing CRLF (\r\n)
func (r *Reader) ReadLine() (string, error) {
	lineBytes, err := r.buffered.ReadBytes('\n')
	trimmed := strings.TrimRight(string(lineBytes), "\r\n")
	return trimmed, err
}

// ReadExact reads an exact number of bytes into a newly allocated slice
func (r *Reader) ReadExact(count int) ([]byte, error) {
	buf := make([]byte, count)

	// io.ReadFull blocks until all 'count' bytes are read or an error occurs
	_, err := io.ReadFull(r.buffered, buf)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

```

Notice why we split this into two read methods:

* **`ReadLine()` for Request-Lines & Headers:** HTTP metadata is plain text separated by CRLF (`\r\n`). We read line-by-line until we hit the empty line that marks the end of the headers.
* **`ReadExact(count)` for the Request Body:** Once we parse the headers, the `Content-Length` header tells us the exact size of the payload in bytes. `io.ReadFull` guarantees we pull exactly that many bytes from the stream; nothing more, nothing less.