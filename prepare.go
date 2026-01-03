package jsonrpc

import (
	"bytes"
	"context"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/rotisserie/eris"
)

type PrepareOpt func(*http.Request)

func WithContext(ctx context.Context) PrepareOpt {
	return func(r *http.Request) {
		*r = *r.WithContext(ctx)
	}
}

func WithContentType(contentType string) PrepareOpt {
	return func(r *http.Request) {
		r.Header.Set("Content-Type", contentType)
	}
}

func WithHeader(key, value string) PrepareOpt {
	return func(r *http.Request) {
		r.Header.Set(key, value)
	}
}

func (r *rpcRequest[Params, Resp]) Prepare(url string, opts ...PrepareOpt) *praparedRPCRequest[Resp] {
	buff := bytes.NewBuffer(nil)

	encoder := sonic.ConfigDefault.NewEncoder(buff)

	if err := encoder.Encode(r); err != nil {
		return &praparedRPCRequest[Resp]{err: eris.Wrap(err, "encode request data")}
	}

	req, err := http.NewRequest(http.MethodPost, url, buff)
	if err != nil {
		return &praparedRPCRequest[Resp]{err: eris.Wrap(err, "create http request")}
	}

	req.Header.Set("Content-Type", "application/json")

	for _, opt := range opts {
		opt(req)
	}

	return &praparedRPCRequest[Resp]{internal: req}
}
