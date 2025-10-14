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
			Timeout:   2 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,

		ForceAttemptHTTP2:   true,
		TLSHandshakeTimeout: 3 * time.Second,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},

		DisableCompression:    true,
		ExpectContinueTimeout: 0,

		MaxIdleConns:        1024,
		MaxIdleConnsPerHost: 1024,

		IdleConnTimeout: 120 * time.Second,

		WriteBufferSize: 64 << 10, // 64 KiB
		ReadBufferSize:  64 << 10, // 64 KiB
	},
	Timeout: 0,
}
