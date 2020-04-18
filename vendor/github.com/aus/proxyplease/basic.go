package proxyplease

import (
	"bufio"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
)

func dialBasic(p Proxy, addr string, baseDial func() (net.Conn, error)) (net.Conn, error) {
	debugf("basic> Attempting to authenticate")

	conn, err := baseDial()
	if err != nil {
		debugf("basic> Could not call dial context with proxy: %s", err)
		return conn, err
	}

	u := fmt.Sprintf("%s:%s", p.Username, p.Password)
	h := p.Headers
	h.Set("Proxy-Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(u))))
	h.Set("Proxy-Connection", "Keep-Alive")
	connect := &http.Request{
		Method: "CONNECT",
		URL:    &url.URL{Opaque: addr},
		Host:   addr,
		Header: *h,
	}
	if err := connect.Write(conn); err != nil {
		debugf("basic> Could not write authorization message to proxy: %s", err)
		return conn, err
	}
	br := bufio.NewReader(conn)
	resp, err := http.ReadResponse(br, connect)
	if err != nil {
		debugf("basic> Could not read response from proxy: %s", err)
		return conn, err
	}

	if resp.StatusCode == http.StatusOK {
		// Succussfully authorized with Basic
		debugf("basic> Successfully injected Basic to connection")
		return conn, nil
	}

	debugf("basic> Expected %d as return status, got: %d", http.StatusOK, resp.StatusCode)
	return conn, errors.New(http.StatusText(resp.StatusCode))
}
