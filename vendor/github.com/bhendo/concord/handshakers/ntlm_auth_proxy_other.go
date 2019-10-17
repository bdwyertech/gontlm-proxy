// +build !windows

package handshakers

import (
	"net"
	"net/http"
)

func handshake(authRes *http.Response, req *http.Request, conn net.Conn) (*http.Response, error) {
	return authRes, nil
}
