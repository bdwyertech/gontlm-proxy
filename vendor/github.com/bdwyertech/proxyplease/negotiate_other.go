//go:build !windows

package proxyplease

import (
	"bufio"
	"errors"
	"net"
)

func dialNegotiate(p Proxy, addr string, _ net.Conn, _ *bufio.Reader) (net.Conn, error) {
	return nil, errors.New("negotiate proxy authentication is only available on Windows")
}
