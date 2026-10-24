package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

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

// checkSame fails when the opened file is not the entry that was scanned
func checkSame(name string, info, stat os.FileInfo) error {
	if !os.SameFile(info, stat) {
		return fmt.Errorf("%s was replaced while being copied", name)
	}
	return nil
}

// fullPathErr adds the root directory to the relative path in an os.Root error
func fullPathErr(root *os.Root, err error) error {
	switch e := err.(type) {
	case *os.PathError:
		e.Path = filepath.Join(root.Name(), e.Path)
	case *os.LinkError:
		e.New = filepath.Join(root.Name(), e.New)
	}
	return err
}

// openDirAt opens name in parent and checks it against the scanned entry
func openDirAt(parent *os.Root, name string, info os.FileInfo) (*os.Root, error) {
	root, err := parent.OpenRoot(name)
	if err != nil {
		return nil, fullPathErr(parent, err)
	}
	stat, err := root.Stat(".")
	if err == nil {
		err = checkSame(root.Name(), info, stat)
	}
	if err != nil {
		root.Close()
		return nil, fullPathErr(root, err)
	}
	return root, nil
}

func copyFile(src *os.Root, name string, dst *os.Root, dstName string, preserve []string, info os.FileInfo, nums chan<- int64, errs chan<- error) {
	r, err := src.OpenFile(name, os.O_RDONLY, 0)
	if err != nil {
		errs <- fullPathErr(src, err)
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
		errs <- fullPathErr(dst, err)
		return
	}

	if _, err := io.Copy(NewProgressWriter(w, nums), r); err != nil {
		errs <- err
		w.Close()
		if err = dst.Remove(dstName); err != nil {
			errs <- fullPathErr(dst, err)
		}
		return
	}

	// set the exact mode, a refusal from a file system like FAT with a fixed owner does not fail the copy
	if slices.Contains(preserve, "mode") {
		if err := w.Chmod(dstMode); err != nil && !os.IsPermission(err) {
			errs <- err
		}
	}

	if err := w.Close(); err != nil {
		errs <- err
		if err = dst.Remove(dstName); err != nil {
			errs <- fullPathErr(dst, err)
		}
		return
	}

	if slices.Contains(preserve, "timestamps") {
		atime := times.Get(info).AccessTime()
		mtime := info.ModTime()
		if err := dst.Chtimes(dstName, atime, mtime); err != nil && !os.IsPermission(err) {
			errs <- fullPathErr(dst, err)
		}
	}
}

func copyDir(src *os.Root, name string, dst *os.Root, dstName string, preserve []string, info os.FileInfo, nums chan<- int64, errs chan<- error) {
	// list the source before creating anything, a self copy must not see the copy
	f, err := src.OpenFile(name, os.O_RDONLY, 0)
	if err != nil {
		errs <- fmt.Errorf("walk: %w", fullPathErr(src, err))
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
	mkMode := dstMode | 0o700
	if slices.Contains(preserve, "mode") {
		dstMode = info.Mode().Perm()
		// keep group and others out while the entries are written, the final mode is set below
		mkMode = dstMode&^0o022 | 0o700
	}
	// Mkdir fails when the name exists, refusing a planted directory or symlink
	if err := dst.Mkdir(dstName, mkMode); err != nil {
		errs <- fmt.Errorf("mkdir: %w", fullPathErr(dst, err))
		return
	}
	dstInfo, err := dst.Lstat(dstName)
	if err != nil {
		errs <- fullPathErr(dst, err)
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
		} else {
			for _, entry := range entries {
				copyEntry(srcDir, entry.Name(), dstDir, entry.Name(), preserve, nums, errs)
			}
			srcDir.Close()
		}
	}

	// set the directory time last, a refusal from a file system like FAT does not fail the copy
	if slices.Contains(preserve, "timestamps") {
		atime := times.Get(info).AccessTime()
		mtime := info.ModTime()
		if err := dst.Chtimes(dstName, atime, mtime); err != nil && !os.IsPermission(err) {
			errs <- fullPathErr(dst, err)
		}
	}

	// set the exact mode last, the umask does not apply here
	if slices.Contains(preserve, "mode") {
		if err := dstDir.Chmod(".", dstMode); err != nil && !os.IsPermission(err) {
			errs <- fullPathErr(dstDir, err)
		}
	}
}

// copyEntry copies the entry name in src to dstName in dst, a symlink as a symlink
func copyEntry(src *os.Root, name string, dst *os.Root, dstName string, preserve []string, nums chan<- int64, errs chan<- error) {
	info, err := src.Lstat(name)
	if err != nil {
		errs <- fmt.Errorf("walk: %w", fullPathErr(src, err))
		return
	}

	switch {
	case info.IsDir():
		nums <- info.Size()
		copyDir(src, name, dst, dstName, preserve, info, nums, errs)
	case info.Mode()&os.ModeSymlink != 0:
		if target, err := src.Readlink(name); err != nil {
			errs <- fmt.Errorf("symlink: %w", fullPathErr(src, err))
		} else if err := symlinkAt(dst, target, dstName); err != nil {
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

// realPath resolves the parent of path through the file system as written and keeps the last name
func realPath(path string) string {
	dir, name := filepath.Split(path)
	if dir, err := filepath.EvalSymlinks(dir + "."); err == nil {
		return filepath.Join(dir, name)
	}
	return path
}

// copyAll copies srcs into dstDir using srcParent as the open parent when it is set
func copyAll(srcs []string, dstDir string, preserve []string, srcParent *os.Root) (nums chan int64, errs chan error) {
	nums = make(chan int64, 1024)
	errs = make(chan error, 1024)

	go func() {
		defer close(errs)

		dst, err := os.OpenRoot(dstDir)
		if err != nil {
			errs <- err
			return
		}
		defer dst.Close()

		for _, src := range srcs {
			// the copy keeps the written name, like cp, a path ending in . or .. takes the name of the directory it names
			file := filepath.Base(src)
			src = realPath(src)
			if file == "." || file == ".." {
				file = filepath.Base(src)
			}

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
					errs <- fullPathErr(dst, err)
					continue
				}
			}

			if rel, err := filepath.Rel(src, realPath(filepath.Join(dstDir, file))); err == nil && rel != "." && filepath.IsLocal(rel) {
				errs <- fmt.Errorf("cannot copy %s into a subdirectory of itself", src)
				continue
			}

			parent := srcParent
			if parent == nil {
				if parent, err = os.OpenRoot(filepath.Dir(src)); err != nil {
					errs <- err
					continue
				}
			}
			copyEntry(parent, filepath.Base(src), dst, file, preserve, nums, errs)
			if srcParent == nil {
				parent.Close()
			}
		}
	}()

	return nums, errs
}
