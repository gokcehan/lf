package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
)

func run() {
	var screen tcell.Screen
	var err error
	if screen, err = tcell.NewScreen(); err != nil {
		log.Fatalf("creating screen: %s", err)
	} else if err = screen.Init(); err != nil {
		log.Fatalf("initializing screen: %s", err)
	}
	if gOpts.mouse {
		screen.EnableMouse()
	}

	if gLogPath != "" {
		f, err := os.OpenFile(gLogPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		log.SetOutput(f)
	} else {
		log.SetOutput(io.Discard)
	}

	log.Print("hi!")

	ui := newUI(screen)
	nav := newNav(ui.wins[0].h)
	app := newApp(ui, nav)

	if err := nav.sync(); err != nil {
		app.ui.echoerrf("sync: %s", err)
	}

	if err := app.nav.readMarks(); err != nil {
		app.ui.echoerrf("reading marks file: %s", err)
	}

	if err := app.nav.readTags(); err != nil {
		app.ui.echoerrf("reading tags file: %s", err)
	}

	if err := app.readHistory(); err != nil {
		app.ui.echoerrf("reading history file: %s", err)
	}

	app.loop()

	app.ui.screen.Fini()
}

func readExpr() <-chan expr {
	ch := make(chan expr)

	go func() {
		duration := 1 * time.Second

		c, err := net.Dial(gSocketProt, gSocketPath)
		for err != nil {
			log.Printf("connecting server: %s", err)
			time.Sleep(duration)
			duration *= 2
			c, err = net.Dial(gSocketProt, gSocketPath)
		}

		fmt.Fprintf(c, "conn %d\n", gClientID)

		ch <- &callExpr{"sync", nil, 1}

		s := bufio.NewScanner(c)
		for s.Scan() {
			log.Printf("recv: %s", s.Text())
			p := newParser(strings.NewReader(s.Text()))
			if p.parse() {
				ch <- p.expr
			}
		}

		c.Close()
	}()

	return ch
}

func remote(cmd string) error {
	c, err := net.Dial(gSocketProt, gSocketPath)
	if err != nil {
		return fmt.Errorf("dialing to send server: %s", err)
	}

	fmt.Fprintln(c, cmd)

	// XXX: Standard net.Conn interface does not include a CloseWrite method
	// but net.UnixConn and net.TCPConn implement it so the following should be
	// safe as long as we do not use other types of connections. We need
	// CloseWrite to notify the server that this is not a persistent connection
	// and it should be closed after the response.
	if v, ok := c.(interface {
		CloseWrite() error
	}); ok {
		v.CloseWrite()
	}

	io.Copy(os.Stdout, c)

	c.Close()

	return nil
}
