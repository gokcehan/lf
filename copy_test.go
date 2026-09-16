package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCopyAllDupName(t *testing.T) {
	old := gOpts.dupfilefmt
	defer func() { gOpts.dupfilefmt = old }()

	tests := []struct {
		name    string
		format  string
		file    string
		wantErr bool
	}{
		{"duplicate with default format", "%f.~%n~", "file.txt", false},
		{"name too long", "%f.~%n~", strings.Repeat("a", 255), true},
		{"format without %n", "%f", "file.txt", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gOpts.dupfilefmt = test.format
			dir := t.TempDir()
			src := filepath.Join(dir, test.file)
			if err := os.WriteFile(src, nil, 0o600); err != nil {
				t.Skipf("creating %d-byte filename: %s", len(test.file), err)
			}

			// copying a file into its own directory forces a duplicate name
			nums, errs := copyAll([]string{src}, dir, nil)
			var gotErr bool
			deadline := time.After(5 * time.Second)
			for {
				select {
				case <-nums:
				case err, ok := <-errs:
					if !ok {
						if gotErr != test.wantErr {
							t.Errorf("expected error: %v, got error: %v", test.wantErr, gotErr)
						}
						return
					}
					t.Log(err)
					gotErr = true
				case <-deadline:
					t.Fatal("copyAll did not finish, duplicate name loop is stuck")
				}
			}
		})
	}
}
