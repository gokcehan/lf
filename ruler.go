package main

import (
	"log"
	"os"
	"strings"
	"text/template"
)

type rulerData struct {
	ESC         string
	Acc         string
	Progress    []string
	Cut         int
	Copy        int
	Select      int
	Visual      int
	Index       int
	Total       int
	Hidden      int
	Percentage  string
	Filter      []string
	Mode        string
	Options     map[string]string
	UserOptions map[string]string
}

func parseRuler() *template.Template {
	funcs := template.FuncMap{
		"df":   func() string { return diskFree(".") },
		"env":  os.Getenv,
		"join": strings.Join,
	}

	var tmpl *template.Template

	for _, path := range gRulerPaths {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			log.Printf("reading file: %s", path)

			tmpl, err = template.New("ruler").Funcs(funcs).ParseFiles(path)
			if err != nil {
				log.Printf("reading ruler file: %s", err)
				continue
			}
		}
	}

	return tmpl
}

func renderRuler(tmpl *template.Template, data rulerData) (string, error) {
	var b strings.Builder
	if err := tmpl.Execute(&b, data); err != nil {
		return "", err
	}

	return strings.ReplaceAll(b.String(), "\n", ""), nil
}
