// +build windows

package proxyplease

import (
	"bufio"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/alexbrainman/sspi"
	"github.com/alexbrainman/sspi/ntlm"
)

func dialNTLM(p Proxy, addr string, baseDial func() (net.Conn, error)) (net.Conn, error) {
	debugf("ntlm> Attempting to authenticate")

	conn, err := baseDial()
	if err != nil {
		debugf("ntlm> Could not call dial context with proxy: %s", err)
		return conn, err
	}

	var cred *sspi.Credentials
	if p.Domain != "" && p.Username != "" && p.Password != "" {
		debugf("ntlm> Using supplied credentials")
		cred, err = ntlm.AcquireUserCredentials(p.Domain, p.Username, p.Password)
	} else {
		debugf("ntlm> No credentials were provided. Assuming current user credentials from SSPI.")
		cred, err = ntlm.AcquireCurrentUserCredentials()
	}
	if err != nil {
		debugf("ntlm> Unable to acquire supplied or current user credentials.")
		return conn, err
	}
	defer cred.Release()

	secctx, negotiate, err := ntlm.NewClientContext(cred)
	if err != nil {
		debugf("ntlm> ntlm.NewClientContext failed.")
		return conn, err
	}
	defer secctx.Release()

	h := p.Headers
	h.Set("Proxy-Authorization", fmt.Sprintf("NTLM %s", base64.StdEncoding.EncodeToString(negotiate)))
	h.Set("Proxy-Connection", "Keep-Alive")
	connect := &http.Request{
		Method: "CONNECT",
		URL:    &url.URL{Opaque: addr},
		Host:   addr,
		Header: *h,
	}
	if err := connect.Write(conn); err != nil {
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
		return conn, errors.New("Unexpected HTTP status code")
	}

	challenge, err := base64.StdEncoding.DecodeString(resp.Header["Proxy-Authenticate"][0][5:])
	if err != nil {
		debugf("ntlm> Could not read challenge response")
		return conn, err
	}

	authenticate, err := secctx.Update(challenge)
	if err != nil {
		debugf("ntlm> Could not read authenticate")
		return conn, err
	}

	resp.Body.Close()
	h = p.Headers
	h.Set("Proxy-Authorization", fmt.Sprintf("NTLM %s", base64.StdEncoding.EncodeToString(authenticate)))
	h.Set("Proxy-Connection", "Keep-Alive")
	connect = &http.Request{
		Method: "CONNECT",
		URL:    &url.URL{Opaque: addr},
		Host:   addr,
		Header: *h,
	}
	if err := connect.Write(conn); err != nil {
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
