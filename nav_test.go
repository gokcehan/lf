package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
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

func TestDisableDirCountsKeepsCalculatedSize(t *testing.T) {
	for _, test := range []struct {
		name  string
		local bool
	}{
		{"global", false},
		{"local", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			opts := gOpts
			dircounts := gLocalOpts.dircounts
			t.Cleanup(func() {
				gOpts = opts
				gLocalOpts.dircounts = dircounts
			})
			gOpts.dircounts = true
			gOpts.preview = false
			gOpts.info = []string{"size"}
			gLocalOpts.dircounts = make(map[string]bool)

			path := t.TempDir()
			child := filepath.Join(path, "dir")
			if err := os.Mkdir(child, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(child, "file"), make([]byte, 4096), 0o644); err != nil {
				t.Fatal(err)
			}
			d := newDir(path)
			d.sort()
			nav := &nav{
				dirPaths: []string{path},
				dirCache: map[string]*dir{path: d, child: newDir(child)},
				dirChan:  make(chan *dir, 1),
				height:   10,
			}
			app := &app{nav: nav, ui: &ui{currentFile: child}}
			if test.local {
				(&setLocalExpr{path, "dircounts", "false"}).eval(app, nil)
				nav.renew()
			} else {
				(&setExpr{"dircounts", "false"}).eval(app, nil)
			}
			(&callExpr{"calcdirsize", nil, 1}).eval(app, nil)
			if app.ui.msg != "" {
				t.Fatal(app.ui.msg)
			}
			want := nav.currFile().dirSize
			if want < 4096 {
				t.Fatalf("directory size was not calculated: %d", want)
			}

			if d.loading {
				select {
				case d = <-nav.dirChan:
					d.sort()
					nav.dirCache[path] = d
				case <-time.After(5 * time.Second):
					t.Fatal("directory refresh did not finish")
				}
			}
			if got := nav.currFile().dirSize; got != want {
				t.Errorf("directory size after refresh: got %d, want %d", got, want)
			}
			info, _, _ := fileInfo(nav.currFile(), d, 0, 0, 0)
			if want := fmt.Sprintf(" %5s", humanize(want)); info != want {
				t.Errorf("size column: got %q, want %q", info, want)
			}
		})
	}
}

func TestDirCountsReload(t *testing.T) {
	for _, test := range []struct {
		name    string
		changed bool
	}{
		{"enable_counts", false},
		{"filesystem_change", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			opts := gOpts
			t.Cleanup(func() { gOpts = opts })
			gOpts.dircounts = test.changed

			path := t.TempDir()
			child := filepath.Join(path, "dir")
			if err := os.Mkdir(child, 0o755); err != nil {
				t.Fatal(err)
			}
			d := newDir(path)
			d.sort()
			if err := os.WriteFile(filepath.Join(child, "file"), nil, 0o644); err != nil {
				t.Fatal(err)
			}
			gOpts.dircounts = !test.changed
			if test.changed {
				if err := os.WriteFile(filepath.Join(path, "new-file"), nil, 0o644); err != nil {
					t.Fatal(err)
				}
				d.loadTime = time.Time{}
			}
			nav := &nav{dirChan: make(chan *dir, 1)}
			nav.checkDir(d)
			if !d.loading {
				t.Fatal("directory was not reloaded")
			}
			select {
			case d = <-nav.dirChan:
			case <-time.After(5 * time.Second):
				t.Fatal("directory refresh did not finish")
			}
			d.sort()
			if test.changed {
				if len(d.files) != 2 {
					t.Errorf("file count: got %d, want 2", len(d.files))
				}
			} else if got := d.files[0].dirCount; got != 1 {
				t.Errorf("directory count: got %d, want 1", got)
			}
		})
	}
}
