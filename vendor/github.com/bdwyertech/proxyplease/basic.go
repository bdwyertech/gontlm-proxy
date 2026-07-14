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

func dialBasic(p Proxy, addr string, conn net.Conn, br *bufio.Reader) (net.Conn, error) {
	debugf("basic> Attempting to authenticate")

	u := fmt.Sprintf("%s:%s", p.Username, p.Password)
	h := p.Headers.Clone()
	h.Set("Proxy-Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(u))))
	h.Set("Proxy-Connection", "Keep-Alive")
	connect := &http.Request{
		Method: "CONNECT",
		URL:    &url.URL{Opaque: addr},
		Host:   addr,
		Header: h,
	}
	if err := connect.WriteProxy(conn); err != nil {
		debugf("basic> Could not write authorization message to proxy: %s", err)
		return conn, err
	}
	resp, err := http.ReadResponse(br, connect)
	if err != nil {
		debugf("basic> Could not read response from proxy: %s", err)
		return conn, err
	}
	resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		// Successfully authorized with Basic
		debugf("basic> Successfully injected Basic to connection")
		return conn, nil
	}

	debugf("basic> Expected %d as return status, got: %d", http.StatusOK, resp.StatusCode)
	return conn, errors.New(http.StatusText(resp.StatusCode))
}
