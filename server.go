package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
)

var (
	gKeepFile bool
	gFileList []string
	gConnList map[int]net.Conn
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
		return
	}
	defer l.Close()

	gConnList = make(map[int]net.Conn)

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
		log.Printf("listen: %s", s.Text())
		word, rest := splitWord(s.Text())
		switch word {
		case "save":
			saveFilesServer(s)
			log.Printf("listen: save, list: %v, keep: %t", gFileList, gKeepFile)
		case "load":
			loadFilesServer(c)
			log.Printf("listen: load, keep: %t", gKeepFile)
		case "conn":
			if rest != "" {
				word2, _ := splitWord(rest)
				id, err := strconv.Atoi(word2)
				if err != nil {
					log.Print("listen: conn: client id should be a number")
				} else {
					gConnList[id] = c
				}
			} else {
				log.Print("listen: conn: requires a client id")
			}
		case "send":
			if rest != "" {
				word2, rest2 := splitWord(rest)
				id, err := strconv.Atoi(word2)
				if err != nil {
					for _, c := range gConnList {
						fmt.Fprintln(c, rest)
					}
				} else {
					if c, ok := gConnList[id]; ok {
						fmt.Fprintln(c, rest2)
					}
				}
			}
		default:
			log.Printf("listen: unexpected command: %s", word)
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
