package jsonrpc

import (
	"fmt"
	"time"
)

const Version = "2.0"

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

type RPCRequest[Resp any] struct {
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
	ID      int64  `json:"id"`
	JSONRPC string `json:"jsonrpc"`
}

func CreateRequest[Result any](method string, params any) *RPCRequest[Result] {
	return &RPCRequest[Result]{
		ID:      time.Now().UnixMilli(),
		Method:  method,
		JSONRPC: Version,
		Params:  params,
	}
}
