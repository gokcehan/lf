package main

import (
	"bufio"
	"fmt"
	"net"
	"path/filepath"
	"testing"
	"time"
)

func TestReadExprRegistersBeforeReturning(t *testing.T) {
	oldSocketPath, oldClientID := gSocketPath, gClientID
	defer func() { gSocketPath, gClientID = oldSocketPath, oldClientID }()

	gSocketPath = filepath.Join(t.TempDir(), "lf.sock")
	gClientID = 1234

	// simulate a server that starts listening after the client is launched
	lines := make(chan string, 1)
	go func() {
		time.Sleep(50 * time.Millisecond)

		l, err := net.Listen("unix", gSocketPath)
		if err != nil {
			lines <- err.Error()
			return
		}
		defer l.Close()

		c, err := l.Accept()
		if err != nil {
			lines <- err.Error()
			return
		}
		defer c.Close()

		s := bufio.NewScanner(c)
		s.Scan()
		lines <- s.Text()
	}()

	readExpr()

	// the registration has already been written when readExpr returns, so it
	// should arrive well before the client would have retried otherwise
	select {
	case got := <-lines:
		if expected := fmt.Sprintf("conn %d", gClientID); got != expected {
			t.Errorf("expected '%s' but got '%s'", expected, got)
		}
	case <-time.After(40 * time.Millisecond):
		t.Error("client was not registered when readExpr returned")
	}
}
