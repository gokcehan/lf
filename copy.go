package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/djherbis/times"
)

type ProgressWriter struct {
	writer io.Writer
	nums   chan<- int64
}

func NewProgressWriter(writer io.Writer, nums chan<- int64) *ProgressWriter {
	return &ProgressWriter{
		writer: writer,
		nums:   nums,
	}
}

func (progressWriter *ProgressWriter) Write(b []byte) (int, error) {
	n, err := progressWriter.writer.Write(b)
	progressWriter.nums <- int64(n)
	return n, err
}

func copySize(srcs []string) (int64, error) {
	var total int64

	for _, src := range srcs {
		_, err := os.Lstat(src)
		if os.IsNotExist(err) {
			return total, fmt.Errorf("src does not exist: %q", src)
		}

		err = filepath.Walk(src, func(_ string, info os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("walk: %w", err)
			}
			total += info.Size()
			return nil
		})
		if err != nil {
			return total, err
		}
	}

	return total, nil
}

// fsDir is a directory that entries are read from or created in
type fsDir interface {
	io.Closer
	Name() string
	Lstat(name string) (os.FileInfo, error)
	OpenFile(name string, flag int, perm os.FileMode) (*os.File, error)
	OpenRoot(name string) (*os.Root, error)
	Mkdir(name string, perm os.FileMode) error
	Symlink(oldname, newname string) error
	Readlink(name string) (string, error)
	Remove(name string) error
	Chtimes(name string, atime, mtime time.Time) error
}

// pathDir names a directory by path instead of holding it open
type pathDir string

func (d pathDir) Name() string { return string(d) }

func (d pathDir) Close() error { return nil }

func (d pathDir) Lstat(name string) (os.FileInfo, error) {
	return os.Lstat(filepath.Join(string(d), name))
}

func (d pathDir) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(filepath.Join(string(d), name), flag, perm)
}

func (d pathDir) OpenRoot(name string) (*os.Root, error) {
	return os.OpenRoot(filepath.Join(string(d), name))
}

func (d pathDir) Mkdir(name string, perm os.FileMode) error {
	return os.Mkdir(filepath.Join(string(d), name), perm)
}

func (d pathDir) Symlink(oldname, newname string) error {
	return os.Symlink(oldname, filepath.Join(string(d), newname))
}

func (d pathDir) Readlink(name string) (string, error) {
	return os.Readlink(filepath.Join(string(d), name))
}

func (d pathDir) Remove(name string) error {
	return os.Remove(filepath.Join(string(d), name))
}

func (d pathDir) Chtimes(name string, atime, mtime time.Time) error {
	return os.Chtimes(filepath.Join(string(d), name), atime, mtime)
}

// openDir holds path open when it can be read, else names it by path
func openDir(path string) fsDir {
	if root, err := os.OpenRoot(path); err == nil {
		return heldDir{root}
	}
	return pathDir(path)
}

// heldDir is a directory held open whose errors carry the full path
type heldDir struct{ *os.Root }

// full adds the parent path to an error that names only the entry
func (d heldDir) full(err error) error {
	if pathErr, ok := errors.AsType[*os.PathError](err); ok {
		pathErr.Path = filepath.Join(d.Name(), pathErr.Path)
	} else if linkErr, ok := errors.AsType[*os.LinkError](err); ok {
		linkErr.New = filepath.Join(d.Name(), linkErr.New)
	}
	return err
}

func (d heldDir) Lstat(name string) (os.FileInfo, error) {
	info, err := d.Root.Lstat(name)
	return info, d.full(err)
}

func (d heldDir) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	f, err := d.Root.OpenFile(name, flag, perm)
	return f, d.full(err)
}

func (d heldDir) OpenRoot(name string) (*os.Root, error) {
	root, err := d.Root.OpenRoot(name)
	return root, d.full(err)
}

func (d heldDir) Mkdir(name string, perm os.FileMode) error {
	return d.full(d.Root.Mkdir(name, perm))
}

