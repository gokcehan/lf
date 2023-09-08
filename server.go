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
	gConnList = make(map[int]net.Conn)
	gQuitChan = make(chan struct{}, 1)
	gListener net.Listener
)

func serve() {
	if gLogPath != "" {
		f, err := os.OpenFile(gLogPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	log.Print("hi!")

	if gSocketProt == "unix" {
		setUserUmask()
	}

	l, err := net.Listen(gSocketProt, gSocketPath)
	if err != nil {
		log.Printf("listening socket: %s", err)
		return
	}
	defer l.Close()

	gListener = l

	listen(l)
}

func listen(l net.Listener) {
	for {
		c, err := l.Accept()
		if err != nil {
			select {
			case <-gQuitChan:
				log.Printf("bye!")
				return
			default:
				log.Printf("accepting connection: %s", err)
			}
		}
		go handleConn(c)
	}
}

func echoerr(c net.Conn, msg string) {
	fmt.Fprintln(c, msg)
	log.Print(msg)
}

func echoerrf(c net.Conn, format string, a ...interface{}) {
	echoerr(c, fmt.Sprintf(format, a...))
}

func handleConn(c net.Conn) {
	s := bufio.NewScanner(c)

Loop:
	for s.Scan() {
		log.Printf("listen: %s", s.Text())
		word, rest := splitWord(s.Text())
		switch word {
		case "conn":
			if rest != "" {
				word2, _ := splitWord(rest)
				id, err := strconv.Atoi(word2)
				if err != nil {
					echoerr(c, "listen: conn: client id should be a number")
				} else {
					// lifetime of the connection is managed by the server and
					// will be cleaned up via the `drop` command
					gConnList[id] = c
					return
				}
			} else {
				echoerr(c, "listen: conn: requires a client id")
			}
		case "drop":
			if rest != "" {
				word2, _ := splitWord(rest)
				id, err := strconv.Atoi(word2)
				if err != nil {
					echoerr(c, "listen: drop: client id should be a number")
				} else {
					if c2, ok := gConnList[id]; ok {
						c2.Close()
					}
					delete(gConnList, id)
				}
			} else {
				echoerr(c, "listen: drop: requires a client id")
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
					if c2, ok := gConnList[id]; ok {
						fmt.Fprintln(c2, rest2)
					} else {
						echoerr(c, "listen: send: no such client id is connected")
					}
				}
			}
		case "query":
			if rest == "" {
				echoerr(c, "listen: query: requires a client id")
				break
			}
			word2, rest2 := splitWord(rest)
			id, err := strconv.Atoi(word2)
			if err != nil {
				echoerr(c, "listen: query: client id should be a number")
				break
			}
			c2, ok := gConnList[id]
			if !ok {
				echoerr(c, "listen: query: no such client id is connected")
				break
			}
			fmt.Fprintln(c2, "query "+rest2)
			s2 := bufio.NewScanner(c2)
			for s2.Scan() && s2.Text() != "" {
				fmt.Fprintln(c, s2.Text())
			}
		case "quit":
			if len(gConnList) == 0 {
				gQuitChan <- struct{}{}
				gListener.Close()
				break Loop
			}
		case "quit!":
			gQuitChan <- struct{}{}
			for _, c := range gConnList {
				fmt.Fprintln(c, "echo server is quitting...")
				c.Close()
			}
			gListener.Close()
			break Loop
		default:
			echoerrf(c, "listen: unexpected command: %s", word)
		}
	}

	if s.Err() != nil {
		echoerrf(c, "listening: %s", s.Err())
	}

	c.Close()
}
