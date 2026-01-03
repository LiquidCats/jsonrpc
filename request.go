package jsonrpc

import (
	"strconv"
	"time"
)

const Version = "2.0"

type rpcRequest[Params any, Resp any] struct {
	Method  string `json:"method"`
	Params  Params `json:"params,omitempty"`
	ID      any    `json:"id"`
	JSONRPC string `json:"jsonrpc"`
}

type RPCOpt[Params any, Resp any] func(*rpcRequest[Params, Resp])

func WithRPCVersion[Params any, Resp any](version string) RPCOpt[Params, Resp] {
	return func(req *rpcRequest[Params, Resp]) {
		req.JSONRPC = version
	}
}

func WithRPCid[Params any, Resp any](id any) RPCOpt[Params, Resp] {
	return func(req *rpcRequest[Params, Resp]) {
		req.ID = id
	}
}

func NewRequest[Params any, Result any](method string, params Params, opts ...RPCOpt[Params, Result]) *rpcRequest[Params, Result] {
	req := &rpcRequest[Params, Result]{
		ID:      strconv.FormatInt(time.Now().UnixNano(), 10),
		Method:  method,
		JSONRPC: Version,
		Params:  params,
	}

	for _, opt := range opts {
		opt(req)
	}

	return req
}
