package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestCopySize(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b"), []byte("world!"), 0o644); err != nil {
		t.Fatal(err)
	}

	size, err := copySize([]string{filepath.Join(dir, "a"), filepath.Join(dir, "b")})
	if err != nil {
		t.Fatal(err)
	}
	if size != 11 {
		t.Errorf("expected 11 but got %d", size)
	}
}

func TestCopySizeNonexistent(t *testing.T) {
	if _, err := copySize([]string{"/nonexistent/path"}); err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestCopyFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	content := []byte("test content")

	if err := os.WriteFile(src, content, 0o644); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(src)
	if err != nil {
		t.Fatal(err)
	}

	nums := make(chan int64, 100)
	errs := make(chan error, 10)
	copyFile(src, dst, nil, info, nums, errs)

	select {
	case err := <-errs:
		t.Fatalf("copy error: %s", err)
	default:
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("expected %q but got %q", content, got)
	}
}

// drainCopyAll consumes the nums channel concurrently while collecting errors
// from errs. copyAll closes errs but not nums; the test owns nums and stops
// draining once errs closes.
func drainCopyAll(t *testing.T, nums chan int64, errs chan error) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		for range nums {
		}
		close(done)
	}()
	for err := range errs {
		t.Errorf("copy error: %s", err)
	}
	// errs closed; copyAll goroutine is done writing to nums.
	close(nums)
	<-done
}

func TestCopyAll(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(srcDir, "a.txt"), []byte("aaa"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "b.txt"), []byte("bbb"), 0o644); err != nil {
		t.Fatal(err)
	}

	nums, errs := copyAll(
		[]string{filepath.Join(srcDir, "a.txt"), filepath.Join(srcDir, "b.txt")},
		dstDir, nil,
	)
	drainCopyAll(t, nums, errs)

	for _, name := range []string{"a.txt", "b.txt"} {
		if _, err := os.Stat(filepath.Join(dstDir, name)); err != nil {
			t.Errorf("file %s not copied: %s", name, err)
		}
	}
}

func TestCopyAllDuplicate(t *testing.T) {
	gOpts.dupfilefmt = "%f.~%n~"

	srcDir := t.TempDir()
	dstDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(srcDir, "f.txt"), []byte("src"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dstDir, "f.txt"), []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}

	nums, errs := copyAll([]string{filepath.Join(srcDir, "f.txt")}, dstDir, nil)
	drainCopyAll(t, nums, errs)

	orig, err := os.ReadFile(filepath.Join(dstDir, "f.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(orig) != "existing" {
		t.Errorf("original overwritten: %q", orig)
	}

	entries, err := os.ReadDir(dstDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) < 2 {
		t.Errorf("expected duplicate, got %d files", len(entries))
	}
}
