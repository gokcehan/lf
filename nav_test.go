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

func TestCopyDirSizes(t *testing.T) {
	tmp := t.TempDir()

	for _, name := range []string{"a", "b"} {
		if err := os.Mkdir(filepath.Join(tmp, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	prev := newDir(tmp)
	for _, f := range prev.allFiles {
		if f.Name() == "a" {
			f.dirSize = 42
		}
	}

	if err := os.Mkdir(filepath.Join(tmp, "c"), 0o755); err != nil {
		t.Fatal(err)
	}

	curr := newDir(tmp)
	curr.copyDirSizes(prev)

	expected := map[string]int64{"a": 42, "b": -1, "c": -1}
	for _, f := range curr.allFiles {
		if got := f.dirSize; got != expected[f.Name()] {
			t.Errorf("at %q expected dirSize %d but got %d", f.Name(), expected[f.Name()], got)
		}
	}
}
