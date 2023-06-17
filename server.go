package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

var (
	gConnList = make(map[int]net.Conn)
	gConnData = make(map[int]map[string]string)
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
					gConnList[id] = c
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
					delete(gConnList, id)
					delete(gConnData, id)
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
		case "store":
			idStr, key := splitWord(rest)
			id, err := strconv.Atoi(idStr)
			if err != nil {
				echoerr(c, "listen: store: client id should be a number")
				break
			}
			data, ok := gConnData[id]
			if !ok {
				data = make(map[string]string)
				gConnData[id] = data
			}
			var builder strings.Builder
			for s.Scan() {
				fmt.Fprintln(&builder, s.Text())
			}
			data[key] = builder.String()
		case "query":
			idStr, key := splitWord(rest)
			id, err := strconv.Atoi(idStr)
			if err != nil {
				echoerr(c, "listen: query: client id should be a number")
				break
			}
			data, ok := gConnData[id]
			if !ok {
				echoerr(c, "listen: query: client id does not exist")
				break
			}
			io.WriteString(c, data[key])
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
