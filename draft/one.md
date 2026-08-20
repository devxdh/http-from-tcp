HTTP is something I interact with every single day, whether I am fetching data on the frontend or building backend APIs. But for a long time, if you had asked me how it actually worked under the hood, I wouldn't have had a good answer. I always treated HTTP like this intimidating, hyper-complex black box that would be miserable to understand, let alone rebuild from scratch. Once I actually dug into the internals, though, I realized the core mechanics are surprisingly straightforward.

I’m using Go for this project, but the concepts are entirely language-agnostic. You can follow along in C, Rust, Python, or whatever language you like; as long as it can open a network socket (sorry, HTML and CSS won't cut it here).

To understand HTTP, we first have to talk about TCP/IP, because HTTP is just a passenger riding on top of a horse named **TCP/IP**.

The **Internet Protocol (IP)** has one basic job: move raw data from Machine A to Machine B across the world. But it doesn't dump a massive file across the wire all at once. If you download a 10 GB file, IP chops that data into tiny pieces called packets, usually around 1,500 bytes each. It does this because network hardware has physical transmission limits, routers need to share bandwidth fairly among thousands of users, and resending one dropped 1.5 KB packet over flaky Wi-Fi is painless compared to re-downloading an entire 10 GB stream.

Every device on the internet gets an IP address so these packets know where to go. When you want to visit `google.com`, your computer doesn't magically know Google's physical server address right away. It first asks a DNS server to translate `google.com` into an actual IP address, like `142.250.190.46`. Once your machine has that destination IP, it stamps its own IP address into the packet header so Google knows where to send the reply, and fires the packets into the wild.

The catch is that IP is completely "best-effort." It launches packets into the network and immediately stops caring. If a router gets overloaded and drops your packet, or if packets take different routes and arrive completely out of order, IP won't fix it.

That is why we need **TCP (Transmission Control Protocol)**.

TCP sits right on top of IP to turn that chaotic packet delivery into a reliable stream. It tracks every byte with sequence numbers so out-of-order data gets assembled correctly. If a packet goes missing, TCP notices and asks the sender to retransmit it. It also introduces the concept of ports; so when data finally arrives at a computer's IP address, the operating system knows whether to hand those bytes to your web server on port 80 or your database on port 5432.

So now TCP has solved our biggest headache. We have a solid, reliable pipe open between our client and our server. You throw bytes into one end, and they pop out the other end in the exact right order without getting lost.

Problem solved, right? Not even close.

Because while TCP is great at moving bytes from one place to another reliably, it is completely blind to what those bytes actually *mean*. It treats everything as one never-ending, continuous river of data (in systems programming, we call it a raw byte stream). If your browser sends a request to load a profile picture, and immediately sends another request for a stylesheet, TCP just mashes all those bytes together in a single stream.

Your server is now sitting there staring at a raw chunk of bytes, completely clueless:

* Where does the first request end and the second one begin?
* Is the client trying to download a file, submit a form, or delete a record?
* Is this data plain text, an image, or JSON?
* If something goes wrong on the server, how do we tell the client?

If we didn't have a standard rulebook, every single developer would invent their own chaotic format. You'd write your own custom protocol where maybe you put an exclamation mark at the end of a message, while someone else uses a random binary flag. Your backend wouldn't be able to talk to any standard browser because neither speaks the same language.

This is where **HTTP (Hypertext Transfer Protocol)** steps in.

HTTP is nothing more than an agreed-upon rulebook. If TCP is a reliable telephone call connecting two people, HTTP is the actual grammar and vocabulary they agree to speak so they understand each other.

At its core, HTTP turns that blind stream of TCP bytes into predictable, structured messages. In HTTP/1.1, it does this entirely using plain text:

First, it forces the client to state its intent right on the very first line: like `GET /index.html HTTP/1.1`. Now the server instantly knows the action (`GET`), the target (`/index.html`), and the protocol version.

Next, it uses standard key-value headers separated by clean line breaks; specifically `\r\n` (CRLF: Carriage Return + Line Feed). Why two characters instead of just `\n`? Because early internet protocols inherited typewriter conventions from telegraph and terminal days, and now we are stuck with it forever.

Then, it solves the boundary problem with an empty line (`\r\n\r\n`), which screams to the parser: *"Hey, the headers are done! Whatever comes next is the actual body payload."*

Finally, the server replies with a standardized response that includes a status code like `200 OK` if everything went well, or `404 Not Found` if you asked for something that doesn't exist.

That's all HTTP really is. It's not magic, and it's not an intimidating engine. It's just a structured text format running over a raw TCP socket. Once you realize it's just plain text over a byte stream, building one yourself becomes a whole lot less scary.

---

## Building an HTTP Server from Raw TCP

Now that the underlying mechanics are clear, we can build the server from scratch.

We will break the implementation into four distinct steps:

1. Initialize a raw TCP socket and bind it to a local port.
2. Accept incoming client connections in a non-blocking loop.
3. Read raw bytes from the socket and parse the HTTP request.
4. Construct an HTTP response and write raw bytes back over the wire.

### 1. Setting Up the Raw TCP Listener

To handle network traffic in Go, we use the standard `net` package. We create our entry function `main()`, bind our listener to port `8080`, and loop to accept incoming connections:

```go
package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	// Bind a raw TCP socket to port 8080
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("[SERVER] Failed to bind to port: ", err)
	}
	// Ensure the socket listener is cleanly closed when main() exits
	defer listener.Close()

	fmt.Println("[SERVER] Listening on port :8080...")

	for {
    // Blocks until a client connects (completes the TCP 3-way handshake)
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("[SERVER] Failed to accept connection: %v\n", err)
			continue
		}

		// Handle each connection concurrently in a separate goroutine
		go handleConnection(conn)
	}
}

```

#### What is happening under the hood?

* **`net.Listen("tcp", ":8080")`:** This invokes the OS socket API (`socket()`, `bind()`, and `listen()`). It claims port `8080` from the kernel so incoming packets targeting this port are routed to our application.
* **`defer listener.Close()`:** The `defer` keyword guarantees that socket cleanup logic executes whenever the enclosing function (`main`) returns.
* **`listener.Accept()`:** This is a blocking system call. The server process goes to sleep until a client connects. Once the TCP 3-way handshake completes, the OS returns a `conn` object representing the bidirectional byte stream.
* **`go handleConnection(conn)`:** If we handled requests synchronously on the main thread, the entire server would freeze for all other users until that one request finished. The `go` keyword spawns a lightweight thread (**goroutine**), handing off the socket to `handleConnection` and immediately freeing the loop to accept the next client.