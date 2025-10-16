package jsonrpc

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/rotisserie/eris"
)

var bufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 512))
	},
}

func UseClient(cli *http.Client) func(in *http.Client) {
	return func(in *http.Client) {
		in = cli
	}
}

func SetHeader(key, value string) func(in *http.Request) {
	return func(in *http.Request) {
		in.Header.Set(key, value)
	}
}

func UseContext(ctx context.Context) func(in *http.Request) {
	return func(in *http.Request) {
		in = in.WithContext(ctx)
	}
}

func (rpc *RPCRequest[Resp]) Execute(url string, opts ...any) (*Resp, error) {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	enc := sonic.ConfigFastest.NewEncoder(buf)
	if err := enc.Encode(rpc); err != nil {
		return nil, eris.Wrap(err, "marshal request")
	}

	req, err := http.NewRequest(http.MethodPost, url, buf)
	if err != nil {
		return nil, eris.Wrap(err, "prepare req")
	}
	req.ContentLength = int64(buf.Len())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	cl := defaultHTTPClient
	for _, opt := range opts {
		switch v := opt.(type) {
		case func(in *http.Request):
			v(req)
		case func(in *http.Client):
			v(cl)
		}
	}

	resp, err := cl.Do(req)
	if err != nil {
		return nil, eris.Wrap(err, "execute req")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		_, _ = io.CopyN(io.Discard, resp.Body, 1024)
		return nil, eris.Errorf("http status %d", resp.StatusCode)
	}

	var result RPCResponse[Resp]
	err = sonic.ConfigFastest.Unmarshal(readAll(resp.Body), &result)
	if err != nil {
		return nil, eris.Wrap(err, "decode response")
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return &result.Result, nil
}

// readAll efficiently reads response body
func readAll(r io.Reader) []byte {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	buf.ReadFrom(r)
	// Return copy to avoid buffer pool corruption
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())
	return result
}
