// +build !windows

package main

import (
	"context"
	"net"
)

// STUB - Wraps a DialContext with NTLM Authentication to a proxy
func WrapDialContext(dialContext DialContext, proxyAddress string) DialContext {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialContext(ctx, network, proxyAddress)
	}
}
