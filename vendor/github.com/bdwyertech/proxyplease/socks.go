package proxyplease

import (
	"errors"
	"net"
	"net/url"

	"golang.org/x/net/proxy"
	hsocks "h12.io/socks"
)

func dialAndNegotiateSOCKS(u *url.URL, user, pass, addr string) (net.Conn, error) {
	debugf("socks> using socks proxy")
	switch u.Scheme {
	case "socks4", "socks4a":
		debugf("socks> connecting via %s", u.Scheme)
		socks4Dial := hsocks.Dial(u.String())
		conn, err := socks4Dial("tcp", addr)
		if err != nil {
			debugf("socks> Could not call dial socks4 context with proxy: %s", err)
			return conn, err
		}
		return conn, err
	case "socks5", "socks5h", "socks":
		debugf("socks> connecting via %s", u.Scheme)
		// use golang.org/x/net/proxy SOCKS5 implementation for authentication support
		auth := &proxy.Auth{User: user, Password: pass}
		sp, _ := proxy.SOCKS5("tcp", u.Host, auth, proxy.Direct)
		conn, err := sp.Dial("tcp", addr)
		return conn, err
	}
	debugf("socks> Unsupported socks scheme: %s", u.Scheme)
	return nil, errors.New("Unsupported socks URL scheme")
}
