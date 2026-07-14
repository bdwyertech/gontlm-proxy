package ntlm_proxy

import (
	"errors"
	"io"
	"net"
	"sync"
	"syscall"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

// captureHook is a logrus hook that records emitted entries for assertions.
type captureHook struct {
	mu      sync.Mutex
	entries []*log.Entry
}

func (h *captureHook) Levels() []log.Level { return log.AllLevels }

func (h *captureHook) Fire(e *log.Entry) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.entries = append(h.entries, e)
	return nil
}

func (h *captureHook) count() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.entries)
}

// countMsg returns how many captured entries have the given message.
func (h *captureHook) countMsg(msg string) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	n := 0
	for _, e := range h.entries {
		if e.Message == msg {
			n++
		}
	}
	return n
}

// newCaptureLogger returns a *log.Entry backed by an isolated logger whose
// output is captured by the returned hook (no writes to os.Stderr).
func newCaptureLogger() (*log.Entry, *captureHook) {
	logger := log.New()
	logger.SetOutput(io.Discard)
	logger.SetLevel(log.DebugLevel)
	hook := &captureHook{}
	logger.AddHook(hook)
	return log.NewEntry(logger), hook
}

// timeoutError is a synthetic net error that reports Timeout() == true, used to
// exercise the upstream_timeout branch of classifyConnError.
type timeoutError struct{}

func (timeoutError) Error() string   { return "i/o timeout" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return true }

// TestParseDurationEnv verifies GONTLM_* duration parsing semantics:
// a valid duration parses, an empty/unset env returns the default, and an
// invalid value falls back to the default (design §3, REQ-TUN-005).
func TestParseDurationEnv(t *testing.T) {
	const key = "GONTLM_TEST_DURATION"

	t.Run("valid", func(t *testing.T) {
		t.Setenv(key, "5s")
		if got := parseDurationEnv(key, time.Second); got != 5*time.Second {
			t.Fatalf("parseDurationEnv(valid) = %v, want %v", got, 5*time.Second)
		}
	})

	t.Run("empty returns default", func(t *testing.T) {
		// t.Setenv registers cleanup; explicitly clear to simulate unset.
		t.Setenv(key, "")
		def := 3 * time.Second
		if got := parseDurationEnv(key, def); got != def {
			t.Fatalf("parseDurationEnv(empty) = %v, want default %v", got, def)
		}
	})

	t.Run("invalid returns default", func(t *testing.T) {
		t.Setenv(key, "nope")
		def := 7 * time.Second
		if got := parseDurationEnv(key, def); got != def {
			t.Fatalf("parseDurationEnv(invalid) = %v, want default %v", got, def)
		}
	})

	t.Run("zero default preserved", func(t *testing.T) {
		t.Setenv(key, "")
		if got := parseDurationEnv(key, 0); got != 0 {
			t.Fatalf("parseDurationEnv(empty, 0) = %v, want 0", got)
		}
	})
}

// TestEnableTCPKeepAlive_TCPConn verifies keepalive is enabled on a real
// loopback TCP connection without error (design §2.1, REQ-TUN-001).
func TestEnableTCPKeepAlive_TCPConn(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	dialed, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer dialed.Close()

	accepted, err := ln.Accept()
	if err != nil {
		t.Fatalf("accept: %v", err)
	}
	defer accepted.Close()

	if _, ok := accepted.(*net.TCPConn); !ok {
		t.Fatalf("expected *net.TCPConn, got %T", accepted)
	}

	// Should apply keepalive without taking a warning/error path.
	enableTCPKeepAlive(accepted, 5*time.Second)

	// underlyingTCPConn on a plain TCPConn returns the conn itself.
	if underlyingTCPConn(accepted) != net.Conn(accepted) {
		t.Fatalf("underlyingTCPConn should return the TCPConn unchanged")
	}
}

// TestEnableTCPKeepAlive_Pipe verifies that a conn lacking keepalive support
// (net.Pipe) is handled as a graceful no-op with no panic (design §2.1).
func TestEnableTCPKeepAlive_Pipe(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	// Must not panic and must be a no-op.
	enableTCPKeepAlive(c1, 5*time.Second)

	if underlyingTCPConn(c1) != c1 {
		t.Fatalf("underlyingTCPConn should return the pipe conn unchanged")
	}
}

// TestClassifyConnError verifies the post-dial transfer-error classification
// table (design §2.3, REQ-TUN-003). This is independent of the dial-time
// isDialError helper in pac.go.
func TestClassifyConnError(t *testing.T) {
	opErr := func(op string, inner error) error {
		return &net.OpError{Op: op, Net: "tcp", Err: inner}
	}

	tests := []struct {
		name string
		op   string
		err  error
		want connErrorKind
	}{
		{"eof read", "read", io.EOF, errKindCleanEOF},
		{"eof write", "write", io.EOF, errKindCleanEOF},
		{"timeout read", "read", opErr("read", timeoutError{}), errKindUpstreamTimeout},
		{"timeout write", "write", opErr("write", timeoutError{}), errKindUpstreamTimeout},
		{"reset read", "read", opErr("read", syscall.ECONNRESET), errKindUpstreamReset},
		{"reset write", "write", opErr("write", syscall.ECONNRESET), errKindUpstreamReset},
		{"epipe write", "write", opErr("write", syscall.EPIPE), errKindClientGone},
		{"epipe read", "read", opErr("read", syscall.EPIPE), errKindUnknown},
		{"generic", "read", errors.New("boom"), errKindUnknown},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := classifyConnError(tc.op, tc.err); got != tc.want {
				t.Fatalf("classifyConnError(%q, %v) = %q, want %q", tc.op, tc.err, got, tc.want)
			}
		})
	}
}

