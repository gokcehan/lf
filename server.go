package main

import (
	"bufio"
	"fmt"
	"log"
	"maps"
	"net"
	"os"
	"slices"
	"strconv"
)

type srvCmd struct {
	op   string
	id   int
	msg  string
	c    net.Conn
	done chan struct{}
}

var (
	gCmdChan  = make(chan srvCmd)
	gQuitChan = make(chan struct{}, 1)
	gListener net.Listener
)

func serve() {
	if gLogPath != "" {
		f, err := os.OpenFile(gLogPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o600)
		if err != nil {
			log.Fatalf("failed to open log file: %s", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	log.Print("*************** starting server ***************")

	setUserUmask()

	l, err := net.Listen("unix", gSocketPath)
	if err != nil {
		log.Printf("listening socket: %s", err)
		return
	}
	defer l.Close()

	gListener = l

	go manage()

	listen(l)
}

func manage() {
	connList := make(map[int]net.Conn)
	for cmd := range gCmdChan {
		switch cmd.op {
		case "conn":
			// lifetime of the connection is managed by the server and
			// will be cleaned up via the `drop` command
			connList[cmd.id] = cmd.c
		case "drop":
			if c2, ok := connList[cmd.id]; ok {
				c2.Close()
				delete(connList, cmd.id)
			}
		case "list":
			for _, id := range slices.Sorted(maps.Keys(connList)) {
				fmt.Fprintln(cmd.c, id)
			}
		case "broadcast":
			for id, c2 := range connList {
				if _, err := fmt.Fprintln(c2, cmd.msg); err != nil {
					echoerrf(cmd.c, "failed to send command to client %v: %s", id, err)
				}
			}
		case "send":
			if c2, ok := connList[cmd.id]; ok {
				if _, err := fmt.Fprintln(c2, cmd.msg); err != nil {
					echoerrf(cmd.c, "failed to send command to client %v: %s", cmd.id, err)
				}
			} else {
				echoerr(cmd.c, "listen: send: no such client id is connected")
			}
		case "query":
			c2, ok := connList[cmd.id]
			if !ok {
				echoerr(cmd.c, "listen: query: no such client id is connected")
				break
			}
			if _, err := fmt.Fprintln(c2, "query "+cmd.msg); err != nil {
				echoerrf(cmd.c, "failed to send query to client %v: %s", cmd.id, err)
				break
			}
			s2 := bufio.NewScanner(c2)
			for s2.Scan() && s2.Text() != "" {
				if _, err := fmt.Fprintln(cmd.c, s2.Text()); err != nil {
					log.Printf("failed to forward query response from client %v: %s", cmd.id, err)
				}
			}
			if s2.Err() != nil {
				echoerrf(cmd.c, "failed to read query response from client %v: %s", cmd.id, s2.Err())
			}
		case "quit":
			if len(connList) == 0 {
				gQuitChan <- struct{}{}
				gListener.Close()
				close(cmd.done)
				return
			}
		case "quit!":
			gQuitChan <- struct{}{}
			for _, c := range connList {
				fmt.Fprintln(c, "echo server is quitting...")
				c.Close()
			}
			gListener.Close()
			close(cmd.done)
			return
		}
		if cmd.done != nil {
			close(cmd.done)
		}
	}
}

func listen(l net.Listener) {
	for {
		c, err := l.Accept()
		if err != nil {
			select {
			case <-gQuitChan:
				log.Print("*************** closing server ***************")
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

func echoerrf(c net.Conn, format string, a ...any) {
	echoerr(c, fmt.Sprintf(format, a...))
}

func send(cmd srvCmd) {
	cmd.done = make(chan struct{})
	gCmdChan <- cmd
	<-cmd.done
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
					send(srvCmd{op: "conn", id: id, c: c})
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
					send(srvCmd{op: "drop", id: id})
				}
			} else {
				echoerr(c, "listen: drop: requires a client id")
			}
		case "list":
			send(srvCmd{op: "list", c: c})
		case "send":
			if rest != "" {
				word2, rest2 := splitWord(rest)
				id, err := strconv.Atoi(word2)
				if err != nil {
					send(srvCmd{op: "broadcast", msg: rest, c: c})
				} else {
					send(srvCmd{op: "send", id: id, msg: rest2, c: c})
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
			send(srvCmd{op: "query", id: id, msg: rest2, c: c})
		case "quit":
			send(srvCmd{op: "quit"})
			break Loop
		case "quit!":
			send(srvCmd{op: "quit!"})
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
