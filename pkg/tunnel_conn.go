package ntlm_proxy

import (
	"errors"
	"io"
	"net"
	"sync/atomic"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

// Package-level tunnel liveness configuration, populated from the environment
// in init() (design §3, REQ-TUN-005). init() runs before Run(), so these values
// are set before the dialer block wires up wrapTunnelConn.
var (
	// keepAlivePeriod is the TCP keepalive probe interval applied to upstream
	// tunnel sockets. Zero disables keepalive tuning. (GONTLM_TCP_KEEPALIVE)
	keepAlivePeriod time.Duration

	// idleWarnThreshold is the no-activity duration after which the idle
	// watchdog logs a warning. Zero disables the watchdog. (GONTLM_TUNNEL_IDLE_WARN)
	idleWarnThreshold time.Duration
)

func init() {
	keepAlivePeriod = parseDurationEnv("GONTLM_TCP_KEEPALIVE", 5*time.Second)
	idleWarnThreshold = parseDurationEnv("GONTLM_TUNNEL_IDLE_WARN", 0)
}

// underlyingTCPConn returns the raw transport conn beneath any wrapper that
// exposes NetConn() (e.g. *tls.Conn since Go 1.18). If no such wrapper is
// present, the conn is returned unchanged (design §2.1).
func underlyingTCPConn(conn net.Conn) net.Conn {
	type netConner interface{ NetConn() net.Conn }
	if nc, ok := conn.(netConner); ok {
		return nc.NetConn()
	}
	return conn
}

// enableTCPKeepAlive turns on TCP keepalive with the given probe period on the
// underlying socket. Conns that do not support keepalive (e.g. net.Pipe in
// tests) are gracefully skipped (design §2.1, REQ-TUN-001).
func enableTCPKeepAlive(conn net.Conn, period time.Duration) {
	underlying := underlyingTCPConn(conn)
	if tc, ok := underlying.(interface {
		SetKeepAlive(bool) error
		SetKeepAlivePeriod(time.Duration) error
	}); ok {
		if err := tc.SetKeepAlive(true); err != nil {
			log.Warnf("Failed to enable TCP keepalive: %v", err)
			return
		}
		if err := tc.SetKeepAlivePeriod(period); err != nil {
			log.Warnf("Failed to set TCP keepalive period: %v", err)
		}
	}
}

// connErrorKind classifies post-dial transfer errors on the upstream tunnel
// socket for logging (design §2.3, REQ-TUN-003). It is distinct from the
// dial-time isDialError helper in pac.go and never influences control flow.
type connErrorKind string

const (
	errKindUpstreamReset   connErrorKind = "upstream_reset"
	errKindUpstreamTimeout connErrorKind = "upstream_timeout"
	errKindClientGone      connErrorKind = "client_gone"
	errKindCleanEOF        connErrorKind = "clean_eof"
	errKindUnknown         connErrorKind = "unknown"
)

// classifyConnError maps a transfer error observed during op ("read"/"write")
// on the upstream tunnel socket to a connErrorKind per the §2.3 table.
func classifyConnError(op string, err error) connErrorKind {
	if err == io.EOF {
		return errKindCleanEOF
	}
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return errKindUpstreamTimeout
		}
		if errors.Is(netErr.Err, syscall.ECONNRESET) {
			return errKindUpstreamReset
		}
		if op == "write" && errors.Is(netErr.Err, syscall.EPIPE) {
			return errKindClientGone
		}
	}
	return errKindUnknown
}

// instrumentedConn wraps an upstream tunnel net.Conn to record transfer metrics
// and classify/log transport errors. All mutable state uses sync/atomic so the
// two goproxy copy goroutines (Read side and Write side) share it lock-free
// (design §2.2, REQ-TUN-002, REQ-TUN-006).
type instrumentedConn struct {
	net.Conn
	target       string // remote host:port for logging
	direction    string // "upstream"
	created      time.Time
	lastRead     atomic.Int64 // unix nanos of last successful Read
	lastWrite    atomic.Int64 // unix nanos of last successful Write
	bytesRead    atomic.Int64
	bytesWritten atomic.Int64
	chunksRead   atomic.Int64 // number of Read calls that returned data
	closed       atomic.Bool
	idleWarned   atomic.Bool // edge-trigger: true while in warned-idle state
	logger       *log.Entry
}

// newInstrumentedConn builds an instrumentedConn around an upstream conn,
// seeding the created/last-activity timestamps to now (design §2.2, §2.5).
func newInstrumentedConn(conn net.Conn, target string, logger *log.Entry) *instrumentedConn {
	now := time.Now()
	ic := &instrumentedConn{
		Conn:      conn,
		target:    target,
		direction: "upstream",
		created:   now,
		logger:    logger.WithField("target", target),
	}
	ic.lastRead.Store(now.UnixNano())
	ic.lastWrite.Store(now.UnixNano())
	return ic
}

