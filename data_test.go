package jsonrpc_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/LiquidCats/jsonrpc"
)

func TestRPCError_ErrorString(t *testing.T) {
	err := (&jsonrpc.RPCError{Code: 123, Message: "boom"}).Error()
	assert.Equal(t, "jsonrpc error: code=123, message=boom", err)
}

func TestCreateRequest_SetsFields(t *testing.T) {
	req := jsonrpc.CreateRequest[int]("sum", []int{1, 2})
	require.NotNil(t, req)
	assert.Equal(t, "2.0", req.JSONRPC)
	assert.Equal(t, "sum", req.Method)
	assert.Equal(t, []int{1, 2}, req.Params)

	// ID should be a millisecond timestamp; just check it's non-zero and within a reasonable range.
	now := time.Now().UnixMilli()
	assert.Greater(t, req.ID, int64(0))
	assert.InDelta(t, float64(now), float64(req.ID), float64(10_000)) // within ~10s
}
