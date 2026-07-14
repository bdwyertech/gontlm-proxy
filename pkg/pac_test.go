package ntlm_proxy

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/bdwyertech/proxyplease"
	"github.com/darren/gpac"
)

// --- Task 2: TestIsDialError ---

func TestIsDialError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "net.OpError (connection refused)",
			err:  &net.OpError{Op: "dial", Net: "tcp", Err: fmt.Errorf("connection refused")},
			want: true,
		},
		{
			name: "net.DNSError",
			err:  &net.DNSError{Err: "no such host", Name: "example.com"},
			want: true,
		},
		{
			name: "wrapped net.OpError",
			err:  fmt.Errorf("connect: %w", &net.OpError{Op: "dial", Net: "tcp", Err: fmt.Errorf("timeout")}),
			want: true,
		},
		{
			name: "generic error",
			err:  errors.New("something went wrong"),
			want: false,
		},
		{
			name: "HTTP-level error",
			err:  fmt.Errorf("proxy returned status 407"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isDialError(tt.err)
			if got != tt.want {
				t.Errorf("isDialError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

// --- Task 3: TestPacProxyToURL ---

func TestPacProxyToURL(t *testing.T) {
	tests := []struct {
		name       string
		proxy      *gpac.Proxy
		wantScheme string
		wantHost   string
	}{
		{
			name:       "PROXY type",
			proxy:      &gpac.Proxy{Type: "PROXY", Address: "10.0.0.1:8080"},
			wantScheme: "http",
			wantHost:   "10.0.0.1:8080",
		},
		{
			name:       "SOCKS type",
			proxy:      &gpac.Proxy{Type: "SOCKS", Address: "10.0.0.2:1080"},
			wantScheme: "socks5",
			wantHost:   "10.0.0.2:1080",
		},
		{
			name:       "SOCKS5 type",
			proxy:      &gpac.Proxy{Type: "SOCKS5", Address: "10.0.0.3:1080"},
			wantScheme: "socks5",
			wantHost:   "10.0.0.3:1080",
		},
		{
			name:       "unknown type lowercased",
			proxy:      &gpac.Proxy{Type: "HTTPS", Address: "10.0.0.4:443"},
			wantScheme: "https",
			wantHost:   "10.0.0.4:443",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pacProxyToURL(tt.proxy)
			if got.Scheme != tt.wantScheme {
				t.Errorf("pacProxyToURL(%s).Scheme = %q, want %q", tt.proxy.Type, got.Scheme, tt.wantScheme)
			}
			if got.Host != tt.wantHost {
				t.Errorf("pacProxyToURL(%s).Host = %q, want %q", tt.proxy.Type, got.Host, tt.wantHost)
			}
		})
	}
}

// --- Task 4: TestLoadPacFile ---

func TestLoadPacFile(t *testing.T) {
	t.Run("file:// scheme", func(t *testing.T) {
		dir := t.TempDir()
		pacPath := filepath.Join(dir, "test.pac")
		pacContent := `function FindProxyForURL(url, host) { return "PROXY 10.0.0.1:8080"; }`
		if err := os.WriteFile(pacPath, []byte(pacContent), 0644); err != nil {
			t.Fatal(err)
		}

		parser, err := loadPacFile("file://" + pacPath)
		if err != nil {
			t.Fatalf("loadPacFile(file://) error: %v", err)
		}
		if parser == nil {
			t.Fatal("loadPacFile(file://) returned nil parser")
		}

		result, err := parser.FindProxyForURL("http://example.com")
		if err != nil {
			t.Fatalf("FindProxyForURL error: %v", err)
		}
		if result != "PROXY 10.0.0.1:8080" {
			t.Errorf("FindProxyForURL = %q, want %q", result, "PROXY 10.0.0.1:8080")
		}
	})

	t.Run("http:// scheme", func(t *testing.T) {
		pacContent := `function FindProxyForURL(url, host) { return "PROXY 10.0.0.2:3128"; }`
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/x-ns-proxy-autoconfig")
			w.Write([]byte(pacContent))
		}))
		defer srv.Close()

		parser, err := loadPacFile(srv.URL)
		if err != nil {
			t.Fatalf("loadPacFile(http://) error: %v", err)
		}
		if parser == nil {
			t.Fatal("loadPacFile(http://) returned nil parser")
		}

		result, err := parser.FindProxyForURL("http://example.com")
		if err != nil {
			t.Fatalf("FindProxyForURL error: %v", err)
		}
		if result != "PROXY 10.0.0.2:3128" {
			t.Errorf("FindProxyForURL = %q, want %q", result, "PROXY 10.0.0.2:3128")
		}
	})

	t.Run("unsupported scheme", func(t *testing.T) {
		_, err := loadPacFile("ftp://example.com/proxy.pac")
		if err == nil {
			t.Fatal("loadPacFile(ftp://) should return error")
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := loadPacFile("file:///nonexistent/path/proxy.pac")
		if err == nil {
			t.Fatal("loadPacFile(nonexistent) should return error")
		}
	})
}

// --- Task 5: TestPacDialerFailover ---

func TestPacDialerFailover(t *testing.T) {
	// Start a real listener for the "working" proxy (third entry)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	workingAddr := listener.Addr().String()

	// Accept connections in background
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	// Set up a PAC parser that returns 3 entries: 2 unreachable + 1 working
	pacContent := fmt.Sprintf(
		`function FindProxyForURL(url, host) { return "PROXY 127.0.0.1:1; PROXY 127.0.0.1:2; PROXY %s"; }`,
		workingAddr,
	)
	dir := t.TempDir()
	pacPath := filepath.Join(dir, "failover.pac")
	if err := os.WriteFile(pacPath, []byte(pacContent), 0644); err != nil {
		t.Fatal(err)
	}

	parser, err := loadPacFile("file://" + pacPath)
	if err != nil {
		t.Fatal(err)
	}

	// Save/restore global pacParser
	oldParser := pacParser
	pacParser = parser
	defer func() { pacParser = oldParser }()

	directDial := func(ctx context.Context, network, addr string) (net.Conn, error) {
		return net.Dial(network, addr)
	}

	dial := pacDialer("http", "example.com:80", directDial, func(scheme, addr string, pxyURL *url.URL) proxyplease.DialContext {
		return func(ctx context.Context, network, dialAddr string) (net.Conn, error) {
			return net.Dial(network, pxyURL.Host)
		}
	})

	conn, err := dial(context.Background(), "tcp", "example.com:80")
	if err != nil {
		t.Fatalf("pacDialer should succeed via third proxy, got: %v", err)
	}
	if conn != nil {
		conn.Close()
	}
}

func TestPacDialerNoFailoverOnHTTPError(t *testing.T) {
	// PAC with 2 entries
	pacContent := `function FindProxyForURL(url, host) { return "PROXY 127.0.0.1:9998; PROXY 127.0.0.1:9999"; }`
	dir := t.TempDir()
	pacPath := filepath.Join(dir, "nofo.pac")
	if err := os.WriteFile(pacPath, []byte(pacContent), 0644); err != nil {
		t.Fatal(err)
	}

	parser, err := loadPacFile("file://" + pacPath)
	if err != nil {
		t.Fatal(err)
	}

	oldParser := pacParser
	pacParser = parser
	defer func() { pacParser = oldParser }()

	httpErr := fmt.Errorf("proxy returned status 407")
	callCount := 0

	directDial := func(ctx context.Context, network, addr string) (net.Conn, error) {
		return net.Dial(network, addr)
	}

	dial := pacDialer("http", "example.com:80", directDial, func(scheme, addr string, pxyURL *url.URL) proxyplease.DialContext {
		return func(ctx context.Context, network, dialAddr string) (net.Conn, error) {
			callCount++
			return nil, httpErr
		}
	})

	_, err = dial(context.Background(), "tcp", "example.com:80")
	if err == nil {
		t.Fatal("expected error from pacDialer")
	}
	if callCount != 1 {
		t.Errorf("expected proxyDial called once (no failover), got %d calls", callCount)
	}
}

func TestPacDialerDirectFallback(t *testing.T) {
	// Start a listener to act as "direct" destination
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	directAddr := listener.Addr().String()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	// PAC: unreachable proxy then DIRECT
	pacContent := `function FindProxyForURL(url, host) { return "PROXY 127.0.0.1:1; DIRECT"; }`
	dir := t.TempDir()
	pacPath := filepath.Join(dir, "direct.pac")
	if err := os.WriteFile(pacPath, []byte(pacContent), 0644); err != nil {
		t.Fatal(err)
	}

	parser, err := loadPacFile("file://" + pacPath)
	if err != nil {
		t.Fatal(err)
	}

	oldParser := pacParser
	pacParser = parser
	defer func() { pacParser = oldParser }()

	// Direct dialer connects to our listener
	directDial := func(ctx context.Context, network, addr string) (net.Conn, error) {
		return net.Dial(network, directAddr)
	}

	dial := pacDialer("https", directAddr, directDial, func(scheme, addr string, pxyURL *url.URL) proxyplease.DialContext {
		return func(ctx context.Context, network, dialAddr string) (net.Conn, error) {
			// Try to connect to the proxy address (which is unreachable port 1)
			return net.Dial(network, pxyURL.Host)
		}
	})

	conn, err := dial(context.Background(), "tcp", directAddr)
	if err != nil {
		t.Fatalf("pacDialer with DIRECT fallback should succeed, got: %v", err)
	}
	if conn != nil {
		conn.Close()
	}
}

// --- Task 7: TestPacDialerJSError ---

func TestPacDialerJSError(t *testing.T) {
	// PAC with invalid JS that will error on evaluation
	pacContent := `function FindProxyForURL(url, host) { throw new Error("JS error"); }`
	dir := t.TempDir()
	pacPath := filepath.Join(dir, "jserror.pac")
	if err := os.WriteFile(pacPath, []byte(pacContent), 0644); err != nil {
		t.Fatal(err)
	}

	parser, err := loadPacFile("file://" + pacPath)
	if err != nil {
		t.Fatal(err)
	}

	oldParser := pacParser
	pacParser = parser
	defer func() { pacParser = oldParser }()

	directCalled := false
	directDial := func(ctx context.Context, network, addr string) (net.Conn, error) {
		directCalled = true
		return nil, nil // Just testing that it's called
	}

	dial := pacDialer("http", "example.com:80", directDial, func(scheme, addr string, pxyURL *url.URL) proxyplease.DialContext {
		return func(ctx context.Context, network, dialAddr string) (net.Conn, error) {
			t.Fatal("proxyDial should not be called on JS error")
			return nil, nil
		}
	})

	// The pacDialer should return directDial on PAC eval error
	_, _ = dial(context.Background(), "tcp", "example.com:80")
	if !directCalled {
		t.Error("expected direct dialer to be called on PAC JS error")
	}
}
