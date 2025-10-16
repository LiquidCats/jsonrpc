package jsonrpc_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/LiquidCats/jsonrpc/tests/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/LiquidCats/jsonrpc"
)

type user struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestExecute_Success(t *testing.T) {
	resp := jsonrpc.RPCResponse[user]{
		JSONRPC: "2.0",
		Result:  user{ID: 7, Name: "alice"},
		ID:      123,
	}
	b, err := json.Marshal(resp)
	require.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Greater(t, r.ContentLength, int64(0))
		_, _ = io.Copy(io.Discard, r.Body)
		_ = r.Body.Close()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(b)
	}))
	defer ts.Close()

	req := jsonrpc.CreateRequest[user]("user.get", map[string]any{"id": 7})
	got, err := req.Execute(context.Background(), ts.URL)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, 7, got.ID)
	assert.Equal(t, "alice", got.Name)
}

func TestExecute_JSONRPCErrorInBody(t *testing.T) {
	resp := jsonrpc.RPCResponse[user]{
		JSONRPC: "2.0",
		Error:   &jsonrpc.RPCError{Code: -32001, Message: "backend unavailable"},
		ID:      1,
	}
	b, err := json.Marshal(resp)
	require.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(b)
	}))
	defer ts.Close()

	req := jsonrpc.CreateRequest[user]("user.get", nil)
	got, err := req.Execute(context.Background(), ts.URL)
	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, "jsonrpc error: code=-32001, message=backend unavailable", err.Error())
}

func TestExecute_HTTPNon2xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write(bytes.Repeat([]byte("x"), 2048))
	}))
	defer ts.Close()

	req := jsonrpc.CreateRequest[user]("user.get", nil)
	got, err := req.Execute(context.Background(), ts.URL)
	require.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "http status 418")
}

func TestExecute_RequestAndClientOptions(t *testing.T) {
	var sawHeader atomic.Bool
	const customUA = "my-agent/1.0"
	var clientUsed atomic.Bool

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == customUA {
			sawHeader.Store(true)
		}
		resp := jsonrpc.RPCResponse[user]{JSONRPC: "2.0", Result: user{ID: 1, Name: "ok"}, ID: 1}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	reqOpt := func(r *http.Request) {
		r.Header.Set("User-Agent", customUA)
	}
	clientOpt := func(c *http.Client) {
		clientUsed.Store(true)
		c.Timeout = time.Second * 5
	}

	req := jsonrpc.CreateRequest[user]("user.get", nil)
	got, err := req.Execute(context.Background(), ts.URL, reqOpt, clientOpt)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.True(t, sawHeader.Load(), "request option did not set header")
	assert.True(t, clientUsed.Load(), "client option was not used")
	assert.Equal(t, "ok", got.Name)
}

func TestExecute_DecodeError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{not-json"))
	}))
	defer ts.Close()

	req := jsonrpc.CreateRequest[user]("user.get", nil)
	got, err := req.Execute(context.Background(), ts.URL)
	require.Error(t, err)
	assert.Nil(t, got)
}

func BenchmarkExecute_Success(b *testing.B) {
	file, err := os.ReadFile("tests/fixtures/btc-block-without-txs.json")
	require.NoError(b, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Minimal work in handler; just serve the fixed payload.
		_, _ = io.Copy(io.Discard, r.Body)
		_ = r.Body.Close()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_, _ = w.Write(file)
	}))
	defer ts.Close()

	ctx := context.Background()
	req := jsonrpc.CreateRequest[types.Block]("user.get", map[string]any{"id": 42})

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := req.Execute(ctx, ts.URL)
		require.NoError(b, err)
	}
}

func BenchmarkExecute_WithOptions(b *testing.B) {
	file, err := os.ReadFile("tests/fixtures/btc-block-without-txs.json")
	require.NoError(b, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		_ = r.Body.Close()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(file)
	}))
	defer ts.Close()

	// Lightweight request option
	reqOpt := func(r *http.Request) {
		r.Header.Set("X-Bench", "1")
	}
	// Lightweight client option
	clientOpt := func(c *http.Client) {
		c.Timeout = 30 * time.Second
	}

	ctx := context.Background()
	req := jsonrpc.CreateRequest[types.Block]("user.get", nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := req.Execute(ctx, ts.URL, reqOpt, clientOpt)
		require.NoError(b, err)
	}
}

func BenchmarkExecute_Parallel(b *testing.B) {
	file, err := os.ReadFile("tests/fixtures/btc-block-without-txs.json")
	require.NoError(b, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		_ = r.Body.Close()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(file)
	}))
	defer ts.Close()

	ctx := context.Background()
	req := jsonrpc.CreateRequest[types.Block]("user.get", map[string]any{"id": 42})

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := req.Execute(ctx, ts.URL)
			require.NoError(b, err)
		}
	})
}
