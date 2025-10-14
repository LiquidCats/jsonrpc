package jsonrpc

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/rotisserie/eris"
)

func (rpc *RPCRequest[Result]) Execute(ctx context.Context, url string, opts ...any) (*Result, error) {
	rpcBytes, err := sonic.Marshal(rpc)
	if err != nil {
		return nil, eris.Wrap(err, "marshal request")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(rpcBytes))
	if err != nil {
		return nil, eris.Wrap(err, "prepare req")
	}
	req.ContentLength = int64(len(rpcBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	for _, opt := range opts {
		switch v := opt.(type) {
		case func(r *http.Request):
			v(req)
		case func(r *http.Client):
			v(defaultHTTPClient)
		}
	}

	resp, err := defaultHTTPClient.Do(req)
	if err != nil {
		return nil, eris.Wrap(err, "execute req")
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		_, _ = io.CopyN(io.Discard, resp.Body, 1024)
		return nil, eris.Errorf("http status %d", resp.StatusCode)
	}

	dec := sonic.ConfigStd.NewDecoder(resp.Body)
	defer func() {
		_ = resp.Body.Close()
	}()

	var result RPCResponse[Result]

	if err = dec.Decode(&result); err != nil {
		return nil, eris.Wrap(err, "decode response")
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return &result.Result, nil
}
