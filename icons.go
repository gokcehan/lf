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
		"fi=🗎",
		"di=🗀",
		"ln=🗎",
		"pi=🗎",
		"so=🗎",
		"bd=🗎",
		"cd=🗎",
		"or=🗎",
		"su=🗎",
		"sg=🗎",
		"tw=🗀",
		"ow=🗀",
		"st=🗀",
		"ex=🗎",
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
		key = replaceTilde(key)
		if filepath.IsAbs(key) {
			key = filepath.Clean(key)
		}
		icons[key] = val
	}

	return icons
}

func (im iconMap) get(f *file) string {
	if val, ok := im[f.path]; ok {
		return val
	}

	if f.IsDir() {
		if val, ok := im[f.Name()+"/"]; ok {
			return val
		}
	}

	if val, ok := im[f.Name()]; ok {
		return val
	}

	if val, ok := im[filepath.Base(f.Name())+".*"]; ok {
		return val
	}

	if val, ok := im["*"+f.ext]; ok {
		return val
	}

	var key string

	switch {
	case f.linkState == working:
		key = "ln"
	case f.linkState == broken:
		key = "or"
	case f.IsDir() && f.Mode()&os.ModeSticky != 0 && f.Mode()&0002 != 0:
		key = "tw"
	case f.IsDir() && f.Mode()&0002 != 0:
		key = "ow"
	case f.IsDir() && f.Mode()&os.ModeSticky != 0:
		key = "st"
	case f.IsDir():
		key = "di"
	case f.Mode()&os.ModeNamedPipe != 0:
		key = "pi"
	case f.Mode()&os.ModeSocket != 0:
		key = "so"
	case f.Mode()&os.ModeDevice != 0:
		key = "bd"
	case f.Mode()&os.ModeCharDevice != 0:
		key = "cd"
	case f.Mode()&os.ModeSetuid != 0:
		key = "su"
	case f.Mode()&os.ModeSetgid != 0:
		key = "sg"
	case f.Mode()&0111 != 0:
		key = "ex"
	default:
		key = "fi"
	}

	if val, ok := im[key]; ok {
		return val
	}

	return " "
}
