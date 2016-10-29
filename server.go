package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

var (
	gKeepFile bool
	gFileList []string
	gConnList []net.Conn
)

func serve() {
	logFile, err := os.Create(gServerLogPath)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	log.Print("hi!")

	l, err := net.Listen("unix", gSocketPath)
	if err != nil {
		log.Printf("listening socket: %s", err)
	}
	defer l.Close()

	listen(l)
}

func listen(l net.Listener) {
	for {
		c, err := l.Accept()
		if err != nil {
			log.Printf("accepting connection: %s", err)
		}
		go handleConn(c)
	}
}

func handleConn(c net.Conn) {
	s := bufio.NewScanner(c)

	for s.Scan() {
		switch s.Text() {
		case "save":
			saveFilesServer(s)
			log.Printf("listen: save, list: %v, keep: %t", gFileList, gKeepFile)
		case "load":
			loadFilesServer(c)
			log.Printf("listen: load, keep: %t", gKeepFile)
		case "conn":
			gConnList = append(gConnList, c)
			log.Print("listen: conn")
		case "ping":
			for _, c := range gConnList {
				fmt.Fprintln(c, "echo ping from server")
			}
			log.Print("listen: ping")
		default:
			log.Print("listen: unexpected command")
		}
	}

	c.Close()
}

func saveFilesServer(s *bufio.Scanner) {
	switch s.Scan(); s.Text() {
	case "keep":
		gKeepFile = true
	case "move":
		gKeepFile = false
	default:
		log.Printf("unexpected option to keep file(s): %s", s.Text())
		return
	}

	gFileList = nil
	for s.Scan() {
		gFileList = append(gFileList, s.Text())
	}

	if s.Err() != nil {
		log.Printf("scanning: %s", s.Err())
		return
	}
}

func loadFilesServer(c net.Conn) {
	if gKeepFile {
		fmt.Fprintln(c, "keep")
	} else {
		fmt.Fprintln(c, "move")
	}

	for _, f := range gFileList {
		fmt.Fprintln(c, f)
	}

	c.Close()
}
