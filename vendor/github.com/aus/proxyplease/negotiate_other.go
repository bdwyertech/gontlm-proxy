// +build !windows

package proxyplease

import (
	"errors"
	"net"
)

func dialNegotiate(p Proxy, addr string, baseDial func() (net.Conn, error)) (net.Conn, error) {
	return nil, errors.New("negotiate proxy authentication is only available on Windows")
}
