package jsonrpc

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bytedance/sonic"
	"github.com/go-faster/errors"
)

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

type RPCRequest[P any] struct {
	Method  string `json:"method"`
	Params  P      `json:"params,omitempty"`
	ID      int64  `json:"id"`
	JSONRPC string `json:"jsonrpc"`
}

func createRequest[P any](method string, params P) *RPCRequest[P] {
	return &RPCRequest[P]{
		ID:      time.Now().UnixMilli(),
		Method:  method,
		JSONRPC: "2.0",
		Params:  params,
	}
}

type Request = http.Request

func Prepare[P any](ctx context.Context, url, method string, params P) (*Request, error) {
	buff := bytes.NewBuffer([]byte{})

	encoder := sonic.ConfigDefault.NewEncoder(buff)
	if err := encoder.Encode(createRequest[P](method, params)); err != nil {
		return nil, errors.Wrap(err, "encode request")
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, buff)
	if err != nil {
		return nil, errors.Wrap(err, "new request")
	}

	return request, nil
}

func Execute[Result any](request *Request) (*Result, error) {
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, errors.Wrap(err, "execute request")
	}

	decoder := sonic.ConfigDefault.NewDecoder(response.Body)
	defer func() {
		_ = response.Body.Close()
	}()

	var result RPCResponse[Result]

	if err = decoder.Decode(&result); err != nil {
		return nil, errors.Wrap(err, "decode response")
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return &result.Result, nil
}
