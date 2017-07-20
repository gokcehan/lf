package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

var (
	gCopyFile bool
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

	l, err := net.Listen(gSocketProt, gSocketPath)
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
			word2, rest2 := splitWord(rest)
			switch word2 {
			case "copy":
				gCopyFile = true
			case "move":
				gCopyFile = false
			default:
				log.Printf("unexpected option to copy file(s): %s", word2)
				break
			}
			gFileList = strings.Split(rest2, ":")
		case "load":
			if gCopyFile {
				fmt.Fprint(c, "copy ")
			} else {
				fmt.Fprint(c, "move ")
			}
			fmt.Fprintln(c, strings.Join(gFileList, ":"))
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
