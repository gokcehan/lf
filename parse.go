package main

// Grammar of the language used in the evaluator
//
// Expr         = SetExpr
//              | SetLocalExpr
//              | MapExpr
//              | CMapExpr
//              | CmdExpr
//              | CallExpr
//              | ExecExpr
//              | ListExpr
//
// SetExpr      = 'set' <opt> <val> ';'
//
// SetLocalExpr = 'setlocal' <dir> <opt> <val> ';'
//
// MapExpr      = 'map' <keys> Expr
//
// CMapExpr     = 'cmap' <key> Expr
//
// CmdExpr      = 'cmd' <name> Expr
//
// CallExpr     = <name> <args> ';'
//
// ExecExpr     = Prefix      <value>      '\n'
//              | Prefix '{{' <value> '}}' ';'
//
// Prefix       = '$' | '%' | '!' | '&'
//
// ListExpr     = ':'      Expr ListRest      '\n'
//              | ':' '{{' Expr ListRest '}}' ';'
//
// ListRest     = Nil
//              | Expr ListExpr

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

type expr interface {
	String() string
	eval(app *app, args []string)
}

type setExpr struct {
	opt string
	val string
}

func (e *setExpr) String() string { return fmt.Sprintf("set %s %s", e.opt, e.val) }

type setLocalExpr struct {
	path string
	opt  string
	val  string
}

func (e *setLocalExpr) String() string { return fmt.Sprintf("setlocal %s %s %s", e.path, e.opt, e.val) }

type mapExpr struct {
	keys string
	expr expr
}

func (e *mapExpr) String() string { return fmt.Sprintf("map %s %s", e.keys, e.expr) }

type cmapExpr struct {
	key  string
	expr expr
}

func (e *cmapExpr) String() string { return fmt.Sprintf("cmap %s %s", e.key, e.expr) }

type cmdExpr struct {
	name string
	expr expr
}

func (e *cmdExpr) String() string { return fmt.Sprintf("cmd %s %s", e.name, e.expr) }

type callExpr struct {
	name  string
	args  []string
	count int
}

func (e *callExpr) String() string { return fmt.Sprintf("%s -- %s", e.name, e.args) }

type execExpr struct {
	prefix string
	value  string
}

func (e *execExpr) String() string {
	var buf bytes.Buffer

	buf.WriteString(e.prefix)
	buf.WriteString("{{ ")

	lines := strings.Split(e.value, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		buf.WriteString(trimmed)

		if len(lines) > 1 {
			buf.WriteString(" ...")
		}

		break
	}

	buf.WriteString(" }}")

	return buf.String()
}

type listExpr struct {
	exprs []expr
	count int
}

func (e *listExpr) String() string {
	var buf bytes.Buffer

	buf.WriteString(":{{ ")

	for _, expr := range e.exprs {
		buf.WriteString(expr.String())
		buf.WriteString("; ")
	}

	buf.WriteString("}}")

	return buf.String()
}

type parser struct {
	scanner *scanner
	expr    expr
	err     error
}

func newParser(r io.Reader) *parser {
	scanner := newScanner(r)

	scanner.scan()

	return &parser{
		scanner: scanner,
	}
}

func (p *parser) parseExpr() expr {
	s := p.scanner

	var result expr

	switch s.typ {
	case tokenEOF:
		return nil
	case tokenIdent:
		switch s.tok {
		case "set":
			var val string

			s.scan()
			if s.typ != tokenIdent {
				p.err = fmt.Errorf("expected identifier: %s", s.tok)
			}
			opt := s.tok

			s.scan()
			if s.typ != tokenSemicolon {
				val = s.tok
				s.scan()
			}

			s.scan()

			result = &setExpr{opt, val}
		case "setlocal":
			var val string

			s.scan()
			if s.typ != tokenIdent {
				p.err = fmt.Errorf("expected directory: %s", s.tok)
			}
			dir := s.tok

			s.scan()
			if s.typ != tokenIdent {
				p.err = fmt.Errorf("expected identifier: %s", s.tok)
			}
			opt := s.tok

			s.scan()
			if s.typ != tokenSemicolon {
				val = s.tok
				s.scan()
			}

			s.scan()

			result = &setLocalExpr{dir, opt, val}
		case "map":
			var expr expr

			s.scan()
			keys := s.tok

			s.scan()
			if s.typ != tokenSemicolon {
				expr = p.parseExpr()
			} else {
				s.scan()
			}

			result = &mapExpr{keys, expr}
		case "cmap":
			var expr expr

			s.scan()
			key := s.tok

			s.scan()
			if s.typ != tokenSemicolon {
				expr = p.parseExpr()
			} else {
				s.scan()
			}

			result = &cmapExpr{key, expr}
		case "cmd":
			var expr expr

			s.scan()
			name := s.tok

			s.scan()
			if s.typ != tokenSemicolon {
				expr = p.parseExpr()
			} else {
				s.scan()
			}

			result = &cmdExpr{name, expr}
		default:
			name := s.tok

			var args []string
			for s.scan() && s.typ != tokenSemicolon {
				args = append(args, s.tok)
			}

			s.scan()

			result = &callExpr{name, args, 1}
		}
	case tokenColon:
		s.scan()

		var exprs []expr
		if s.typ == tokenLBraces {
			s.scan()
			for {
				e := p.parseExpr()
				if e == nil {
					return nil
				}
				exprs = append(exprs, e)
				if s.typ == tokenRBraces {
					break
				}
			}
			s.scan()
		} else {
			for {
				e := p.parseExpr()
				if e == nil {
					return nil
				}
				exprs = append(exprs, e)
				if s.tok == "\n" {
					break
				}
			}
		}

		s.scan()

		result = &listExpr{exprs, 1}
	case tokenPrefix:
		var expr string

		prefix := s.tok

		s.scan()
		if s.typ == tokenLBraces {
			s.scan()
			expr = s.tok
			s.scan()
		} else {
			expr = s.tok
		}

		s.scan()
		s.scan()

		result = &execExpr{prefix, expr}
	default:
		p.err = fmt.Errorf("unexpected token: %s", s.tok)
	}

	return result
}

func (p *parser) parse() bool {
	if p.expr = p.parseExpr(); p.expr == nil {
		return false
	}

	return true
}
