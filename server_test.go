package main

import (
	"errors"
	"net"
	"sync/atomic"
	"testing"
	"time"
)

// flakyListener fails the first Accept call, then serves queued connections.
type flakyListener struct {
	calls atomic.Int32
	conns chan net.Conn
}

func (l *flakyListener) Accept() (net.Conn, error) {
	if l.calls.Add(1) == 1 {
		return nil, errors.New("transient accept error")
	}
	c, ok := <-l.conns
	if !ok {
		return nil, net.ErrClosed
	}
	return c, nil
}

func (l *flakyListener) Close() error   { close(l.conns); return nil }
func (l *flakyListener) Addr() net.Addr { return &net.UnixAddr{Name: "test", Net: "unix"} }

// closedListener always fails with net.ErrClosed.
type closedListener struct{}

func (l *closedListener) Accept() (net.Conn, error) { return nil, net.ErrClosed }
func (l *closedListener) Close() error              { return nil }
func (l *closedListener) Addr() net.Addr            { return &net.UnixAddr{Name: "test", Net: "unix"} }

// TestListenSurvivesAcceptError verifies that a non-fatal Accept error does not
// stop the server or hand a nil connection to handleConn.
func TestListenSurvivesAcceptError(t *testing.T) {
	fl := &flakyListener{conns: make(chan net.Conn, 4)}
	go listen(fl)

	// Let the listener hit its first Accept error.
	time.Sleep(50 * time.Millisecond)

	for i := 0; i < 3; i++ {
		srv, cli := net.Pipe()
		fl.conns <- srv
		_ = cli.Close()
		time.Sleep(20 * time.Millisecond)
	}

	// 1 failed Accept + 3 accepted connections.
	if got := fl.calls.Load(); got < 4 {
		t.Fatalf("listener stopped accepting after an error: calls=%d", got)
	}
}

// TestListenReturnsWhenListenerClosed verifies that a closed listener stops the
// accept loop instead of spinning on net.ErrClosed or panicking on a nil conn.
func TestListenReturnsWhenListenerClosed(t *testing.T) {
	done := make(chan struct{})
	go func() {
		listen(&closedListener{})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("listen did not return after the listener was closed")
	}
}
