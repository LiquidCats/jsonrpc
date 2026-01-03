package jsonrpc

import "fmt"

type RPCResponse[D any] struct {
	JSONRPC string    `json:"jsonrpc"`
	Result  D         `json:"result"`
	Error   *RPCError `json:"error,omitempty"`
	ID      any       `json:"id"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("jsonrpc error: code=%d, message=%s", e.Code, e.Message)
}
