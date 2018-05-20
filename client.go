package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"github.com/nsf/termbox-go"
)

func run() {
	f, err := os.Create(gLogPath)
	if err != nil {
		panic(err)
	}
	defer os.Remove(gLogPath)
	defer f.Close()
	log.SetOutput(f)

	log.Print("hi!")

	if err := termbox.Init(); err != nil {
		log.Fatalf("initializing termbox: %s", err)
	}
	defer termbox.Close()

	termbox.SetOutputMode(termbox.Output256)

	app := newApp()

	if _, err := os.Stat(gConfigPath); !os.IsNotExist(err) {
		app.readFile(gConfigPath)
	}

	app.loop()
}

func readExpr() <-chan expr {
	ch := make(chan expr)

	go func() {
		c, err := net.Dial(gSocketProt, gSocketPath)
		if err != nil {
			log.Printf(fmt.Sprintf("connecting server: %s", err))
			return
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

func saveFiles(list []string, copy bool) error {
	c, err := net.Dial(gSocketProt, gSocketPath)
	if err != nil {
		return fmt.Errorf("dialing to save files: %s", err)
	}
	defer c.Close()

	log.Printf("saving files: %v", list)

	fmt.Fprintln(c, "save")

	if copy {
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

func loadFiles() (list []string, copy bool, err error) {
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
		copy = true
	case "move":
		copy = false
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

func sendRemote(cmd string) error {
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
