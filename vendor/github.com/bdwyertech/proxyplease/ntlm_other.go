// +build !windows

package proxyplease

import (
	"bufio"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/launchdarkly/go-ntlmssp"
)

func dialNTLM(p Proxy, addr string, baseDial func() (net.Conn, error)) (net.Conn, error) {
	debugf("ntlm> Attempting to authenticate")

	conn, err := baseDial()
	if err != nil {
		debugf("ntlm> Could not call dial context with proxy: %s", err)
		return conn, err
	}

	negotiateMsg, err := ntlmssp.NewNegotiateMessage(p.Domain, p.Username)
	if err != nil {
		debugf("ntlm> Error creating Negotiate message")
		return conn, err
	}

	h := p.Headers.Clone()
	h.Set("Proxy-Authorization", fmt.Sprintf("NTLM %s", base64.StdEncoding.EncodeToString(negotiateMsg)))
	h.Set("Proxy-Connection", "Keep-Alive")
	connect := &http.Request{
		Method: "CONNECT",
		URL:    &url.URL{Opaque: addr},
		Host:   addr,
		Header: h,
	}
	if err := connect.WriteProxy(conn); err != nil {
		debugf("ntlm> Could not write negotiate message to proxy: %s", err)
		return conn, err
	}
	br := bufio.NewReader(conn)
	resp, err := http.ReadResponse(br, connect)
	if err != nil {
		debugf("ntlm> Could not read negotiate response from proxy: %s", err)
		return conn, err
	}
	if err := resp.Body.Close(); err != nil {
		return conn, err
	}

	if resp.StatusCode != http.StatusProxyAuthRequired {
		debugf("ntlm> Expected %d as return status, got: %d", http.StatusProxyAuthRequired, resp.StatusCode)
		return conn, errors.New("unexpected HTTP status code")
	}

	challengeHeaders, found := resp.Header["Proxy-Authenticate"]
	if !found {
		return conn, errors.New("did not receive a challenge from the server")
	}
	if len(challengeHeaders) != 1 {
		return conn, errors.New("received malformed challenge from the server")
	}
	if len(challengeHeaders[0]) < 6 || !strings.HasPrefix(challengeHeaders[0], "NTLM ") {
		return conn, errors.New("received malformed challenge from the server")
	}

	challengeBytes, err := base64.StdEncoding.DecodeString(challengeHeaders[0][5:])
	if err != nil {
		debugf("ntlm> Could not read challenge response")
		return conn, err
	}

	authBytes, err := ntlmssp.ProcessChallenge(challengeBytes, p.Username, p.Password)
	if err != nil {
		debugf("ntlm> Error processing challenge message")
		return conn, err
	}

	// Rewind the request body, the handshake needs it
	if connect.GetBody != nil {
		if connect.Body, err = connect.GetBody(); err != nil {
			return conn, err
		}
	}

	connect.Header.Set("Proxy-Authorization", fmt.Sprintf("NTLM %s", base64.StdEncoding.EncodeToString(authBytes)))

	if err := connect.WriteProxy(conn); err != nil {
		debugf("ntlm> Could not write authenticate message to proxy: %s", err)
		return conn, err
	}
	br = bufio.NewReader(conn)
	resp, err = http.ReadResponse(br, connect)
	if err != nil {
		debugf("ntlm> Could not read authenticate response from proxy: %s", err)
		return conn, err
	}
	resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		debugf("ntlm> Successfully injected NTLM to connection")
		return conn, nil
	}

	debugf("ntlm> Expected %d as return status, got: %d", http.StatusOK, resp.StatusCode)
	return conn, errors.New(http.StatusText(resp.StatusCode))
}
