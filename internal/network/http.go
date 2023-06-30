package network

import (
	"net"
	"net/http"
	"time"
)

var c = http.Client{}
var DefaultClientTransport = &http.Transport{
	DialContext: (&net.Dialer{
		Timeout:   60 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,
	MaxIdleConns:        1000,
	MaxIdleConnsPerHost: 1000,
	IdleConnTimeout:     120 * time.Second,
	TLSHandshakeTimeout: 30 * time.Second,
}
