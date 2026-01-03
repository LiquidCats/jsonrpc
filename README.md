# JSON‑RPC Client for Go

A lightweight, type‑safe client that implements the JSON‑RPC 2.0 specification over HTTP.  
It uses the high‑performance **sonic** JSON library for encoding/decoding and
the robust **eris** package for error handling.

## Features

- Fully typed requests & responses via generics.
- Zero‑alloc JSON with *sonic*.
- Extensible options: request‑level (headers, context, content‑type) and client‑level.
- Production‑ready HTTP client with tuned timeouts, connection pooling and HTTP/2.
- Rich error handling: JSON‑RPC errors are wrapped in `jsonrpc.RPCError`.

## Installation

```bash
go get github.com/LiquidCats/jsonrpc/v2
```

> The module is published under the `v2` path. Use that import path in your
> projects.

## Usage

### 1. Create a request

```go
package main

import (
	"fmt"
	"log"

	"github.com/LiquidCats/jsonrpc/v2"
)

func main() {
	type Params struct{ Value int }

	// Create a request that expects a string result.
	req := jsonrpc.NewRequest[Params, string](
		"exampleMethod",
		Params{Value: 123},
	)

	// Prepare the request for a specific endpoint.
	pReq := req.Prepare("https://your.rpc")

	// Execute with the library’s default HTTP client.
	result, err := pReq.Execute(nil)
	if err != nil {
		log.Fatalf("request failed: %v", err)
	}

	fmt.Printf("Result: %s\n", *result)
}
```

### 2. Customising a request

```go
package main

import (
	"context"
	"log"

	"github.com/LiquidCats/jsonrpc/v2"
)

func main() {
	req := jsonrpc.NewRequest[map[string]int, string](
		"exampleMethod",
		map[string]int{"value": 123},
	)

	pReq := req.Prepare(
		"https://your.rpc",
		jsonrpc.WithHeader("Authorization", "Bearer token"),
		jsonrpc.WithContext(context.Background()),
	)

	result, err := pReq.Execute(nil)
	if err != nil {
		log.Fatalf("request failed: %v", err)
	}

	fmt.Printf("Result: %s\n", *result)
}
```

### 3. Using a custom HTTP client

```go
package main

import (
	"net/http"
	"time"

	"github.com/LiquidCats/jsonrpc/v2"
)

func main() {
	req := jsonrpc.NewRequest[struct{}, string]("ping", struct{}{})
	pReq := req.Prepare("https://your.rpc")

	// Custom client with a 10‑second timeout.
	custom := &http.Client{Timeout: 10 * time.Second}

	result, err := pReq.Execute(custom)
	if err != nil {
		panic(err)
	}

	println(*result)
}
```

## API Reference

### `NewRequest[Params any, Result any](method string, params Params) *rpcRequest[Params, Result]`

Creates a new JSON‑RPC 2.0 request.

| Parameter | Type   | Description          |
|-----------|--------|----------------------|
| `method`  | string | RPC method name.     |
| `params`  | Params | Method parameters.   |

### `(*rpcRequest[Params, Result]) Prepare(url string, opts ...PrepareOpt) *praparedRPCRequest[Result]`

Prepares the request for a specific URL, applying any provided options.

| Parameter | Type          | Description                          |
|-----------|---------------|--------------------------------------|
| `url`     | string        | Target endpoint.                     |
| `opts`    | ...PrepareOpt | Request‑level options (headers, etc).|

### `(*praparedRPCRequest[Result]) Execute(client *http.Client, opts ...ExecuteOpt) (*Result, error)`

Executes the prepared request.

| Parameter | Type          | Description                                        |
|-----------|---------------|----------------------------------------------------|
| `client`  | *http.Client  | HTTP client to use; if nil, the library’s default is used. |
| `opts`    | ...ExecuteOpt | Client‑level options (currently none).             |

### Request‑level option helpers

| Function | Signature | Description |
|----------|-----------|-------------|
| `WithContext(ctx context.Context)` | `func(context.Context) PrepareOpt` | Sets the request’s context. |
| `WithHeader(key, value string)` | `func(string, string) PrepareOpt` | Adds or overrides an HTTP header. |
| `WithContentType(contentType string)` | `func(string) PrepareOpt` | Sets the `Content‑Type` header. |

### Error handling

- JSON‑RPC errors returned by the server are wrapped in `jsonrpc.RPCError`, which implements the `error` interface.
- HTTP status codes outside 2xx are returned as wrapped errors with the status code.

## Performance Notes

- **Connection pooling**: up to 4096 idle connections, 1024 per host.
- **Buffers**: 64 KB read/write buffers for efficient I/O.
- **HTTP/2**: enabled by default; multiplexed streams per connection.
- **Compression**: gzip/deflate automatically handled (`DisableCompression: false`).
- **TLS session cache**: 4096 entries.

## Contributing

Feel free to open issues or pull requests. All contributions are welcome!

## License

This project is licensed under the GNU Affero General Public License v3.0 – see the [LICENSE](LICENSE) file for details.