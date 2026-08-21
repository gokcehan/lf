package main

import (
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestDialServerWaitsForStartupRace pins #2660: an lfrc `lf -remote "send
// $id ..."` runs while this lf instance is still spawning the server, so
// the socket file may not exist yet. dialServer must wait briefly and then
// connect once the server creates the socket.
func TestDialServerWaitsForStartupRace(t *testing.T) {
	dir := t.TempDir()
	gSocketPath = filepath.Join(dir, "lf.sock")
	t.Cleanup(func() { gSocketPath = gDefaultSocketPath })

	ready := make(chan net.Listener, 1)
	go func() {
		// Simulate the server process creating the socket mid-race.
		time.Sleep(300 * time.Millisecond)
		l, err := net.Listen("unix", gSocketPath)
		if err != nil {
			close(ready)
			return
		}
		ready <- l
		go func() {
			for {
				conn, err := l.Accept()
				if err != nil {
					return
				}
				conn.Close()
			}
		}()
	}()

	start := time.Now()
	conn, err := dialServer()
	if err != nil {
		t.Fatalf("dialServer: %v", err)
	}
	defer conn.Close()

	if elapsed := time.Since(start); elapsed < 250*time.Millisecond {
		t.Errorf("dialServer connected after %v, expected it to wait for the late socket", elapsed)
	}
	if l, ok := <-ready; ok {
		defer l.Close()
	}
}

// TestDialServerFailsWhenServerNeverStarts ensures the wait is bounded: a
// socket that never appears errors out instead of hanging.
func TestDialServerFailsWhenServerNeverStarts(t *testing.T) {
	dir := t.TempDir()
	gSocketPath = filepath.Join(dir, "lf.sock")
	t.Cleanup(func() { gSocketPath = gDefaultSocketPath })

	if _, err := os.Stat(gSocketPath); !os.IsNotExist(err) {
		t.Fatalf("precondition: no socket at %s", gSocketPath)
	}

	start := time.Now()
	_, err := dialServer()
	if err == nil {
		t.Fatal("dialServer should fail when the server never starts")
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("dialServer took %v to fail, wait budget is 2s", elapsed)
	}
}

// TestDialServerFailsFastWhenSocketExistsButDead ensures the retry only
// covers the missing-socket race: a socket file that exists but refuses
// connections errors immediately.
func TestDialServerFailsFastWhenSocketExistsButDead(t *testing.T) {
	dir := t.TempDir()
	gSocketPath = filepath.Join(dir, "lf.sock")
	t.Cleanup(func() { gSocketPath = gDefaultSocketPath })

	// A path that exists but is not a listening socket (a plain file): the
	// stat succeeds, so dialServer must not enter the retry loop.
	if err := os.WriteFile(gSocketPath, []byte("not a socket"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(gSocketPath) })

	start := time.Now()
	_, err := dialServer()
	if err == nil {
		t.Fatal("dialServer should fail on a dead socket")
	}
	if elapsed := time.Since(start); elapsed > 250*time.Millisecond {
		t.Errorf("dialServer took %v on a dead socket, expected immediate failure", elapsed)
	}
}
