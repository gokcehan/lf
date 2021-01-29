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

	f, err := os.Create(gLogPath)
	if err != nil {
		panic(err)
	}
	defer os.Remove(gLogPath)
	defer f.Close()
	log.SetOutput(f)

	log.Print("hi!")

	app := newApp(screen)

	if err := app.nav.readMarks(); err != nil {
		app.ui.echoerrf("reading marks file: %s", err)
	}

	if err := app.readHistory(); err != nil {
		app.ui.echoerrf("reading history file: %s", err)
	}

	go app.nav.previewLoop(app.ui)
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

func saveFiles(list []string, cp bool) error {
	c, err := net.Dial(gSocketProt, gSocketPath)
	if err != nil {
		return fmt.Errorf("dialing to save files: %s", err)
	}
	defer c.Close()

	log.Printf("saving files: %v", list)

	fmt.Fprintln(c, "save")

	if cp {
		fmt.Fprintln(c, "copy")
	} else {
		fmt.Fprintln(c, "move")
	}

	for _, f := range list {
		fmt.Fprintln(c, f)
	}
	fmt.Fprintln(c)

	return nil
}

func loadFiles() (list []string, cp bool, err error) {
	c, e := net.Dial(gSocketProt, gSocketPath)
	if e != nil {
		err = fmt.Errorf("dialing to load files: %s", e)
		return
	}
	defer c.Close()

	fmt.Fprintln(c, "load")

	s := bufio.NewScanner(c)

	s.Scan()

	switch s.Text() {
	case "copy":
		cp = true
	case "move":
		cp = false
	default:
		err = fmt.Errorf("unexpected option to copy file(s): %s", s.Text())
		return
	}

	for s.Scan() && s.Text() != "" {
		list = append(list, s.Text())
	}

	if s.Err() != nil {
		err = fmt.Errorf("scanning file list: %s", s.Err())
		return
	}

	log.Printf("loading files: %v", list)

	return
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
