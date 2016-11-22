package main

import (
	"io"
	"io/ioutil"
	"log"
	"unicode"
)

type TokenType int

const (
	TokenEOF TokenType = iota
	// no explicit keyword type
	TokenIdent     // e.g. set, ratios, 1:2:3
	TokenColon     // :
	TokenPrefix    // $, !, &, / or ?
	TokenLBraces   // {{
	TokenRBraces   // }}
	TokenCommand   // in between a prefix to \n or between {{ and }}
	TokenSemicolon // ;
	// comments are stripped
)

type Scanner struct {
	buf []byte    // input buffer
	off int       // current offset in buf
	chr byte      // current character in buf
	sem bool      // insert semicolon
	nln bool      // insert newline
	eof bool      // buffer ended
	key bool      // scanning keys
	blk bool      // scanning block
	cmd bool      // scanning command
	typ TokenType // scanned token type
	tok string    // scanned token value
	// TODO: pos
}

func newScanner(r io.Reader) *Scanner {
	buf, err := ioutil.ReadAll(r)
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

	return &Scanner{
		buf: buf,
		eof: eof,
		chr: chr,
	}
}

func (s *Scanner) next() {
	if s.off+1 < len(s.buf) {
		s.off++
		s.chr = s.buf[s.off]
		return
	}

	s.off = len(s.buf)
	s.chr = 0
	s.eof = true
}

func (s *Scanner) peek() byte {
	if s.off+1 < len(s.buf) {
		return s.buf[s.off+1]
	}

	return 0
}

func isSpace(b byte) bool {
	return unicode.IsSpace(rune(b))
}

func isPrefix(b byte) bool {
	// TODO: how to differentiate slash in path vs search?
	return b == '$' || b == '!' || b == '&' // || b == '/' || b == '?'
}

func (s *Scanner) scan() bool {
scan:
	switch {
	case s.eof:
		s.next()
		if s.sem {
			s.typ = TokenSemicolon
			s.tok = "\n"
			s.sem = false
			return true
		}
		if s.nln {
			s.typ = TokenSemicolon
			s.tok = "\n"
			s.nln = false
			return true
		}
		s.typ = TokenEOF
		s.tok = "EOF"
		return false
	case s.key:
		beg := s.off
		for !s.eof && !isSpace(s.chr) {
			s.next()
		}
		s.typ = TokenIdent
		s.tok = string(s.buf[beg:s.off])
		s.key = false
	case s.blk:
		// return here by setting s.cmd to false
		// after scanning the command in the loop below
		if !s.cmd {
			s.next()
			s.next()
			s.typ = TokenRBraces
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
					s.typ = TokenCommand
					s.tok = string(s.buf[beg:s.off])
					s.cmd = false
					return true
				}
			}
		}

		s.typ = TokenEOF
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
				s.typ = TokenLBraces
				s.tok = "{{"
				s.blk = true
				return true
			}
		}

		beg := s.off

		for !s.eof && s.chr != '\n' {
			s.next()
		}

		s.typ = TokenCommand
		s.tok = string(s.buf[beg:s.off])
		s.cmd = false
		s.sem = true
	case s.chr == '\n':
		if s.sem {
			s.typ = TokenSemicolon
			s.tok = "\n"
			s.sem = false
			return true
		}
		s.next()
		if s.nln {
			s.typ = TokenSemicolon
			s.tok = "\n"
			s.nln = false
			return true
		}
		goto scan
	case isSpace(s.chr):
		for !s.eof && isSpace(s.chr) {
			s.next()
		}
		goto scan
	case s.chr == ';':
		s.typ = TokenSemicolon
		s.tok = ";"
		s.sem = false
		s.next()
	case s.chr == '#':
		for !s.eof && s.chr != '\n' {
			s.next()
		}
		goto scan
	case s.chr == '\'':
		s.next()
		beg := s.off
		for !s.eof && s.chr != '\'' {
			s.next()
		}
		s.typ = TokenIdent
		s.tok = string(s.buf[beg:s.off])
		s.next()
	case s.chr == ':':
		s.typ = TokenColon
		s.tok = ":"
		s.nln = true
		s.next()
	case s.chr == '{' && s.peek() == '{':
		s.next()
		s.next()
		s.typ = TokenLBraces
		s.tok = "{{"
		s.sem = false
		s.nln = false
	case s.chr == '}' && s.peek() == '}':
		s.next()
		s.next()
		s.typ = TokenRBraces
		s.tok = "}}"
		s.sem = true
	case isPrefix(s.chr):
		s.typ = TokenPrefix
		s.tok = string(s.chr)
		s.cmd = true
		s.next()
	default:
		beg := s.off
		for !s.eof && !isSpace(s.chr) && s.chr != ';' && s.chr != '#' {
			s.next()
		}

		s.typ = TokenIdent
		s.tok = string(s.buf[beg:s.off])
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
