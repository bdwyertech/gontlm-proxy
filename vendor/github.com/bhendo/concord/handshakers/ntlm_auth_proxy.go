package handshakers

import (
	"net"
	"net/http"
)

// NTLMProxyAuthorizer is an implementation of Handshaker that supports
// proxies that require NTLM authentication. Windows only.
type NTLMProxyAuthorizer struct{}

// Handshake implements the Handshaker interface
func (a *NTLMProxyAuthorizer) Handshake(authRes *http.Response, req *http.Request, conn net.Conn) (*http.Response, error) {
	return handshake(authRes, req, conn)
}
