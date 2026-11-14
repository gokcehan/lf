package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDirSize(t *testing.T) {
	tmp := t.TempDir()

	dir := filepath.Join(tmp, "dir")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "file"), make([]byte, 4096), 0o644); err != nil {
		t.Fatal(err)
	}

	link := filepath.Join(tmp, "link")
	if err := os.Symlink(dir, link); err != nil {
		t.Skipf("creating symlink: %s", err)
	}

	expected, err := calcSize(newFile(dir))
	if err != nil {
		t.Fatal(err)
	}

	got, err := calcSize(newFile(link))
	if err != nil {
		t.Fatal(err)
	}

	if got != expected {
		t.Errorf("at symlink to directory expected %d but got %d", expected, got)
	}
}

// a newline name must not reach shell commands through $f, $fx or $fv
func TestExportFilesNewline(t *testing.T) {
	t.Setenv("PWD", "") // restored after the test, exportFiles changes it

	// the cursor and the visual selection are on the newline name
	d := &dir{path: "/tmp/dir", files: []*file{{path: "/tmp/dir/a\n"}}}
	nav := &nav{dirPaths: []string{d.path}, dirCache: map[string]*dir{d.path: d}}

	if err := nav.exportFiles(); err == nil {
		t.Error("no error for the dropped name")
	}
	for _, name := range []string{"f", "fx", "fv"} {
		if v := os.Getenv(name); v != "" {
			t.Errorf("$%s = %q, want empty", name, v)
		}
	}

	// a selection of only newline paths must not fall back to the cursor file
	d.files = []*file{{path: "/tmp/dir/c"}}
	nav.selections = map[string]int{"/tmp/dir/a\n": 0}
	if err := nav.exportFiles(); err == nil {
		t.Error("no error for the dropped name")
	}
	if v := os.Getenv("fx"); v != "" {
		t.Errorf("$fx = %q, want empty", v)
	}
}
