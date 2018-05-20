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
	gCopyFile bool
	gFileList []string
	gConnList = make(map[int]net.Conn)
)

func serve() {
	f, err := os.Create(gServerLogPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	log.SetOutput(f)

	log.Print("hi!")

	l, err := net.Listen(gSocketProt, gSocketPath)
	if err != nil {
		log.Printf("listening socket: %s", err)
		return
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
		log.Printf("listen: %s", s.Text())
		word, rest := splitWord(s.Text())
		switch word {
		case "save":
			s.Scan()
			switch s.Text() {
			case "copy":
				gCopyFile = true
			case "move":
				gCopyFile = false
			default:
				log.Printf("unexpected option to copy file(s): %s", s.Text())
				break
			}
			gFileList = nil
			for s.Scan() && s.Text() != "" {
				gFileList = append(gFileList, s.Text())
			}
		case "load":
			if gCopyFile {
				fmt.Fprintln(c, "copy")
			} else {
				fmt.Fprintln(c, "move")
			}
			for _, f := range gFileList {
				fmt.Fprintln(c, f)
			}
			fmt.Fprintln(c)
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

	if s.Err() != nil {
		log.Printf("listening: %s", s.Err())
	}

	c.Close()
}
