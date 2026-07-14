package ntlm_proxy

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/bdwyertech/proxyplease"
	"github.com/darren/gpac"
)

// pacParser holds the compiled PAC file, initialized once at startup.
var pacParser *gpac.Parser

// loadPacFile loads and compiles a PAC file from the given URL.
// Supports file:// for local filesystem and http(s):// for remote fetches.
func loadPacFile(pacURL string) (*gpac.Parser, error) {
	if afterPath, ok := strings.CutPrefix(pacURL, "file://"); ok {
		return gpac.FromFile(afterPath)
	}
	if strings.HasPrefix(pacURL, "http://") || strings.HasPrefix(pacURL, "https://") {
		return gpac.FromURL(pacURL)
	}
	return nil, fmt.Errorf("unsupported PAC URL scheme: %s", pacURL)
}

// isDialError returns true if the error represents a connection-level failure
// that should trigger failover to the next proxy entry.
func isDialError(err error) bool {
	if err == nil {
		return false
	}
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		return true
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return true
	}
	return false
}

// pacProxyToURL converts a parsed PAC proxy entry to a standard *url.URL.
func pacProxyToURL(pxy *gpac.Proxy) *url.URL {
	switch strings.ToUpper(pxy.Type) {
	case "PROXY":
		return &url.URL{Scheme: "http", Host: pxy.Address}
	case "SOCKS", "SOCKS5":
		return &url.URL{Scheme: "socks5", Host: pxy.Address}
	default:
		return &url.URL{Scheme: strings.ToLower(pxy.Type), Host: pxy.Address}
	}
}

// pacDialer evaluates the PAC file for the given scheme+addr and returns a
// DialContext that iterates the proxy list with failover on dial errors.
func pacDialer(scheme, addr string, directDial proxyplease.DialContext, proxyDial func(string, string, *url.URL) proxyplease.DialContext) proxyplease.DialContext {
	targetURL := scheme + "://" + addr
	result, err := pacParser.FindProxyForURL(targetURL)
	if err != nil {
		log.Warnf("PAC evaluation error for %s: %v — falling back to DIRECT", targetURL, err)
		return directDial
	}

	proxies := gpac.ParseProxy(result)
	if len(proxies) == 0 {
		log.Warnf("PAC returned empty proxy list for %s — falling back to DIRECT", targetURL)
		return directDial
	}

	return func(ctx context.Context, network, dialAddr string) (net.Conn, error) {
		var lastErr error
		for i, pxy := range proxies {
			var dial proxyplease.DialContext
			if strings.ToUpper(pxy.Type) == "DIRECT" {
				dial = directDial
			} else {
				pxyURL := pacProxyToURL(pxy)
				dial = proxyDial(scheme, dialAddr, pxyURL)
			}

			conn, dialErr := dial(ctx, network, dialAddr)
			if dialErr == nil {
				return conn, nil
			}

			lastErr = dialErr
			if !isDialError(dialErr) {
				// Non-dial error: do not failover
				return nil, dialErr
			}

			log.Warnf("PAC failover [%d/%d]: %s %s failed: %v",
				i+1, len(proxies), pxy.Type, pxy.Address, dialErr)
		}
		return nil, fmt.Errorf("all PAC proxies exhausted for %s: %w", addr, lastErr)
	}
}
