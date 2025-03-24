package jsonrpc_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LiquidCats/jsonrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrepare(t *testing.T) {
	params := []any{"0.001", true, float64(0)}
	method := "testMethod"
	url := "http://example.com"
	ctx := context.Background()

	req, err := jsonrpc.Prepare[[]any](ctx, url, method, params)
	if err != nil {
		t.Fatalf("Prepare returned error: %v", err)
	}

	// Read the request body
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("Failed to read request body: %v", err)
	}
	defer func(t *testing.T) {
		err := req.Body.Close()
		require.NoError(t, err)
	}(t)

	// Unmarshal the JSON payload
	var rpcReq map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &rpcReq); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	paramsData, ok := rpcReq["params"].([]any)
	assert.True(t, ok)
	for i := range paramsData {
		assert.Equal(t, params[i], paramsData[i])
	}

	assert.Equal(t, "2.0", rpcReq["jsonrpc"])
	assert.Equal(t, method, rpcReq["method"])
	assert.IsType(t, float64(0), rpcReq["id"])
}

func TestExecuteSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func(t *testing.T) {
			assert.NoError(t, r.Body.Close())
		}(t)

		var reqData map[string]interface{}

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&reqData)
		assert.NoError(t, err)

		response := map[string]interface{}{
			"jsonrpc": "2.0",
			"result":  42,
			"id":      reqData["id"],
		}

		w.Header().Set("Content-Type", "application/json")

		encoder := json.NewEncoder(w)
		err = encoder.Encode(response)
		assert.NoError(t, err)
	}))
	defer ts.Close()

	// Prepare the request
	params := struct{}{} // no parameters required
	req, err := jsonrpc.Prepare(context.Background(), ts.URL, "dummyMethod", params)
	require.NoError(t, err)

	result, err := jsonrpc.Execute[int](req)
	require.NoError(t, err)

	assert.Equal(t, 42, *result)
}

func TestExecuteError(t *testing.T) {
	// Setup a test server that returns an error JSON-RPC response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func(t *testing.T) {
			assert.NoError(t, r.Body.Close())
		}(t)

		var reqData map[string]interface{}

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&reqData)
		assert.NoError(t, err)

		w.Header().Set("Content-Type", "application/json")

		response := map[string]interface{}{
			"jsonrpc": "2.0",
			"error": map[string]interface{}{
				"code":    -32000,
				"message": "Test error",
			},
			"id": reqData["id"],
		}

		encoder := json.NewEncoder(w)
		err = encoder.Encode(response)
		assert.NoError(t, err)
	}))
	defer ts.Close()

	// Prepare the request
	params := struct{}{}
	req, err := jsonrpc.Prepare(context.Background(), ts.URL, "dummyMethod", params)
	if err != nil {
		t.Fatalf("Prepare returned error: %v", err)
	}

	result, err := jsonrpc.Execute[int](req)
	require.Error(t, err)

	expectedError := "jsonrpc error: code=-32000, message=Test error"
	assert.Equal(t, expectedError, err.Error())

	assert.Nil(t, result)
}
