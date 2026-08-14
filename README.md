# http-from-scratch

Ever wondered what actually happens when you hit `Enter` on a URL, run `curl`, or fire off a `fetch()` request? 

Frameworks make it feel like magic. But if we strip away the layers, there’s no magic at all—just you, me, an operating system, and a pipe spitting raw bytes.

I built this project to stop guessing and actually see the machinery. It is a working HTTP/1.1 server written in Go, built straight from raw TCP sockets without leaning on `net/http` to do the heavy lifting for requests or responses.

Here is the story of how data turns into meaning, step by step.

---

## The Journey of a Request

### 1. Opening the Door
We start by telling the OS to listen on port `8080`. When your computer knocks, my server accepts the connection. But here’s the rule: we can’t make everyone else wait in line while we talk. So the second you connect, I hand our conversation off to its own lightweight Go thread (`go handleConnection(conn)`). The front door stays wide open for the next visitor.

### 2. Drinking from the Firehose
If we ask the operating system for every single byte one by one, performance tanks immediately. Each ask is an expensive system call. Instead, we place a 4KB buffer under the stream. The OS fills our bucket once, and we read smoothly from memory.

### 3. Reading the Envelope
Before we get to the actual message, we have to read the envelope. HTTP headers are just plain text lines separated by a carriage return and newline (`\r\n`):
* **The opener:** `GET /users HTTP/1.1` tells us the intent and the path.
* **The metadata:** Key-value lines like `Host:` and `Content-Type:`.
* **The secret handshake:** A completely empty line (`\r\n`). That’s the protocol's way of whispering: *“Headers are done. What follows is the actual payload.”*

### 4. The Body Plot Twist
Here is where things usually break if you aren't careful. We can’t read a request body line-by-line. Why? Because a JSON payload or image upload might contain line breaks inside it—or none at all. 

So we look at the `Content-Length` header (say, 26 bytes), stop looking for newlines, and tell the socket: *“Give me exactly 26 raw bytes, nothing more, nothing less.”*

### 5. Replying Back
A polite server never leaves you hanging. Once we process your request, we assemble a raw text frame in memory using `strings.Builder`:

```http
HTTP/1.1 200 OK\r\n
Content-Length: 12\r\n
Content-Type: text/plain\r\n
\r\n
Hello World!
```

We flush those bytes directly back down the TCP pipe, and gracefully close the socket. Conversation complete.

---

## How It's Organized

```text
.
├── pkg/
│   ├── stream/     # Socket reader handling buffering, line-splits, and exact-byte reads
│   ├── parser/     # Converts the raw incoming byte stream into a structured Request
│   └── response/   # Builds and serializes valid HTTP/1.1 response frames
├── main.go         # Concurrent TCP listener & endpoint handler
└── README.md
```

---

## Run It Yourself

Fire up the server:

```bash
go run main.go
```

Open another terminal and talk to it:

```bash
# Say hello
curl -i http://localhost:8080/

# Send some data
curl -i -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name": "devxdh"}'
```

---

## The Takeaway

Building this taught me one big lesson: the web isn't abstract magic. It's a simple, strict conversation between two machines that agree on where to put line breaks (`\r\n`), how to count bytes (`Content-Length`), and when to hang up the phone. It's just an agreement on how the conversation between two computers should be carried out and in what format that is it.

---