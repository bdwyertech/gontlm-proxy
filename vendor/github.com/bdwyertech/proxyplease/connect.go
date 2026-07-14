package proxyplease

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
)

func dialAndNegotiateHTTP(p Proxy, addr string, baseDial func() (net.Conn, error)) (net.Conn, error) {
	// establish TCP with proxy. baseDial will negotiate TLS if needed.
	conn, err := baseDial()
	if err != nil {
		debugf("connect> Could not call dial context with proxy: %s", err)
		return nil, err
	}

	// Single buffered reader for this connection — threaded into sub-dialers
	// so that any bytes buffered past the response boundary are not lost.
	br := bufio.NewReader(conn)

	// build and write first CONNECT request (probe)
	h := p.Headers.Clone()
	h.Set("Proxy-Connection", "Keep-Alive")
	connect := &http.Request{
		Method: "CONNECT",
		URL:    &url.URL{Opaque: addr},
		Host:   addr,
		Header: h,
	}
	if err := connect.WriteProxy(conn); err != nil {
		debugf("connect> CONNECT to proxy failed: %s", err)
		conn.Close()
		return nil, err
	}

	// read first response
	resp, err := http.ReadResponse(br, connect)
	if err != nil {
		debugf("connect> Could not read response from proxy: %s", err)
		conn.Close()
		return nil, err
	}
	resp.Body.Close()

	// if StatusOK, no auth is required and proxy is established
	if resp.StatusCode == http.StatusOK {
		debugf("connect> Proxy successfully established. No authentication was required.")
		return conn, nil
	}

	// if authentication is required
	if resp.StatusCode == http.StatusProxyAuthRequired {
		debugf("connect> Proxy authentication is required. Attempting to select a authentication scheme.")

		// read authentication scheme options
		schemes := resp.Header["Proxy-Authenticate"]
		var lastErr error
		for _, s := range schemes {
			// only test for first word in scheme
			trimmed := strings.Split(s, " ")[0]
			switch trimmed {
			case "NTLM":
				if !contains(p.AuthSchemeFilter, "NTLM") {
					debugf("connect> Skipping NTLM due to AuthSchemeFilter")
					continue
				}
				c, err := dialNTLM(p, addr, conn, br)
				if err != nil {
					debugf("connect> NTLM authentication failed. Trying next available scheme.")
					lastErr = err
					// Clean fallthrough: close failed conn, dial fresh for next scheme
					conn.Close()
					conn, err = baseDial()
					if err != nil {
						debugf("connect> Could not re-dial for next scheme: %s", err)
						return nil, err
					}
					br = bufio.NewReader(conn)
					continue
				}
				return c, nil
			case "Basic", "BASIC":
				if !contains(p.AuthSchemeFilter, "Basic") {
					debugf("connect> Skipping Basic due to AuthSchemeFilter")
					continue
				}
				c, err := dialBasic(p, addr, conn, br)
				if err != nil {
					debugf("connect> Basic authentication failed. Trying next available scheme.")
					lastErr = err
					// Clean fallthrough: close failed conn, dial fresh for next scheme
					conn.Close()
					conn, err = baseDial()
					if err != nil {
						debugf("connect> Could not re-dial for next scheme: %s", err)
						return nil, err
					}
					br = bufio.NewReader(conn)
					continue
				}
				return c, nil

			case "Negotiate", "NEGOTIATE":
				if !contains(p.AuthSchemeFilter, "Negotiate") {
					debugf("connect> Skipping Negotiate due to AuthSchemeFilter")
					continue
				}
				c, err := dialNegotiate(p, addr, conn, br)
				if err != nil {
					debugf("connect> Negotiate authentication failed. Trying next available scheme.")
					lastErr = err
					// Clean fallthrough: close failed conn, dial fresh for next scheme
					conn.Close()
					conn, err = baseDial()
					if err != nil {
						debugf("connect> Could not re-dial for next scheme: %s", err)
						return nil, err
					}
					br = bufio.NewReader(conn)
					continue
				}
				return c, nil

			case "Kerberos":
				debugf("connect> Kerberos not implemented yet. Trying next available scheme.")
				continue

			case "Digest":
				debugf("connect> Digest not implemented yet. Trying next available scheme.")
				continue

			default:
				debugf("connect> Unsupported proxy authentication scheme: '%s'. Trying next available scheme.", trimmed)
				continue
			}
		}

		debugf("connect> No proxy authentication completed successfully")
		conn.Close()
		if lastErr != nil {
			return nil, lastErr
		}
		return nil, errors.New("no proxy authentication scheme succeeded")
	}

	debugf("connect> Unhandled HTTP status, got: %d", resp.StatusCode)
	conn.Close()
	return nil, errors.New(http.StatusText(resp.StatusCode))
}

func contains(s []string, e string) bool {
	// if no filter supplied, assume scheme is wanted
	if s == nil {
		return true
	}
	// otherwise, test if filter matches
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
