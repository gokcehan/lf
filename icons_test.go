package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLinkTargetIconOverrides(t *testing.T) {
	link := &file{FileInfo: &fakeStat{name: "link"}, linkState: working}
	for _, test := range []struct {
		name, config, want string
	}{
		{"plain-icon", "ln L\n", "L"},
		{"styled-icon", "ln L 31\n", "L"},
		{"target-icon", "ln target\n", "F"},
	} {
		t.Run(test.name, func(t *testing.T) {
			im := iconMap{icons: make(map[string]iconDef)}
			im.parseEnv("ln=target:fi=F")
			path := filepath.Join(t.TempDir(), "icons")
			if err := os.WriteFile(path, []byte(test.config), 0o600); err != nil {
				t.Fatal(err)
			}
			im.parseFile(path)
			if got := im.get(link).icon; got != test.want {
				t.Errorf("link icon = %q, want %q", got, test.want)
			}
		})
	}
}