func (d heldDir) Symlink(oldname, newname string) error {
	return d.full(d.Root.Symlink(oldname, newname))
}

func (d heldDir) Readlink(name string) (string, error) {
	target, err := d.Root.Readlink(name)
	return target, d.full(err)
}

func (d heldDir) Remove(name string) error {
	return d.full(d.Root.Remove(name))
}

func (d heldDir) Chtimes(name string, atime, mtime time.Time) error {
	return d.full(d.Root.Chtimes(name, atime, mtime))
}

// checkSame fails when the opened file is not the entry that was scanned
func checkSame(name string, info, stat os.FileInfo) error {
	if !os.SameFile(info, stat) {
		return fmt.Errorf("%s was replaced while being copied", name)
	}
	return nil
}

// openDirAt opens name in parent and checks it against the scanned entry
func openDirAt(parent fsDir, name string, info os.FileInfo) (heldDir, error) {
	root, err := parent.OpenRoot(name)
	if err != nil {
		return heldDir{}, err
	}
	dir := heldDir{root}
	stat, err := root.Stat(".")
	if err == nil {
		err = checkSame(dir.Name(), info, stat)
	}
	if err != nil {
		root.Close()
		return heldDir{}, dir.full(err)
	}
	return dir, nil
}

func copyFile(src fsDir, name string, dst fsDir, dstName string, preserve []string, info os.FileInfo, nums chan<- int64, errs chan<- error) {
	r, err := src.OpenFile(name, os.O_RDONLY, 0)
	if err != nil {
		errs <- err
		return
	}
	defer r.Close()

	// check the open file is the scanned file, not one swapped in by name
	stat, err := r.Stat()
	if err == nil {
		err = checkSame(r.Name(), info, stat)
	}
	if err != nil {
		errs <- err
		return
	}

	// keep only the permission bits, never setuid or setgid
	var dstMode os.FileMode = 0o666
	if slices.Contains(preserve, "mode") {
		dstMode = info.Mode().Perm()
	}
	// O_EXCL never writes through an existing file or a symlink at the name
	w, err := dst.OpenFile(dstName, os.O_RDWR|os.O_CREATE|os.O_EXCL, dstMode)
	if err != nil {
		errs <- err
		return
	}

	if _, err := io.Copy(NewProgressWriter(w, nums), r); err != nil {
		errs <- err
		w.Close()
		if err = dst.Remove(dstName); err != nil {
			errs <- err
		}
		return
	}

	if err := w.Close(); err != nil {
		errs <- err
		if err = dst.Remove(dstName); err != nil {
			errs <- err
		}
		return
	}

	if slices.Contains(preserve, "timestamps") {
		atime := times.Get(info).AccessTime()
		mtime := info.ModTime()
		if err := dst.Chtimes(dstName, atime, mtime); err != nil {
			errs <- err
		}
	}
}

