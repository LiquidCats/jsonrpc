package jsonrpc

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

var defaultHTTPClient = &http.Client{
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2: true,

		// Large response tuning: allow many idle conns but cap concurrency
		MaxIdleConns:        4_096,
		MaxIdleConnsPerHost: 1_024,
		MaxConnsPerHost:     512, // cap to bound memory while handling many large bodies

		// Keep generous idle timeout for reuse; large responses take longer
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 250 * time.Millisecond,

		// Enable compression to save bandwidth for multi-MB JSON; CPU tradeoff
		DisableCompression: false,

		// Increase per-connection IO buffers to reduce syscalls for large bodies
		// Defaults are 4KB; bump to 64KB (common page multiple).
		ReadBufferSize:  64 << 10,
		WriteBufferSize: 64 << 10,

		// TLS session cache to lower handshake CPU when many connections exist
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12,
			ClientSessionCache: tls.NewLRUClientSessionCache(4096),
		},

		// HTTP/2: raise concurrent streams per connection for multiplexing large responses
		// (Go picks defaults; env GODEBUG may tune; leaving default to avoid incompat issues)
	},
	Timeout: 0,
}
