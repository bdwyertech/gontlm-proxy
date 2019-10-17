package concord

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// default ports for http and https
var defaultPorts = map[string]string{
	"http":  "80",
	"https": "443",
}

// Handshaker is an interface representing a mechanism for
// handling proxy responses that request authentication (StatusCode 407)
type Handshaker interface {
	// Handshake executes a series of http transactions required for
	// proxy authentication.
	//
	// authResponse is a response from a proxy server that requires
	// authentication. It should be be used to determine the type of
	// authentication allowed by the proxy server (e.g. Basic, Negotiate,
	// or NTLM). Handshake is responsible for explicitly closing the body of
	// authResponse (if there is one) before writing to conn
	//
	// Handshake should apply any necessary proxy-authorization headers to the
	// http.Request, write request to conn and return a single response that is
	// the result of successful or failed authentication. In the cases of
	// successful authentication the response returned is often the desired
	// response for the request.
	Handshake(authResponse *http.Response, request *http.Request, conn net.Conn) (*http.Response, error)
}

// Transport is an implementation of RoundTripper that supports HTTP,
// HTTPS, and HTTP proxies (for either HTTP or HTTPS with CONNECT).
type Transport struct {
	// Proxy specifies a function to return a proxy for a given
	// Request. If the function returns a non-nil error, the
	// request is aborted with the provided error.
	//
	// The proxy type is determined by the URL scheme. "http"
	// is supported. If the scheme is empty, "http" is assumed.
	//
	// If Proxy is nil or returns a nil *URL, no proxy is used.
	Proxy func(*http.Request) (*url.URL, error)

	// ProxyAuthorizer specifies the mechanism by which proxies
	// that require authentication (Status code 407) are handled.
	// If nil, proxy authentication responses are returned as
	// they are received.
	ProxyAuthorizer Handshaker

	// TLSClientConfig specifies the TLS configuration to use with
	// tls.Client.
	// If nil, the default configuration is used.
	TLSClientConfig *tls.Config

	// TLSHandshakeTimeout specifies the maximum amount of time waiting to
	// wait for a TLS handshake. Zero means no timeout.
	TLSHandshakeTimeout time.Duration

	setupDefaultsOnce sync.Once
}

// RoundTrip implements the RoundTripper interface.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.setupDefaultsOnce.Do(t.setupDefaults)
	if err := checkRequest(req); err != nil {
		return nil, err
	}
	dialAddr := canonicalAddress(req.URL)
	var (
		conn  net.Conn
		proxy *url.URL
		err   error
	)
	if t.Proxy != nil {
		proxy, err = t.Proxy(req)
		if err != nil {
			return nil, err
		}
		if proxy != nil {
			dialAddr = canonicalAddress(proxy)
		}
	}

	if proxy != nil {
		conn, err = net.Dial("tcp", dialAddr)
	} else {
		switch req.URL.Scheme {
		case "http":
			conn, err = net.Dial("tcp", dialAddr)
		case "https":
			conn, err = tls.Dial("tcp", dialAddr, t.TLSClientConfig)
		}
	}
	if err != nil {
		return nil, err
	}

	if proxy != nil {
		switch req.URL.Scheme {
		case "http":
			if err := req.WriteProxy(conn); err != nil {
				conn.Close()
				return nil, err
			}
			res, err := http.ReadResponse(bufio.NewReader(conn), req)
			if err != nil {
				conn.Close()
				return nil, err
			}
			if res.StatusCode == http.StatusProxyAuthRequired && t.ProxyAuthorizer != nil {
				// Rewind the request body, the handshaker needs it.
				if req.GetBody != nil {
					if req.Body, err = req.GetBody(); err != nil {
						conn.Close()
						return nil, err
					}
				}
				res, err = t.ProxyAuthorizer.Handshake(res, req, conn)
				if err != nil {
					conn.Close()
					return nil, err
				}
			}
			return wrapConnBody(conn, res)
		case "https":
			targetAddr := canonicalAddress(req.URL)
			hdr := make(http.Header)
			hdr.Set("Proxy-Connection", "Keep-Alive")
			connectReq := &http.Request{
				Method: "CONNECT",
				URL:    &url.URL{Opaque: targetAddr},
				Host:   targetAddr,
				Header: hdr,
			}
			if err := connectReq.Write(conn); err != nil {
				conn.Close()
				return nil, err
			}
			connectRes, err := http.ReadResponse(bufio.NewReader(conn), req)
			if err != nil {
				conn.Close()
				return nil, err
			}
			if connectRes.StatusCode == http.StatusProxyAuthRequired && t.ProxyAuthorizer != nil {
				connectRes, err = t.ProxyAuthorizer.Handshake(connectRes, connectReq, conn)
				if err != nil {
					conn.Close()
					return nil, err
				}
			}
			if connectRes.StatusCode == http.StatusOK {
				if err := connectRes.Body.Close(); err != nil {
					conn.Close()
					return nil, err
				}
				tlsConfig := t.TLSClientConfig.Clone()
				if tlsConfig.ServerName == "" {
					tlsConfig.ServerName = req.Host
				}
				tlsConn, err := t.addTLS(conn, tlsConfig)
				if err != nil {
					conn.Close()
					return nil, err
				}
				if err := req.Write(tlsConn); err != nil {
					tlsConn.Close()
					return nil, err
				}
				res, err := http.ReadResponse(bufio.NewReader(tlsConn), req)
				if err != nil {
					tlsConn.Close()
					return nil, err
				}
				return wrapConnBody(tlsConn, res)
			}
			return connectRes, nil
		}
	}

	if err := req.Write(conn); err != nil {
		conn.Close()
		return nil, err
	}
	res, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return wrapConnBody(conn, res)
}

// Apply default configurations if none are supplied
func (t *Transport) setupDefaults() {
	if t.TLSClientConfig == nil {
		t.TLSClientConfig = &tls.Config{}
	}
}

// Negotiate a TLS session
func (t *Transport) addTLS(conn net.Conn, cfg *tls.Config) (*tls.Conn, error) {
	tlsConn := tls.Client(conn, cfg)
	errc := make(chan error, 2)
	var timer *time.Timer
	if d := t.TLSHandshakeTimeout; d != 0 {
		timer = time.AfterFunc(d, func() {
			errc <- fmt.Errorf("tls handshake timeout")
		})
	}
	go func() {
		err := tlsConn.Handshake()
		if timer != nil {
			timer.Stop()
		}
		errc <- err
	}()
	if err := <-errc; err != nil {
		conn.Close()
		return nil, err
	}
	if !cfg.InsecureSkipVerify {
		if err := tlsConn.VerifyHostname(cfg.ServerName); err != nil {
			conn.Close()
			return nil, err
		}
	}
	return tlsConn, nil
}

// Returns nil if the request is properly formatted
func checkRequest(req *http.Request) error {
	return nil
}

// Returns an address in the form host:port, adding the default
// port value when necessary.
func canonicalAddress(url *url.URL) string {
	host := url.Hostname()
	port := url.Port()
	if port == "" {
		port = defaultPorts[url.Scheme]
	}
	return fmt.Sprintf("%s:%s", host, port)
}

type connBodyReadCloser struct {
	io.ReadCloser
	conn net.Conn
}

func (b *connBodyReadCloser) Close() error {
	b.conn.Close()
	return b.ReadCloser.Close()
}

// wrap the conn with the body so that when the body is closed the conn is closed
func wrapConnBody(conn net.Conn, res *http.Response) (*http.Response, error) {
	res.Body = &connBodyReadCloser{
		ReadCloser: res.Body,
		conn:       conn,
	}
	return res, nil
}
