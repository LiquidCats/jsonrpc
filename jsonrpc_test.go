package jsonrpc_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jsonrpc "github.com/LiquidCats/jsonrpc/v2"
	"github.com/stretchr/testify/require"
)

func TestNewRequestDefaults(t *testing.T) {
	t.Parallel()

	params := map[string]int{"value": 42}
	req := jsonrpc.NewRequest[map[string]int, string]("add", params)

	require.Equal(t, "add", req.Method)
	require.Equal(t, params, req.Params)
	require.Equal(t, jsonrpc.Version, req.JSONRPC)
	require.NotEmpty(t, req.ID)
}

func TestNewRequestWithOptions(t *testing.T) {
	t.Parallel()

	req := jsonrpc.NewRequest[[]string, string](
		"echo",
		[]string{"hello"},
		jsonrpc.WithRPCVersion[[]string, string]("1.1"),
		jsonrpc.WithRPCid[[]string, string]("custom-id"),
	)

	require.Equal(t, "echo", req.Method)
	require.Equal(t, []string{"hello"}, req.Params)
	require.Equal(t, "1.1", req.JSONRPC)
	require.Equal(t, "custom-id", req.ID)
}

func TestPrepareAndExecuteSuccess(t *testing.T) {
	t.Parallel()

	params := map[string]any{"value": 123}
	req := jsonrpc.NewRequest[map[string]any, map[string]string](
		"echo",
		params,
		jsonrpc.WithRPCid[map[string]any, map[string]string]("req-1"),
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "application/custom-json", r.Header.Get("Content-Type"))
		require.Equal(t, "token", r.Header.Get("Authorization"))

		var decoded struct {
			Method  string         `json:"method"`
			Params  map[string]int `json:"params"`
			ID      any            `json:"id"`
			JSONRPC string         `json:"jsonrpc"`
		}

		require.NoError(t, json.Unmarshal(body, &decoded))
		require.Equal(t, req.Method, decoded.Method)
		require.Equal(t, 123, decoded.Params["value"])
		require.Equal(t, req.ID, decoded.ID)
		require.Equal(t, req.JSONRPC, decoded.JSONRPC)

		resp := jsonrpc.RPCResponse[map[string]string]{
			JSONRPC: jsonrpc.Version,
			Result:  map[string]string{"status": "ok"},
			ID:      decoded.ID,
		}

		require.NoError(t, json.NewEncoder(w).Encode(resp))
	}))
	defer server.Close()

	ctx := context.WithValue(context.Background(), "trace", "abc123")

	prepared := req.Prepare(
		server.URL,
		jsonrpc.WithContext(ctx),
		jsonrpc.WithHeader("Authorization", "token"),
		jsonrpc.WithContentType("application/custom-json"),
	)

	result, err := prepared.Execute(server.Client())
	require.NoError(t, err)
	require.Equal(t, map[string]string{"status": "ok"}, *result)
}

func TestExecuteHTTPStatusError(t *testing.T) {
	t.Parallel()

	req := jsonrpc.NewRequest[struct{}, string](
		"boom",
		struct{}{},
		jsonrpc.WithRPCid[struct{}, string]("req-2"),
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))
	defer server.Close()

	prepared := req.Prepare(server.URL)

	_, err := prepared.Execute(server.Client())
	require.Error(t, err)
	require.Contains(t, err.Error(), "http status 418")
}

func TestExecuteRPCError(t *testing.T) {
	t.Parallel()

	req := jsonrpc.NewRequest[struct{}, string](
		"rpc_error",
		struct{}{},
		jsonrpc.WithRPCid[struct{}, string]("rpc-err"),
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"jsonrpc":"2.0","error":{"code":-32000,"message":"boom"},"id":"rpc-err"}`)
	}))
	defer server.Close()

	prepared := req.Prepare(server.URL)

	_, err := prepared.Execute(server.Client())
	require.Error(t, err)

	var rpcErr *jsonrpc.RPCError
	require.ErrorAs(t, err, &rpcErr)
	require.Equal(t, -32000, rpcErr.Code)
	require.Equal(t, "boom", rpcErr.Message)
}

func TestExecuteDecodeError(t *testing.T) {
	t.Parallel()

	req := jsonrpc.NewRequest[struct{}, string](
		"bad_json",
		struct{}{},
		jsonrpc.WithRPCid[struct{}, string]("decode"),
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "{not-json")
	}))
	defer server.Close()

	prepared := req.Prepare(server.URL)

	_, err := prepared.Execute(server.Client())
	require.Error(t, err)
	require.Contains(t, err.Error(), "decode response")
}

func TestExecuteCanceledContext(t *testing.T) {
	t.Parallel()

	req := jsonrpc.NewRequest[struct{}, string](
		"ctx",
		struct{}{},
		jsonrpc.WithRPCid[struct{}, string]("ctx-1"),
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	prepared := req.Prepare(server.URL, jsonrpc.WithContext(ctx))
	cancel()

	client := server.Client()
	client.Timeout = 200 * time.Millisecond

	_, err := prepared.Execute(client)
	require.Error(t, err)
	require.Contains(t, err.Error(), "context canceled")
}

func TestExecuteClientOptions(t *testing.T) {
	t.Parallel()

	req := jsonrpc.NewRequest[struct{}, string](
		"opt",
		struct{}{},
		jsonrpc.WithRPCid[struct{}, string]("opt-1"),
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"jsonrpc":"2.0","result":"","id":"opt-1"}`)
	}))
	defer server.Close()

	client := server.Client()

	var called bool
	opt := func(cli *http.Client) {
		called = true
		cli.Timeout = 123 * time.Millisecond
	}

	prepared := req.Prepare(server.URL)

	_, err := prepared.Execute(client, opt)
	require.NoError(t, err)
	require.True(t, called, "execute option should be applied")
	require.Equal(t, 123*time.Millisecond, client.Timeout)
}

func TestPrepareErrorPropagates(t *testing.T) {
	t.Parallel()

	req := jsonrpc.NewRequest[struct{}, string]("invalid-url", struct{}{})

	prepared := req.Prepare("://bad url")

	_, err := prepared.Execute(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "execute")
	require.Contains(t, err.Error(), "create http request")
}
