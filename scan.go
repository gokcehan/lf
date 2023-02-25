package main

import (
	"io"
	"log"
	"strconv"
)

type tokenType byte

const (
	tokenEOF tokenType = iota
	// no explicit keyword type
	tokenIdent     // e.g. set, ratios, 1:2:3
	tokenColon     // :
	tokenPrefix    // $, %, !, &
	tokenLBraces   // {{
	tokenRBraces   // }}
	tokenCommand   // in between a prefix to \n or between {{ and }}
	tokenSemicolon // ;
	// comments are stripped
)

type scanner struct {
	buf []byte    // input buffer
	off int       // current offset in buf
	chr byte      // current character in buf
	sem bool      // insert semicolon
	nln bool      // insert newline
	eof bool      // buffer ended
	key bool      // scanning keys
	blk bool      // scanning block
	cmd bool      // scanning command
	typ tokenType // scanned token type
	tok string    // scanned token value
	// TODO: pos
}

func newScanner(r io.Reader) *scanner {
	buf, err := io.ReadAll(r)
	if err != nil {
		log.Printf("scanning: %s", err)
	}

	var eof bool
	var chr byte

	if len(buf) == 0 {
		eof = true
	} else {
		eof = false
		chr = buf[0]
	}

	return &scanner{
		buf: buf,
		eof: eof,
		chr: chr,
	}
}

func (s *scanner) next() {
	if s.off+1 < len(s.buf) {
		s.off++
		s.chr = s.buf[s.off]
		return
	}

	s.off = len(s.buf)
	s.chr = 0
	s.eof = true
}

func (s *scanner) peek() byte {
	if s.off+1 < len(s.buf) {
		return s.buf[s.off+1]
	}

	return 0
}

func isSpace(b byte) bool {
	switch b {
	case '\t', '\n', '\v', '\f', '\r', ' ':
		return true
	}
	return false
}

func isDigit(b byte) bool {
	return '0' <= b && b <= '9'
}

func isPrefix(b byte) bool {
	switch b {
	case '$', '%', '!', '&':
		return true
	}
	return false
}

func (s *scanner) scan() bool {
scan:
	switch {
	case s.eof:
		s.next()
		if s.sem {
			s.typ = tokenSemicolon
			s.tok = "\n"
			s.sem = false
			return true
		}
		if s.nln {
			s.typ = tokenSemicolon
			s.tok = "\n"
			s.nln = false
			return true
		}
		s.typ = tokenEOF
		s.tok = "EOF"
		return false
	case s.key:
		beg := s.off
		for !s.eof && !isSpace(s.chr) {
			s.next()
		}
		s.typ = tokenIdent
		s.tok = string(s.buf[beg:s.off])
		s.key = false
	case s.blk:
		// return here by setting s.cmd to false
		// after scanning the command in the loop below
		if !s.cmd {
			s.next()
			s.next()
			s.typ = tokenRBraces
			s.tok = "}}"
			s.blk = false
			s.sem = true
			return true
		}

		beg := s.off

		for !s.eof {
			s.next()
			if s.chr == '}' {
				if !s.eof && s.peek() == '}' {
					s.typ = tokenCommand
					s.tok = string(s.buf[beg:s.off])
					s.cmd = false
					return true
				}
			}
		}

		s.typ = tokenEOF
		s.tok = "EOF"
		return false
	case s.cmd:
		for !s.eof && isSpace(s.chr) {
			s.next()
		}

		if !s.eof && s.chr == '{' {
			if s.peek() == '{' {
				s.next()
				s.next()
				s.typ = tokenLBraces
				s.tok = "{{"
				s.blk = true
				return true
			}
		}

		beg := s.off

		for !s.eof && s.chr != '\n' {
			s.next()
		}

		s.typ = tokenCommand
		s.tok = string(s.buf[beg:s.off])
		s.cmd = false
		s.sem = true
	case s.chr == '\n':
		if s.sem {
			s.typ = tokenSemicolon
			s.tok = "\n"
			s.sem = false
			return true
		}
		s.next()
		if s.nln {
			s.typ = tokenSemicolon
			s.tok = "\n"
			s.nln = false
			return true
		}
		goto scan
	case isSpace(s.chr):
		for !s.eof && isSpace(s.chr) && s.chr != '\n' {
			s.next()
		}
		goto scan
	case s.chr == ';':
		s.typ = tokenSemicolon
		s.tok = ";"
		s.sem = false
		s.next()
	case s.chr == '#':
		for !s.eof && s.chr != '\n' {
			s.next()
		}
		goto scan
	case s.chr == '"':
		s.next()
		var buf []byte
		for !s.eof && s.chr != '"' {
			if s.chr == '\\' {
				s.next()
				switch {
				case s.chr == '"' || s.chr == '\\':
					buf = append(buf, s.chr)
				case s.chr == 'a':
					buf = append(buf, '\a')
				case s.chr == 'b':
					buf = append(buf, '\b')
				case s.chr == 'f':
					buf = append(buf, '\f')
				case s.chr == 'n':
					buf = append(buf, '\n')
				case s.chr == 'r':
					buf = append(buf, '\r')
				case s.chr == 't':
					buf = append(buf, '\t')
				case s.chr == 'v':
					buf = append(buf, '\v')
				case isDigit(s.chr):
					var oct []byte
					for isDigit(s.chr) {
						oct = append(oct, s.chr)
						s.next()
					}
					n, err := strconv.ParseInt(string(oct), 8, 0)
					if err != nil {
						log.Printf("scanning: %s", err)
					}
					buf = append(buf, byte(n))
					continue
				default:
					buf = append(buf, '\\', s.chr)
				}
				s.next()
				continue
			}
			buf = append(buf, s.chr)
			s.next()
		}
		s.typ = tokenIdent
		s.tok = string(buf)
		s.next()
	case s.chr == '\'':
		s.next()
		beg := s.off
		for !s.eof && s.chr != '\'' {
			s.next()
		}
		s.typ = tokenIdent
		s.tok = string(s.buf[beg:s.off])
		s.next()
	case s.chr == ':':
		s.typ = tokenColon
		s.tok = ":"
		s.nln = true
		s.next()
	case s.chr == '{' && s.peek() == '{':
		s.next()
		s.next()
		s.typ = tokenLBraces
		s.tok = "{{"
		s.sem = false
		s.nln = false
	case s.chr == '}' && s.peek() == '}':
		s.next()
		s.next()
		s.typ = tokenRBraces
		s.tok = "}}"
		s.sem = true
	case isPrefix(s.chr):
		s.typ = tokenPrefix
		s.tok = string(s.chr)
		s.cmd = true
		s.next()
	default:
		var buf []byte
		for !s.eof && !isSpace(s.chr) && s.chr != ';' && s.chr != '#' {
			if s.chr == '\\' {
				s.next()
				buf = append(buf, s.chr)
				s.next()
				continue
			}
			buf = append(buf, s.chr)
			s.next()
		}

		s.typ = tokenIdent
		s.tok = string(buf)
		s.sem = true

		if s.tok == "push" {
			s.key = true
			for !s.eof && isSpace(s.chr) && s.chr != '\n' {
				s.next()
			}
		}
	}

	return true
}
