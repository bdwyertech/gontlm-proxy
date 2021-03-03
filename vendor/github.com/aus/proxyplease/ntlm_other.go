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

	"github.com/Azure/go-ntlmssp"
	"github.com/git-lfs/go-ntlm/ntlm"
)

func dialNTLM(p Proxy, addr string, baseDial func() (net.Conn, error)) (net.Conn, error) {
	debugf("ntlm> Attempting to authenticate")

	conn, err := baseDial()
	if err != nil {
		debugf("ntlm> Could not call dial context with proxy: %s", err)
		return conn, err
	}

	session, err := ntlm.CreateClientSession(ntlm.Version1, ntlm.ConnectionlessMode)
	if err != nil {
		debugf("ntlm> Error creating client session")
		return conn, err
	}

	session.SetUserInfo(p.Username, p.Password, p.Domain)

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

	challenge, err := ntlm.ParseChallengeMessage(challengeBytes)
	if err != nil {
		debugf("ntlm> Error parsing challenge message")
		return conn, err
	}

	err = session.ProcessChallengeMessage(challenge)
	if err != nil {
		debugf("ntlm> Error processing challenge messgage")
		return conn, err
	}

	authenticate, err := session.GenerateAuthenticateMessage()
	if err != nil {
		debugf("ntlm> Error generating  authenticate message")
		return conn, err
	}

	resp.Body.Close()
	h = p.Headers.Clone()
	h.Set("Proxy-Authorization", fmt.Sprintf("NTLM %s", base64.StdEncoding.EncodeToString(authenticate.Bytes())))
	h.Set("Proxy-Connection", "Keep-Alive")
	connect = &http.Request{
		Method: "CONNECT",
		URL:    &url.URL{Opaque: addr},
		Host:   addr,
		Header: h,
	}
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

	if resp.StatusCode == http.StatusOK {
		debugf("ntlm> Successfully injected NTLM to connection")
		return conn, nil
	}

	debugf("ntlm> Expected %d as return status, got: %d", http.StatusOK, resp.StatusCode)
	return conn, errors.New(http.StatusText(resp.StatusCode))

}
