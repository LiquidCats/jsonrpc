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
		return bytes.NewBuffer(nil)
	},
}

type Option func()
type RequestOption func(in *http.Request)
type ClientOption func(in *http.Client)

func (rpc *RPCRequest[Resp]) Execute(ctx context.Context, url string, opts ...any) (*Resp, error) {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	encoder := sonic.ConfigFastest.NewEncoder(buf)
	if err := encoder.Encode(rpc); err != nil {
		return nil, eris.Wrap(err, "marshal request")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, buf)
	if err != nil {
		return nil, eris.Wrap(err, "prepare req")
	}
	req.ContentLength = int64(buf.Len())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	cl := http.DefaultClient
	for _, opt := range opts {
		switch v := opt.(type) {
		case RequestOption:
			v(req)
		case ClientOption:
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

	dec := sonic.ConfigFastest.NewDecoder(resp.Body)

	dec.Buffered()

	var result RPCResponse[Resp]
	if err = dec.Decode(&result); err != nil {
		return nil, eris.Wrap(err, "decode response")
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return &result.Result, nil
}
