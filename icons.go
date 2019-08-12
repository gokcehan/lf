package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

type iconMap map[string]string

func parseIcons() iconMap {
	if env := os.Getenv("LF_ICONS"); env != "" {
		return parseIconsEnv(env)
	}

	defaultIcons := []string{
		"tw=🗀",
		"st=🗀",
		"ow=🗀",
		"di=🗀",
		"fi=🗎",
	}

	return parseIconsEnv(strings.Join(defaultIcons, ":"))
}

func parseIconsEnv(env string) iconMap {
	icons := make(iconMap)

	entries := strings.Split(env, ":")
	for _, entry := range entries {
		if entry == "" {
			continue
		}
		pair := strings.Split(entry, "=")
		if len(pair) != 2 {
			log.Printf("invalid $LF_ICONS entry: %s", entry)
			return icons
		}
		key, val := pair[0], pair[1]
		icons[key] = val
	}

	return icons
}

func (im iconMap) get(f *file) string {
	var key string

	switch {
	case f.IsDir() && f.Mode()&os.ModeSticky != 0 && f.Mode()&0002 != 0:
		key = "tw"
	case f.IsDir() && f.Mode()&os.ModeSticky != 0:
		key = "st"
	case f.IsDir() && f.Mode()&0002 != 0:
		key = "ow"
	case f.IsDir():
		key = "di"
	case f.linkState == working:
		key = "ln"
	case f.linkState == broken:
		key = "or"
	case f.Mode()&os.ModeNamedPipe != 0:
		key = "pi"
	case f.Mode()&os.ModeSocket != 0:
		key = "so"
	case f.Mode()&os.ModeCharDevice != 0:
		key = "cd"
	case f.Mode()&os.ModeDevice != 0:
		key = "bd"
	case f.Mode()&os.ModeSetuid != 0:
		key = "su"
	case f.Mode()&os.ModeSetgid != 0:
		key = "sg"
	case f.Mode().IsRegular() && f.Mode()&0111 != 0:
		key = "ex"
	default:
		key = "*" + filepath.Ext(f.Name())
	}

	if val, ok := im[key]; ok {
		return val
	}

	if val, ok := im["fi"]; ok {
		return val
	}

	return " "
}
