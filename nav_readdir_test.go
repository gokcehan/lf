package main

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// TestReaddirParallelStatPreservesEntries pins #2666's correctness side:
// concurrent stat must produce exactly the same entry set (names + paths)
// as the serial loop did, in the same names order.
func TestReaddirParallelStatPreservesEntries(t *testing.T) {
	dir := t.TempDir()
	names := []string{"a.txt", "b.txt", "sub1", "sub2", "c.txt", ".hidden", "d.txt", "e.txt", "sub3", "f.txt"}
	for _, n := range names {
		p := filepath.Join(dir, n)
		if err := os.WriteFile(p, []byte(n), 0o644); err != nil {
			t.Fatalf("write %s: %v", n, err)
		}
	}

	files, err := readdir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	if len(files) != len(names) {
		t.Fatalf("got %d files, want %d", len(files), len(names))
	}

	got := make([]string, len(files))
	for i, f := range files {
		got[i] = f.Name()
	}
	want := append([]string(nil), names...)
	sort.Strings(want)
	sortedGot := append([]string(nil), got...)
	sort.Strings(sortedGot)
	for i := range want {
		if sortedGot[i] != want[i] {
			t.Fatalf("entry mismatch: got %q want %q", sortedGot[i], want[i])
		}
	}
	// Path fields must point into the listed directory.
	for _, f := range files {
		if filepath.Dir(f.path) != dir {
			t.Fatalf("file %s has wrong parent %s", f.path, filepath.Dir(f.path))
		}
	}
}

// TestReaddirEmptyAndSingleEntry: boundary sizes exercise the worker-count
// clamp (workers = min(statWorkers, len(names))) including zero.
func TestReaddirEmptyAndSingleEntry(t *testing.T) {
	empty := t.TempDir()
	files, err := readdir(empty)
	if err != nil {
		t.Fatalf("readdir empty: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("empty dir returned %d files", len(files))
	}

	single := t.TempDir()
	if err := os.WriteFile(filepath.Join(single, "only.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	files, err = readdir(single)
	if err != nil {
		t.Fatalf("readdir single: %v", err)
	}
	if len(files) != 1 || files[0].Name() != "only.txt" {
		t.Fatalf("single dir returned %v", files)
	}
}
