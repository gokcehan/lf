package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"testing"
)

// runCopy copies srcs into dstDir and returns the reported errors
func runCopy(srcs []string, dstDir string, preserve []string) []error {
	nums, errs := copyAll(srcs, dstDir, preserve, nil)
	go func() {
		for range nums {
		}
	}()

	var all []error
	for err := range errs {
		all = append(all, err)
	}
	return all
}

func writeFile(t *testing.T, path, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCopyAll(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	dst := filepath.Join(tmp, "dst")
	if err := os.Mkdir(dst, 0o755); err != nil {
		t.Fatal(err)
	}

	writeFile(t, filepath.Join(src, "file"), "data")
	writeFile(t, filepath.Join(src, "dir", "nested"), "nested")
	link := os.Symlink("file", filepath.Join(src, "link")) == nil

	if errs := runCopy([]string{src}, dst, nil); len(errs) != 0 {
		t.Fatalf("copy failed: %v", errs)
	}

	for name, want := range map[string]string{"file": "data", "dir/nested": "nested"} {
		got, err := os.ReadFile(filepath.Join(dst, "src", name))
		if err != nil || string(got) != want {
			t.Errorf("%s: got %q, %v, want %q", name, got, err, want)
		}
	}

	if link {
		target, err := os.Readlink(filepath.Join(dst, "src", "link"))
		if err != nil || target != "file" {
			t.Errorf("symlink not copied as a symlink: got %q, %v", target, err)
		}
	}
}

func TestCopyAllPreserve(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file modes are not supported on windows")
	}

	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	dst := filepath.Join(tmp, "dst")
	if err := os.Mkdir(dst, 0o755); err != nil {
		t.Fatal(err)
	}

	// the permission bits must be kept and the special bits must be dropped
	special := os.ModeSetuid | os.ModeSetgid | os.ModeSticky
	modes := map[string]os.FileMode{
		"file": 0o664 | os.ModeSetuid | os.ModeSetgid,
		"dir":  0o775 | os.ModeSetgid | os.ModeSticky,
	}
	writeFile(t, filepath.Join(src, "file"), "data")
	writeFile(t, filepath.Join(src, "dir", "nested"), "nested")
	for name, mode := range modes {
		if err := os.Chmod(filepath.Join(src, name), mode); err != nil {
			t.Skipf("setting mode: %s", err)
		}
	}

	if errs := runCopy([]string{src}, dst, []string{"mode"}); len(errs) != 0 {
		t.Fatalf("copy failed: %v", errs)
	}

	for name, mode := range modes {
		info, err := os.Stat(filepath.Join(dst, "src", name))
		if err != nil {
			t.Fatal(err)
		}
		if got, want := info.Mode()&(os.ModePerm|special), mode.Perm(); got != want {
			t.Errorf("%s: got mode %v, want %v", name, got, want)
		}
	}
}

// the copy must not leave the source or the destination when a directory there is swapped for a symlink
func TestCopyAllSymlinkRace(t *testing.T) {
	for _, side := range []string{"source", "destination"} {
		for range 5 {
			tmp := t.TempDir()
			src := filepath.Join(tmp, "src")
			dst := filepath.Join(tmp, "dst")
			outside := filepath.Join(tmp, "outside")
			for i := range 200 {
				writeFile(t, filepath.Join(src, "sub", fmt.Sprintf("file%d", i)), "data")
				writeFile(t, filepath.Join(outside, fmt.Sprintf("file%d", i)), "outside")
			}
			if err := os.Mkdir(dst, 0o755); err != nil {
				t.Fatal(err)
			}

			target := filepath.Join(src, "sub")
			if side == "destination" {
				target = filepath.Join(dst, "src", "sub")
			}

			var stop atomic.Bool
			done := make(chan struct{})
			go func() {
				defer close(done)
				for !stop.Load() {
					// wait until the copy has started
					if _, err := os.Lstat(filepath.Join(dst, "src")); err != nil {
						continue
					}
					if os.Rename(target, target+".moved") == nil {
						_ = os.Symlink(outside, target)
						return
					}
				}
			}()

			runCopy([]string{src}, dst, nil)
			stop.Store(true)
			<-done

			// nothing may be read from or written to the directory behind the symlink
			check := func(dir, want string) {
				err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
					if err != nil || !d.Type().IsRegular() {
						return err
					}
					if data, _ := os.ReadFile(path); string(data) != want {
						t.Fatalf("%s swap: %s contains %q, want %q", side, path, data, want)
					}
					return nil
				})
				if err != nil {
					t.Fatal(err)
				}
			}
			check(filepath.Join(dst, "src"), "data")
			check(outside, "outside")
		}
	}
}