func copyDir(src fsDir, name string, dst fsDir, dstName string, preserve []string, info os.FileInfo, nums chan<- int64, errs chan<- error) {
	// list the source before creating anything, a self copy must not see the copy
	f, err := src.OpenFile(name, os.O_RDONLY, 0)
	if err != nil {
		errs <- fmt.Errorf("walk: %w", err)
		return
	}
	stat, err := f.Stat()
	if err == nil {
		err = checkSame(f.Name(), info, stat)
	}
	if err != nil {
		f.Close()
		errs <- err
		return
	}
	entries, err := f.ReadDir(-1)
	f.Close()
	if err != nil {
		errs <- fmt.Errorf("walk: %w", err)
		return
	}

	dstMode := os.ModePerm
	if slices.Contains(preserve, "mode") {
		dstMode = info.Mode().Perm()
	}
	// create with owner access so the new directory can be opened and written
	// Mkdir fails if the name exists, refusing a planted directory or symlink
	if err := dst.Mkdir(dstName, dstMode|0o700); err != nil {
		errs <- fmt.Errorf("mkdir: %w", err)
		return
	}
	dstInfo, err := dst.Lstat(dstName)
	if err != nil {
		errs <- err
		return
	}
	dstDir, err := openDirAt(dst, dstName, dstInfo)
	if err != nil {
		errs <- err
		return
	}
	defer dstDir.Close()

	// hold the source directory open while copying its entries
	if len(entries) > 0 {
		srcDir, err := openDirAt(src, name, info)
		if err != nil {
			errs <- fmt.Errorf("walk: %w", err)
			return
		}
		for _, entry := range entries {
			copyEntry(srcDir, entry.Name(), dstDir, entry.Name(), preserve, nums, errs)
		}
		srcDir.Close()
	}

	// set the directory time last, after its entries are written
	if slices.Contains(preserve, "timestamps") {
		atime := times.Get(info).AccessTime()
		mtime := info.ModTime()
		if err := dst.Chtimes(dstName, atime, mtime); err != nil {
			errs <- err
		}
	}

	// restore the real mode last
	if dstMode&0o700 != 0o700 {
		if err := dstDir.Chmod(".", dstMode); err != nil {
			errs <- dstDir.full(err)
		}
	}
}

// copyEntry copies the entry name in src to dstName in dst, a symlink as a symlink
func copyEntry(src fsDir, name string, dst fsDir, dstName string, preserve []string, nums chan<- int64, errs chan<- error) {
	info, err := src.Lstat(name)
	if err != nil {
		errs <- fmt.Errorf("walk: %w", err)
		return
	}

	switch {
	case info.IsDir():
		nums <- info.Size()
		copyDir(src, name, dst, dstName, preserve, info, nums, errs)
	case info.Mode()&os.ModeSymlink != 0:
		if target, err := src.Readlink(name); err != nil {
			errs <- fmt.Errorf("symlink: %w", err)
		} else if err := dst.Symlink(target, dstName); err != nil {
			errs <- fmt.Errorf("symlink: %w", err)
		}
		nums <- info.Size()
	case !info.Mode().IsRegular():
		errs <- fmt.Errorf("cannot copy irregular file %s (named pipe, socket or device)", filepath.Join(src.Name(), name))
		nums <- info.Size()
	default:
		copyFile(src, name, dst, dstName, preserve, info, nums, errs)
	}
}

func copyAll(srcs []string, dstDir string, preserve []string) (nums chan int64, errs chan error) {
	nums = make(chan int64, 1024)
	errs = make(chan error, 1024)

	go func() {
		defer close(errs)

		// hold the destination open when it can be read, else use it by path
		dst := openDir(dstDir)
		defer dst.Close()

		for _, src := range srcs {
			src = filepath.Clean(src)
			file := filepath.Base(src)

			if lstat, err := dst.Lstat(file); err == nil {
				ext := getFileExtension(lstat)
				basename := file[:len(file)-len(ext)]
				for i := 1; err == nil; i++ {
					file = strings.ReplaceAll(gOpts.dupfilefmt, "%f", basename+ext)
					file = strings.ReplaceAll(file, "%b", basename)
					file = strings.ReplaceAll(file, "%e", ext)
					file = strings.ReplaceAll(file, "%n", strconv.Itoa(i))
					_, err = dst.Lstat(file)
				}
				if !os.IsNotExist(err) {
					errs <- err
					continue
				}
			}

			if rel, err := filepath.Rel(src, filepath.Join(dstDir, file)); err == nil && rel != "." && filepath.IsLocal(rel) {
				errs <- fmt.Errorf("cannot copy %s into a subdirectory of itself", src)
				continue
			}

			parent := openDir(filepath.Dir(src))
			copyEntry(parent, filepath.Base(src), dst, file, preserve, nums, errs)
			parent.Close()
		}
	}()

	return nums, errs
}
