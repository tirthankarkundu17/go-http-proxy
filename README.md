# Go HTTP Proxy

A lightweight, robust HTTP proxy written in Go. This proxy allows you to seamlessly forward standard HTTP requests (GET, POST, PUT, DELETE, etc.) to target URLs, while automatically injecting standard default headers when they are not explicitly provided.

## Features

- **Transparent Request Proxying**: Forwards your HTTP method, body payload, and custom headers seamlessly.
- **Smart Target Resolution**: Target URLs can be passed via a query parameter (`?url=...`) or via the `X-Proxy-Url` header to avoid complex double URL-encoding.
- **Graceful Shutdown**: Listens for OS interrupt signals (`SIGINT`, `SIGTERM`) to cleanly finish inflight requests before safely shutting down.
- **Context-Aware Connections**: Cleans up resources immediately if the calling client drops the connection early.
- **Enterprise-Ready**: Implements structured JSON logging, strict server timeouts to prevent slowloris attacks, and adheres to the standard Go project layout.

## Project Structure

```text
go-http-proxy/
├── cmd/
│   └── proxy/
│       └── main.go         # Application entrypoint
├── internal/
│   ├── config/
│   │   └── config.go       # Configuration and default proxy headers
│   └── proxy/
│       └── handler.go      # Core HTTP proxying logic
├── Makefile                # Build/Run automation
├── go.mod                  # Go module definition
└── README.md
```

## Getting Started

### Prerequisites
- [Go 1.21+](https://golang.org/doc/install) installed on your machine.
- Optional: `make` for utilizing the included Makefile.

### Building & Running

**Using Make:**
```bash
# Build and run the proxy server
make run

# Run tests
make test

# Build binary only
make build
```

**Using Standard Go Tools:**
```bash
# Run directly
go run cmd/proxy/main.go

# Build executable
go build -o bin/proxy cmd/proxy/main.go
```

The server will start locally on port `8080` (by default).

## Usage Examples

Once the server is running on `http://localhost:8080`, you can start proxying requests.

### 1. Using the `url` query parameter (Standard GET)
```bash
curl -X GET "http://localhost:8080/?url=https://httpbin.org/get"
```

### 2. Forwarding a POST request with a JSON body
```bash
curl -X POST "http://localhost:8080/?url=https://httpbin.org/post" \
  -H "Content-Type: application/json" \
  -d '{"hello":"world"}'
```

### 3. Using the `X-Proxy-Url` header
*This method is highly recommended if your target URL has its own query parameters, as it ensures you do not need to URL-encode the ampersands (`&`) or question marks (`?`).*
```bash
curl -X GET "http://localhost:8080/" \
  -H "X-Proxy-Url: https://httpbin.org/get?search=golang&page=1"
```

## Default Headers

The proxy will automatically inject the following standard web headers on outgoing requests, **unless** the client explicitly provides them in the incoming request:

- `User-Agent`
- `Accept`
- `Accept-Language`
- `Connection`
- `Cache-Control`

You can review or modify these exact default values in `internal/config/config.go`.
