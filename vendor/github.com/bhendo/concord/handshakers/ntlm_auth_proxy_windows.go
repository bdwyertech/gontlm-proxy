// +build windows

package handshakers

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/alexbrainman/sspi/ntlm"
)

func handshake(authRes *http.Response, req *http.Request, conn net.Conn) (*http.Response, error) {
	if authRes.Body != nil {
		if err := authRes.Body.Close(); err != nil {
			return nil, err
		}
	}
	cred, err := ntlm.AcquireCurrentUserCredentials()
	if err != nil {
		return nil, err
	}
	defer cred.Release()
	secctx, negotiate, err := ntlm.NewClientContext(cred)
	if err != nil {
		return nil, err
	}
	defer secctx.Release()
	challenge, err := doNTLMNegotiate(req, conn, negotiate)
	if err != nil {
		return nil, err
	}
	authenticate, err := secctx.Update(challenge)
	if err != nil {
		return nil, err
	}
	return doNTLMAuthenticate(req, conn, authenticate)
}

func doNTLMNegotiate(req *http.Request, conn net.Conn, negotiate []byte) ([]byte, error) {
	// set proxy headers
	req.Header.Set("Proxy-Connection", "Keep-Alive")
	req.Header.Set("Proxy-Authorization", fmt.Sprintf("NTLM %s", base64.StdEncoding.EncodeToString(negotiate)))
	if err := req.WriteProxy(conn); err != nil {
		return nil, err
	}
	res, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		return nil, err
	}
	if err := res.Body.Close(); err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusProxyAuthRequired {
		return nil, fmt.Errorf("did not receive a challenge from the server. received status code: %d", res.StatusCode)
	}
	challengeHeaders, found := res.Header["Proxy-Authenticate"]
	if !found {
		return nil, fmt.Errorf("did not receive a challenge from the server")
	}
	if len(challengeHeaders) != 1 {
		return nil, fmt.Errorf("received malformed challenge from the server")
	}
	if len(challengeHeaders[0]) < 6 || !strings.HasPrefix(challengeHeaders[0], "NTLM ") {
		return nil, fmt.Errorf("received malformed challenge from the server")
	}

	// Rewind the request body, the handshake needs it
	if req.GetBody != nil {
		if req.Body, err = req.GetBody(); err != nil {
			return nil, err
		}
	}
	return base64.StdEncoding.DecodeString(challengeHeaders[0][5:])
}

func doNTLMAuthenticate(req *http.Request, conn net.Conn, authenticate []byte) (*http.Response, error) {
	req.Header.Set("Proxy-Connection", "Keep-Alive")
	req.Header.Set("Proxy-Authorization", fmt.Sprintf("NTLM %s", base64.StdEncoding.EncodeToString(authenticate)))
	if err := req.WriteProxy(conn); err != nil {
		return nil, err
	}
	return http.ReadResponse(bufio.NewReader(conn), req)
}
