package main

// Grammar of the language used in the evaluator
//
// Expr     = SetExpr
//          | MapExpr
//          | CMapExpr
//          | CmdExpr
//          | CallExpr
//          | ExecExpr
//          | ListExpr
//
// SetExpr  = 'set' <opt> <val> ';'
//
// MapExpr  = 'map' <keys> Expr ';'
//
// CMapExpr = 'cmap' <key> <cmd> ';'
//
// CmdExpr  = 'cmd' <name> Expr ';'
//
// CallExpr = <name> <args> ';'
//
// ExecExpr = Prefix      <expr>      '\n'
//          | Prefix '{{' <expr> '}}' ';'
//
// Prefix   = '$' | '!' | '&' | '/' | '?'
//
// ListExpr = ':'      ListRest      '\n'
//          | ':' '{{' ListRest '}}' ';'
//
// ListRest = Nil
//          | Expr ListRest

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

type expr interface {
	String() string

	eval(app *app, args []string)
	// TODO: add a bind method to avoid passing args in eval
}

type setExpr struct {
	opt string
	val string
}

func (e *setExpr) String() string { return fmt.Sprintf("set %s %s", e.opt, e.val) }

type mapExpr struct {
	keys string
	expr expr
}

func (e *mapExpr) String() string { return fmt.Sprintf("map %s %s", e.keys, e.expr) }

type cmapExpr struct {
	key string
	cmd string
}

func (e *cmapExpr) String() string { return fmt.Sprintf("cmap %s %s", e.key, e.cmd) }

type cmdExpr struct {
	name string
	expr expr
}

func (e *cmdExpr) String() string { return fmt.Sprintf("cmd %s %s", e.name, e.expr) }

type callExpr struct {
	name string
	args []string
}

func (e *callExpr) String() string { return fmt.Sprintf("%s -- %s", e.name, e.args) }

type execExpr struct {
	pref string
	expr string
}

func (e *execExpr) String() string {
	var buf bytes.Buffer

	buf.WriteString(e.pref)
	buf.WriteString("{{ ")

	lines := strings.Split(e.expr, "\n")

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
}

func (e *listExpr) String() string {
	buf := []byte{':', '{', '{', ' '}
	for _, expr := range e.exprs {
		buf = append(buf, expr.String()...)
		buf = append(buf, ';', ' ')
	}
	buf = append(buf, '}', '}')
	return string(buf)
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
			var cmd string

			s.scan()
			key := s.tok

			s.scan()
			if s.typ != tokenSemicolon {
				if s.typ != tokenIdent {
					p.err = fmt.Errorf("expected command: %s", s.tok)
				}
				cmd = s.tok
				s.scan()
			}

			s.scan()

			result = &cmapExpr{key, cmd}
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

			result = &callExpr{name, args}
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

		result = &listExpr{exprs}
	case tokenPrefix:
		var expr string

		pref := s.tok

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

		result = &execExpr{pref, expr}
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
