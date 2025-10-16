# JSON-RPC Client for Go

A lightweight and efficient JSON-RPC 2.0 client library for Go. This library simplifies the process of creating and executing JSON-RPC requests over HTTP, utilizing the high-performance JSON encoder/decoder from [sonic](https://github.com/bytedance/sonic) and robust error handling with [eris](https://github.com/rotisserie/eris).

## Features

- **JSON-RPC 2.0 Compliance:** Easily create and handle JSON-RPC 2.0 requests.
- **Flexible Options:** Support for custom HTTP clients, headers, and contexts.
- **Efficient Execution:** Send JSON-RPC requests with optimized connection pooling and buffer management.
- **Robust Error Handling:** Errors are wrapped with additional context using eris to aid in debugging.
- **High Performance:** Leverages [sonic](https://github.com/bytedance/sonic) for fast JSON encoding/decoding.
- **Production-Ready HTTP Client:** Pre-configured with optimized timeouts, connection pooling, and HTTP/2 support.

## Installation

Install the package using Go modules:

```bash
go get github.com/LiquidCats/jsonrpc
```

## Usage

### Basic Usage

The simplest way to use the library is to create a request and execute it:

```go
package main

import (
	"fmt"
	"log"

	"github.com/LiquidCats/jsonrpc"
)

func main() {
	// Define the parameters for the JSON-RPC call
	type Params struct {
		Value int `json:"value"`
	}

	// Create a JSON-RPC request expecting a string result
	req := jsonrpc.CreateRequest[string]("exampleMethod", Params{Value: 123})

	// Execute the request
	result, err := req.Execute("https://your.rpc")
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}

	fmt.Printf("Received result: %s\n", *result)
}
```

### Using Options

You can customize requests using options for headers, custom HTTP clients, or contexts:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/LiquidCats/jsonrpc"
)

func main() {
	// Create a custom HTTP client
	customClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create a request
	req := jsonrpc.CreateRequest[string]("exampleMethod", map[string]int{"value": 123})

	// Execute with options
	result, err := req.Execute(
		"https://your.rpc",
		jsonrpc.UseClient(customClient),
		jsonrpc.SetHeader("Authorization", "Bearer token"),
		jsonrpc.UseContext(context.WithValue(context.Background(), "key", "value")),
	)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}

	fmt.Printf("Received result: %s\n", *result)
}
```

## Examples

### Working with Complex Types

```go
package main

import (
	"fmt"
	"log"

	"github.com/LiquidCats/jsonrpc"
)

type BlockResult struct {
	Number     string   `json:"number"`
	Hash       string   `json:"hash"`
	ParentHash string   `json:"parentHash"`
	Timestamp  string   `json:"timestamp"`
}

func main() {
	// Request a block by number
	req := jsonrpc.CreateRequest[BlockResult]("eth_getBlockByNumber", []any{"latest", false})

	result, err := req.Execute("https://eth.llamarpc.com")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("Block #%s: %s\n", result.Number, result.Hash)
}
```

### Error Handling

The library automatically handles JSON-RPC errors and HTTP errors:

```go
package main

import (
	"fmt"
	"log"

	"github.com/LiquidCats/jsonrpc"
)

func main() {
	req := jsonrpc.CreateRequest[string]("invalidMethod", nil)

	result, err := req.Execute("https://your.rpc")
	if err != nil {
		// Error will include context about where it occurred
		fmt.Printf("Request failed: %v\n", err)
		return
	}

	fmt.Printf("Result: %s\n", *result)
}
```

## API Reference

### `CreateRequest`

```go
func CreateRequest[Result any](method string, params any) *RPCRequest[Result]
```

- **Description:** Creates a new JSON-RPC 2.0 request with the specified method and parameters.
- **Parameters:**
    - `method`: The JSON-RPC method to be invoked.
    - `params`: The parameters to be included in the JSON-RPC request (can be any type).
- **Returns:** A pointer to an `RPCRequest` with the specified result type.

### `Execute`

```go
func (rpc *RPCRequest[Resp]) Execute(url string, opts ...any) (*Resp, error)
```

- **Description:** Executes the JSON-RPC request against the specified URL and decodes the response.
- **Parameters:**
    - `url`: The target endpoint URL.
    - `opts`: Optional configuration functions for customizing the request.
- **Returns:** A pointer to the decoded result or an error if the execution fails.

### Option Functions

#### `UseClient`

```go
func UseClient(cli *http.Client) func(in *http.Client)
```

- **Description:** Sets a custom HTTP client for the request.

#### `SetHeader`

```go
func SetHeader(key, value string) func(in *http.Request)
```

- **Description:** Adds or sets a header on the HTTP request.

#### `UseContext`

```go
func UseContext(ctx context.Context) func(in *http.Request)
```

- **Description:** Sets a context for the HTTP request.

## Performance Features

The library includes several performance optimizations:

- **Connection Pooling:** Pre-configured with up to 4,096 idle connections and 1,024 per host
- **Buffer Pooling:** Reuses buffers to reduce memory allocations
- **HTTP/2 Support:** Enabled by default for multiplexing
- **Compression:** Enabled to reduce bandwidth for large JSON responses
- **Optimized Timeouts:** Carefully tuned connection and TLS handshake timeouts
- **Large Buffer Sizes:** 64KB read/write buffers for efficient handling of large responses

## Contributing

Contributions are welcome! If you have suggestions, bug fixes, or improvements, please submit an issue or create a pull request.

## License

This project is licensed under the GNU Affero General Public License v3.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgements

- **[sonic](https://github.com/bytedance/sonic):** For providing high-performance JSON encoding and decoding.
- **[eris](https://github.com/rotisserie/eris):** For enhanced error handling with stack traces and context wrapping.
