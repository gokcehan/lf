package main

import (
	"fmt"
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

func parseRuler() (*template.Template, error) {
	funcs := template.FuncMap{
		"df":       func() string { return diskFree(".") },
		"env":      os.Getenv,
		"humanize": humanize,
		"join":     strings.Join,
		"substr":   func(s string, start, length int) string { return string([]rune(s)[start : start+length]) },
	}

	for i := len(gRulerPaths) - 1; i >= 0; i-- {
		path := gRulerPaths[i]
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		log.Printf("reading file: %s", path)
		return template.New("ruler").Funcs(funcs).ParseFiles(path)
	}

	return template.New("ruler").Funcs(funcs).Parse(gDefaultRuler)
}

func renderRuler(ruler *template.Template, data rulerData, width int) (string, error) {
	var b strings.Builder
	if err := ruler.Execute(&b, data); err != nil {
		return "", err
	}

	s := strings.TrimSuffix(b.String(), "\n")
	s = strings.ReplaceAll(s, "\n", "\033[0;7m\\n\033[0m")
	sections := strings.Split(s, "\x1f")

	if len(sections) == 1 {
		return s, nil
	}

	wtot := 0
	for _, section := range sections {
		wtot += printLength(section)
	}

	wspacer := max(width-wtot, 0) / (len(sections) - 1)
	wspacerLast := max(width-wtot-wspacer*(len(sections)-2), 0)

	b.Reset()
	for i, section := range sections {
		switch i {
		case 0:
			b.WriteString(section)
		case len(sections) - 1:
			fmt.Fprintf(&b, "%*s%s", wspacerLast, "", section)
		default:
			fmt.Fprintf(&b, "%*s%s", wspacer, "", section)
		}
	}

	return b.String(), nil
}
