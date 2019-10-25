package ntlm_proxy

import (
	"context"
	"net"
)

// DialContext is the DialContext function that should be wrapped with NTLM Authentication
//
// Example for DialContext:
//
// dialContext := (&net.Dialer{KeepAlive: 30*time.Second, Timeout: 30*time.Second}).DialContext
type DialContext func(ctx context.Context, network, addr string) (net.Conn, error)
