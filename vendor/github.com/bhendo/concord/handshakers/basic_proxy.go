package handshakers

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"strings"
)

// BasicProxyAuthorizer is an implementation of Handshaker that supports
// proxies that require basic authentication.
type BasicProxyAuthorizer struct {
	Username string
	Password string
}

// Handshake implements the Handshaker interface
func (p *BasicProxyAuthorizer) Handshake(authRes *http.Response, req *http.Request, conn net.Conn) (*http.Response, error) {
	if authRes.Body != nil {
		if err := authRes.Body.Close(); err != nil {
			return nil, err
		}
	}
	authHeaders, found := authRes.Header["Proxy-Authenticate"]
	if !found {
		return nil, fmt.Errorf("did not receive a proxy-authenticate response from the server")
	}
	if len(authHeaders) != 1 {
		return nil, fmt.Errorf("received malformed proxy-authenticate response header from the server")
	}
	if len(authHeaders[0]) < 7 || !strings.HasPrefix(authHeaders[0], "Basic ") {
		return nil, fmt.Errorf("cannot authenticate to this proxy server")
	}

	req.Header.Set("Proxy-Connection", "Keep-Alive")
	req.Header.Set(
		"Proxy-Authorization",
		fmt.Sprintf(
			"Basic %s",
			base64.StdEncoding.EncodeToString(
				[]byte(fmt.Sprintf("%s:%s", p.Username, p.Password)),
			),
		),
	)
	if err := req.WriteProxy(conn); err != nil {
		return nil, err
	}
	return http.ReadResponse(bufio.NewReader(conn), req)
}
