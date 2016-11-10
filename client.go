package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/nsf/termbox-go"
)

func client() {
	logFile, err := os.Create(gLogPath)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	log.Print("hi!")

	if err := termbox.Init(); err != nil {
		log.Fatalf("initializing termbox: %s", err)
	}
	defer termbox.Close()

	app := newApp()

	app.ui.loadFile(app.nav)

	if err := app.nav.sync(); err != nil {
		msg := fmt.Sprintf("sync: %s", err)
		app.ui.message = msg
		log.Printf(msg)
	}

	if _, err := os.Stat(gConfigPath); err == nil {
		log.Printf("reading configuration file: %s", gConfigPath)

		rcFile, err := os.Open(gConfigPath)
		if err != nil {
			msg := fmt.Sprintf("opening configuration file: %s", err)
			app.ui.message = msg
			log.Printf(msg)
		}
		defer rcFile.Close()

		p := newParser(rcFile)
		for p.parse() {
			p.expr.eval(app, nil)
		}

		if p.err != nil {
			app.ui.message = p.err.Error()
			log.Print(p.err)
		}
	}

	app.ui.draw(app.nav)

	app.handleInp()
}

func readExpr(c net.Conn) chan Expr {
	ch := make(chan Expr)

	go func() {
		fmt.Fprintf(c, "conn %d\n", gClientId)

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
	c, err := net.Dial("unix", gSocketPath)
	if err != nil {
		return fmt.Errorf("dialing to save files: %s", err)
	}
	defer c.Close()

	log.Printf("saving files: %v", list)

	fmt.Fprint(c, "save ")

	if copy {
		fmt.Fprint(c, "copy ")
	} else {
		fmt.Fprint(c, "move ")
	}

	fmt.Fprintln(c, strings.Join(list, ":"))

	return nil
}

func loadFiles() (list []string, copy bool, err error) {
	c, e := net.Dial("unix", gSocketPath)
	if e != nil {
		err = fmt.Errorf("dialing to load files: %s", e)
		return
	}
	defer c.Close()

	fmt.Fprintln(c, "load")

	s := bufio.NewScanner(c)

	s.Scan()

	word, rest := splitWord(s.Text())
	log.Printf("load: %s", s.Text())

	switch word {
	case "copy":
		copy = true
	case "move":
		copy = false
	default:
		err = fmt.Errorf("unexpected option to copy file(s): %s", word)
		return
	}

	list = strings.Split(rest, ":")

	if s.Err() != nil {
		err = fmt.Errorf("scanning file list: %s", s.Err())
		return
	}

	log.Printf("loading files: %v", list)

	return
}

func sendServer(cmd string) error {
	c, err := net.Dial("unix", gSocketPath)
	if err != nil {
		return fmt.Errorf("dialing to send server: %s", err)
	}
	defer c.Close()

	fmt.Fprintln(c, "send sync")

	return nil
}
