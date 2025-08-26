package main

import (
	"log"
	"os"
	"strings"
	"text/template"

	_ "embed"
)

//go:embed etc/ruler.default
var gDefaultRuler string

type statData struct {
	Path        string
	Name        string
	Size        uint64
	Permissions string
	ModTime     string
	LinkCount   string
	User        string
	Group       string
	Target      string
}

type rulerData struct {
	SPACER      string
	Message     string
	Keys        string
	Progress    []string
	Copy        []string
	Cut         []string
	Select      []string
	Visual      []string
	Index       int
	Total       int
	Hidden      int
	Percentage  string
	Filter      []string
	Mode        string
	Options     map[string]string
	UserOptions map[string]string
	Stat        *statData
}

func parseRuler() *template.Template {
	funcs := template.FuncMap{
		"df":       func() string { return diskFree(".") },
		"env":      os.Getenv,
		"humanize": humanize,
		"join":     strings.Join,
		"substr":   func(s string, start, length int) string { return string([]rune(s)[start : start+length]) },
	}

	var ruler *template.Template

	for i := len(gRulerPaths) - 1; i >= 0; i-- {
		path := gRulerPaths[i]
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		log.Printf("reading file: %s", path)

		tmpl, err := template.New("ruler").Funcs(funcs).ParseFiles(path)
		if err != nil {
			log.Printf("parsing ruler file: %s", err)
			continue
		}

		ruler = tmpl
		break
	}

	if ruler == nil {
		ruler, _ = template.New("ruler").Funcs(funcs).Parse(gDefaultRuler)
	}

	return ruler
}

func renderRuler(ruler *template.Template, data rulerData) (string, string, error) {
	var b strings.Builder
	if err := ruler.Execute(&b, data); err != nil {
		return "", "", err
	}

	s := strings.TrimSuffix(b.String(), "\n")
	s = strings.ReplaceAll(s, "\n", "\033[0;7m\\n\033[0m")

	left, right, found := strings.Cut(s, "\x1f")
	if !found {
		return s, "", nil
	}

	return left, right, nil
}
