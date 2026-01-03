package jsonrpc

import (
	"io"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/rotisserie/eris"
)

type praparedRPCRequest[Resp any] struct {
	internal *http.Request
	err      error
}

type ExecuteOpt func(*http.Client)

func (rpc *praparedRPCRequest[Resp]) Execute(client *http.Client, opts ...ExecuteOpt) (*Resp, error) {
	if rpc.err != nil {
		return nil, eris.Wrap(rpc.err, "execute prepared request")
	}

	cli := client
	if client == nil {
		cli = defaultHTTPClient
	}

	for _, opt := range opts {
		opt(cli)
	}

	resp, err := cli.Do(rpc.internal)
	if err != nil {
		return nil, eris.Wrap(err, "execute req")
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil, eris.Errorf("http status %d", resp.StatusCode)
	}

	var result RPCResponse[Resp]

	decoder := sonic.ConfigDefault.NewDecoder(resp.Body)
	if err := decoder.Decode(&result); err != nil {
		return nil, eris.Wrap(err, "decode response")
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return &result.Result, nil
}
