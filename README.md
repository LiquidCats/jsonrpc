# JSON-RPC Client for Go

A lightweight and efficient JSON-RPC 2.0 client library for Go. This library simplifies the process of preparing and executing JSON-RPC requests over HTTP, utilizing the high-performance JSON encoder/decoder from [sonic](https://github.com/bytedance/sonic) and robust error handling with [go-faster/errors](https://github.com/go-faster/errors).

## Features

- **JSON-RPC 2.0 Compliance:** Easily create and handle JSON-RPC 2.0 requests.
- **Simple Request Preparation:** Use the `Prepare` function to generate HTTP requests with JSON-RPC payloads.
- **Efficient Execution:** Send JSON-RPC requests using the standard `http` package and decode responses seamlessly.
- **Robust Error Handling:** Errors are wrapped with additional context to aid in debugging.
- **High Performance:** Leverages [sonic](https://github.com/bytedance/sonic) for fast JSON encoding/decoding.

## Installation

Install the package using Go modules:

```bash
go get github.com/LiquidCats/jsonrpc
```

*Replace `github.com/LiquidCats/jsonrpc` with the actual repository path.*

## Usage

### Preparing a Request

The `Prepare` function creates an HTTP POST request with a JSON-RPC payload. Below is an example demonstrating how to prepare a request:

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/LiquidCats/jsonrpc" // adjust the import path according to your project structure
)

func main() {
	// Define the parameters for the JSON-RPC call.
	type Params struct {
		Value int `json:"value"`
	}

	params := Params{Value: 123}
	url := "http://yours.rpc"
	method := "exampleMethod"

	// Prepare the JSON-RPC request.
	req, err := jsonrpc.Prepare(context.Background(), url, method, params)
	if err != nil {
		log.Fatalf("Failed to prepare request: %v", err)
	}

	fmt.Println("Request prepared successfully.")
	// The request can now be sent using the Execute function.
}
```

### Executing a Request

Once a request is prepared, use the `Execute` function to send the request and decode the response. The function returns a pointer to the result if the request was successful, or an error if it failed.

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/LiquidCats/jsonrpc" // adjust the import path according to your project structure
)

func main() {
	// Prepare a request (parameters and method details)
	type Params struct {
		Value int `json:"value"`
	}

	params := Params{Value: 123}
	url := "https://your.rpc"
	method := "exampleMethod"

	req, err := jsonrpc.Prepare(context.Background(), url, method, params)
	if err != nil {
		log.Fatalf("Failed to prepare request: %v", err)
	}

	// Execute the JSON-RPC request. For example, expecting a result of type string.
	result, err := jsonrpc.Execute[string](req)
	if err != nil {
		log.Fatalf("Request execution failed: %v", err)
	}

	fmt.Printf("Received result: %s\n", *result)
}
```

## Examples

### Successful Request Example

This example demonstrates a complete flow where a JSON-RPC request is prepared and executed successfully. It assumes the server responds with a valid JSON-RPC result.

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/LiquidCats/jsonrpc" // adjust the import path accordingly
)

func main() {
	// Define parameters (if any)
	type Params struct {
		Value int `json:"value"`
	}
	params := Params{Value: 456}

	// Prepare the JSON-RPC request
	req, err := jsonrpc.Prepare(context.Background(), "https://your.rpc", "testMethod", params)
	if err != nil {
		log.Fatalf("Error preparing request: %v", err)
	}

	// Execute the request expecting a string result
	result, err := jsonrpc.Execute[string](req)
	if err != nil {
		log.Fatalf("Error executing request: %v", err)
	}

	fmt.Printf("Server response: %s\n", *result)
}
```

### Error Handling Example

This example shows how to handle errors when the JSON-RPC response contains an error.

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/LiquidCats/jsonrpc"
)

func main() {
	// Prepare a JSON-RPC request with dummy parameters
	type Params struct{}
	params := Params{}

	req, err := jsonrpc.Prepare(context.Background(), "https://yuor.rpc", "errorMethod", params)
	if err != nil {
		log.Fatalf("Error preparing request: %v", err)
	}

	// Execute the request, expecting an error
	result, err := jsonrpc.Execute[string](req)
	if err != nil {
		fmt.Printf("Received expected error: %v\n", err)
	} else {
		log.Fatalf("Expected error, but got result: %v", *result)
	}
}
```

## API Reference

### `Prepare`

```go
func Prepare[P any](ctx context.Context, url, method string, params P) (*http.Request, error)
```

- **Description:** Constructs an HTTP POST request with a JSON-RPC compliant payload.
- **Parameters:**
    - `ctx`: The context for the HTTP request.
    - `url`: The target endpoint URL.
    - `method`: The JSON-RPC method to be invoked.
    - `params`: The parameters to be included in the JSON-RPC request.
- **Returns:** A pointer to an `http.Request` or an error if the request could not be constructed.

### `Execute`

```go
func Execute[Result any](request *http.Request) (*Result, error)
```

- **Description:** Sends the prepared HTTP request, decodes the JSON-RPC response, and handles any errors.
- **Parameters:**
    - `request`: The HTTP request created by `Prepare`.
- **Returns:** A pointer to the decoded result or an error if the execution or decoding fails.

## Contributing

Contributions are welcome! If you have suggestions, bug fixes, or improvements, please submit an issue or create a pull request.

## License

This project is licensed under the GNU Affero General Public License v3.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgements

- **[sonic](https://github.com/bytedance/sonic):** For providing high-performance JSON encoding and decoding.
- **[go-faster/errors](https://github.com/go-faster/errors):** For enhanced error handling support.