func (ic *instrumentedConn) Read(b []byte) (int, error) {
	n, err := ic.Conn.Read(b)
	if n > 0 {
		ic.lastRead.Store(time.Now().UnixNano())
		ic.bytesRead.Add(int64(n))
		ic.chunksRead.Add(1)
		ic.idleWarned.Store(false) // reset edge-trigger on activity
	}
	if err != nil && !ic.closed.Load() {
		ic.logConnectionError("read", err)
	}
	return n, err
}

func (ic *instrumentedConn) Write(b []byte) (int, error) {
	n, err := ic.Conn.Write(b)
	if n > 0 {
		ic.lastWrite.Store(time.Now().UnixNano())
		ic.bytesWritten.Add(int64(n))
		ic.idleWarned.Store(false) // reset edge-trigger on activity
	}
	if err != nil && !ic.closed.Load() {
		ic.logConnectionError("write", err)
	}
	return n, err
}

func (ic *instrumentedConn) Close() error {
	ic.closed.Store(true)
	age := time.Since(ic.created)
	ic.logger.WithFields(log.Fields{
		"age":           age.Round(time.Millisecond),
		"bytes_read":    ic.bytesRead.Load(),
		"bytes_written": ic.bytesWritten.Load(),
		"chunks_read":   ic.chunksRead.Load(),
	}).Debug("Tunnel closed")
	return ic.Conn.Close()
}

// logConnectionError emits a structured warning enriched with the classified
// error kind, activity metrics, and the raw error (design §2.3). It never
// influences control flow.
func (ic *instrumentedConn) logConnectionError(op string, err error) {
	kind := classifyConnError(op, err)
	lastActivity := max(ic.lastRead.Load(), ic.lastWrite.Load())
	idle := time.Since(time.Unix(0, lastActivity))

	entry := ic.logger.WithFields(log.Fields{
		"kind":          string(kind),
		"op":            op,
		"target":        ic.target,
		"age":           time.Since(ic.created).Round(time.Millisecond),
		"idle_since":    idle.Round(time.Millisecond),
		"bytes_read":    ic.bytesRead.Load(),
		"bytes_written": ic.bytesWritten.Load(),
		"chunks_read":   ic.chunksRead.Load(),
		"error":         err.Error(),
	})

	// Benign, expected terminations (graceful close by the remote or the
	// client going away) are logged at Debug to avoid flooding the default
	// Info-level output. Genuine transport failures (upstream reset/timeout)
	// and unclassified errors remain at Warn as diagnostic signal.
	switch kind {
	case errKindCleanEOF, errKindClientGone:
		entry.Debug("Tunnel connection error")
	default:
		entry.Warn("Tunnel connection error")
	}
}

// startIdleWatchdog launches a goroutine that logs once each time the tunnel
// crosses the idle threshold. It is edge-triggered via idleWarned.Swap(true):
// a single warning fires per idle episode, re-arming after any Read/Write
// activity resets idleWarned. The goroutine exits on the first tick after the
// conn is closed (design §2.4, REQ-TUN-004).
func (ic *instrumentedConn) startIdleWatchdog(threshold time.Duration) {
	go func() {
		ticker := time.NewTicker(threshold)
		defer ticker.Stop()
		for range ticker.C {
			if ic.closed.Load() {
				return
			}
			lastActivity := max(ic.lastRead.Load(), ic.lastWrite.Load())
			if lastActivity == 0 {
				lastActivity = ic.created.UnixNano()
			}
			idle := time.Since(time.Unix(0, lastActivity))
			if idle >= threshold && !ic.idleWarned.Swap(true) {
				ic.logger.WithFields(log.Fields{
					"idle":   idle.Round(time.Millisecond),
					"target": ic.target,
				}).Warn("Tunnel idle — at risk of upstream proxy timeout")
			}
		}
	}()
}

// wrapTunnelConn is the top-level wrapper applied to every successfully-dialed
// upstream tunnel connection (design §2.5, REQ-TUN-001/002/004/006). It enables
// TCP keepalive on the underlying socket, wraps the conn for metrics/error
// classification, and starts the idle watchdog when configured. It is applied
// only to upstream conns post-dial; dial errors short-circuit before this runs.
func wrapTunnelConn(conn net.Conn, target string) net.Conn {
	if keepAlivePeriod > 0 {
		enableTCPKeepAlive(conn, keepAlivePeriod)
	}

	ic := newInstrumentedConn(conn, target, log.WithField("tunnel_target", target))

	if idleWarnThreshold > 0 {
		ic.startIdleWatchdog(idleWarnThreshold)
	}

	return ic
}
