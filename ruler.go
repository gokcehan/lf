package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	_ "embed"
)

//go:embed etc/ruler.default
var gDefaultRuler string

type statData struct {
	Path        string
	Name        string
	Extension   string
	Size        uint64
	DirSize     *uint64
	DirCount    *uint64
	Permissions string
	ModTime     string
	AccessTime  string
	BirthTime   string
	ChangeTime  string
	LinkCount   string
	User        string
	Group       string
	Target      string
	CustomInfo  string
}

type rulerData struct {
	SPACER           string
	Message          string
	Keys             string
	Progress         []string
	Copy             []string
	Cut              []string
	Select           []string
	Visual           []string
	Index            int
	Total            int
	Hidden           int
	LinePercentage   string
	ScrollPercentage string
	Filter           []string
	Mode             string
	Options          map[string]string
	UserOptions      map[string]string
	Stat             *statData
}

func parseRuler(path string) (*template.Template, error) {
	funcs := template.FuncMap{
		"df":       func() string { return diskFree(".") },
		"env":      os.Getenv,
		"humanize": humanize,
		"join":     strings.Join,
		"lower":    strings.ToLower,
		"substr":   func(s string, start, length int) string { return string([]rune(s)[start : start+length]) },
		"upper":    strings.ToUpper,
	}

	if path == "" {
		return template.New("ruler").Funcs(funcs).Parse(gDefaultRuler)
	}

	return template.New(filepath.Base(path)).Funcs(funcs).ParseFiles(path)
}

func renderRuler(ruler *template.Template, data rulerData, width int) (string, string, error) {
	var b strings.Builder
	if err := ruler.Execute(&b, data); err != nil {
		return "", "", err
	}

	s := strings.ReplaceAll(b.String(), "\n", "")
	sections := strings.Split(s, "\x1f")

	if len(sections) == 1 {
		return s, "", nil
	}

	wtot := 0
	for _, section := range sections {
		wtot += printLength(section)
	}

	if wtot > width {
		return sections[0], strings.Join(sections[1:], ""), nil
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

	return b.String(), "", nil
}
