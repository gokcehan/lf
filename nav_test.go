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
