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
	"strings"

	"github.com/alexbrainman/sspi"
	"github.com/alexbrainman/sspi/negotiate"
)

func dialNegotiate(p Proxy, addr string, baseDial func() (net.Conn, error)) (net.Conn, error) {
	debugf("negotiate> Attempting to authenticate")

	conn, err := baseDial()
	if err != nil {
		debugf("negotiate> Could not call dial context with proxy: %s", err)
		return conn, err
	}

	h, err := canonicalizeHostname(p.URL.Hostname())
	if err != nil {
		debugf("negotiate> Error canonicalizing hostname: %s", err)
		return conn, err
	}
	spn := "HTTP/" + h

	var cred *sspi.Credentials
	if p.Domain != "" && p.Username != "" && p.Password != "" {
		cred, err = negotiate.AcquireUserCredentials(p.Domain, p.Username, p.Password)
	} else {
		cred, err = negotiate.AcquireCurrentUserCredentials()
	}
	if err != nil {
		return conn, err
	}
	defer cred.Release()

	secctx, token, err := negotiate.NewClientContext(cred, spn)
	if err != nil {
		return conn, err
	}
	defer secctx.Release()

	head := p.Headers
	head.Set("Proxy-Authorization", fmt.Sprintf("Negotiate %s", base64.StdEncoding.EncodeToString(token)))
	head.Set("Proxy-Connection", "Keep-Alive")
	connect := &http.Request{
		Method: "CONNECT",
		URL:    &url.URL{Opaque: addr},
		Host:   addr,
		Header: *head,
	}
	if err := connect.Write(conn); err != nil {
		debugf("negotiate> Could not write token message to proxy: %s", err)
		return conn, err
	}
	br := bufio.NewReader(conn)
	resp, err := http.ReadResponse(br, connect)
	if err != nil {
		debugf("negotiate> Could not read token response from proxy: %s", err)
		return conn, err
	}

	if resp.StatusCode != http.StatusOK {
		debugf("negotiate> Expected %d as return status, got: %d", http.StatusOK, resp.StatusCode)
		return conn, errors.New(http.StatusText(resp.StatusCode))
	}

	resp.Body.Close()

	debugf("negotiate> Successfully injected Negotiate::Kerberos to connection")
	return conn, nil
}

func canonicalizeHostname(hostname string) (string, error) {
	addrs, err := net.LookupHost(hostname)
	if err != nil {
		return "", err
	}
	if len(addrs) < 1 {
		return hostname, nil
	}

	names, err := net.LookupAddr(addrs[0])
	if err != nil {
		return "", err
	}
	if len(names) < 1 {
		return hostname, nil
	}

	return strings.TrimRight(names[0], "."), nil
}