// TestInstrumentedConn_Metrics verifies Read/Write update byte counters, chunk
// counts, and last-activity timestamps over a net.Pipe (design §2.2, REQ-TUN-002).
func TestInstrumentedConn_Metrics(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	logger, _ := newCaptureLogger()
	ic := newInstrumentedConn(client, "example.com:443", logger)

	payload := []byte("hello world")

	// Server echoes exactly one payload back.
	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, len(payload))
		if _, err := io.ReadFull(server, buf); err != nil {
			return
		}
		_, _ = server.Write(buf)
	}()

	if n, err := ic.Write(payload); err != nil || n != len(payload) {
		t.Fatalf("Write = (%d, %v), want (%d, nil)", n, err, len(payload))
	}

	readBuf := make([]byte, len(payload))
	if n, err := io.ReadFull(ic, readBuf); err != nil || n != len(payload) {
		t.Fatalf("Read = (%d, %v), want (%d, nil)", n, err, len(payload))
	}
	<-done

	if got := ic.bytesWritten.Load(); got != int64(len(payload)) {
		t.Fatalf("bytesWritten = %d, want %d", got, len(payload))
	}
	if got := ic.bytesRead.Load(); got != int64(len(payload)) {
		t.Fatalf("bytesRead = %d, want %d", got, len(payload))
	}
	if got := ic.chunksRead.Load(); got < 1 {
		t.Fatalf("chunksRead = %d, want >= 1", got)
	}
	if ic.lastWrite.Load() == 0 {
		t.Fatalf("lastWrite timestamp was not set")
	}
	if ic.lastRead.Load() == 0 {
		t.Fatalf("lastRead timestamp was not set")
	}
}

// TestInstrumentedConn_Close verifies Close sets the closed flag, delegates to
// the underlying conn, and emits a "Tunnel closed" log entry (design §2.2).
func TestInstrumentedConn_Close(t *testing.T) {
	client, server := net.Pipe()
	defer server.Close()

	logger, hook := newCaptureLogger()
	ic := newInstrumentedConn(client, "example.com:443", logger)

	if err := ic.Close(); err != nil {
		t.Fatalf("Close = %v, want nil", err)
	}
	if !ic.closed.Load() {
		t.Fatalf("closed flag not set after Close")
	}
	if hook.count() != 1 {
		t.Fatalf("expected exactly 1 log entry, got %d", hook.count())
	}

	// Underlying conn should be closed: a subsequent Read errors.
	_ = client.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
	if _, err := client.Read(make([]byte, 1)); err == nil {
		t.Fatalf("underlying conn was not closed by Close")
	}
}

const idleWarnMsg = "Tunnel idle — at risk of upstream proxy timeout"

// TestIdleWatchdog_Fires verifies the watchdog logs one warning after the
// tunnel stays idle past the threshold (design §2.4, REQ-TUN-004).
func TestIdleWatchdog_Fires(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	logger, hook := newCaptureLogger()
	ic := newInstrumentedConn(client, "example.com:443", logger)

	// Seed activity to now-threshold in the past so the first tick sees idle.
	ic.startIdleWatchdog(50 * time.Millisecond)

	deadline := time.After(2 * time.Second)
	for hook.countMsg(idleWarnMsg) < 1 {
		select {
		case <-deadline:
			t.Fatalf("watchdog did not fire within timeout")
		case <-time.After(10 * time.Millisecond):
		}
	}
	ic.Close()
}

// TestIdleWatchdog_EdgeTriggered verifies the watchdog logs at most once while
// the tunnel remains continuously idle across multiple ticks (design §2.4).
func TestIdleWatchdog_EdgeTriggered(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	logger, hook := newCaptureLogger()
	ic := newInstrumentedConn(client, "example.com:443", logger)

	ic.startIdleWatchdog(30 * time.Millisecond)

	// Remain idle across several tick intervals.
	time.Sleep(200 * time.Millisecond)
	ic.Close()

	if got := hook.countMsg(idleWarnMsg); got != 1 {
		t.Fatalf("expected exactly 1 idle warning, got %d", got)
	}
}

// TestIdleWatchdog_ReArms verifies that after a warning fires, activity clears
// the edge-trigger and a subsequent idle period produces a second warning
// (design §2.4).
func TestIdleWatchdog_ReArms(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	logger, hook := newCaptureLogger()
	ic := newInstrumentedConn(client, "example.com:443", logger)

	// Drain whatever the test writes so Write does not block.
	go func() {
		buf := make([]byte, 64)
		for {
			if _, err := server.Read(buf); err != nil {
				return
			}
		}
	}()

	ic.startIdleWatchdog(30 * time.Millisecond)

	// Wait for the first warning.
	waitFor := func(n int) {
		deadline := time.After(2 * time.Second)
		for hook.countMsg(idleWarnMsg) < n {
			select {
			case <-deadline:
				t.Fatalf("expected >= %d idle warnings, got %d", n, hook.countMsg(idleWarnMsg))
			case <-time.After(10 * time.Millisecond):
			}
		}
	}
	waitFor(1)

	// Activity re-arms the edge-trigger.
	if _, err := ic.Write([]byte("ping")); err != nil {
		t.Fatalf("Write: %v", err)
	}

	// A second idle period should produce a second warning.
	waitFor(2)
	ic.Close()
}
