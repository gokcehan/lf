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

func copyFile(src string, root *os.Root, dstRel string, preserve []string, info os.FileInfo, nums chan<- int64, errs chan<- error) {
	r, err := os.Open(src)
	if err != nil {
		errs <- err
		return
	}
	defer r.Close()

	var dstMode os.FileMode = 0o666
	if slices.Contains(preserve, "mode") {
		dstMode = info.Mode().Perm()
	}
	w, err := root.OpenFile(dstRel, os.O_RDWR|os.O_CREATE|os.O_TRUNC, dstMode)
	if err != nil {
		errs <- err
		return
	}

	if _, err := io.Copy(NewProgressWriter(w, nums), r); err != nil {
		errs <- err
		w.Close()
		if err = root.Remove(dstRel); err != nil {
			errs <- err
		}
		return
	}

	if err := w.Close(); err != nil {
		errs <- err
		if err = root.Remove(dstRel); err != nil {
			errs <- err
		}
		return
	}

	if slices.Contains(preserve, "timestamps") {
		atime := times.Get(info).AccessTime()
		mtime := info.ModTime()
		if err := root.Chtimes(dstRel, atime, mtime); err != nil {
			errs <- err
		}
	}
}

func copyAll(srcs []string, dstDir string, preserve []string) (nums chan int64, errs chan error) {
	nums = make(chan int64, 1024)
	errs = make(chan error, 1024)

	go func() {
		root, err := os.OpenRoot(dstDir)
		if err != nil {
			errs <- fmt.Errorf("open destination: %w", err)
			close(errs)
			return
		}
		defer root.Close()

		dirInfos := make(map[string]os.FileInfo)

		for _, src := range srcs {
			file := filepath.Base(src)
			dstRel := file

			if lstat, err := root.Lstat(dstRel); err == nil {
				ext := getFileExtension(lstat)
				basename := file[:len(file)-len(ext)]
				var newName string
				for i := 1; ; i++ {
					newName = strings.ReplaceAll(gOpts.dupfilefmt, "%f", basename+ext)
					newName = strings.ReplaceAll(newName, "%b", basename)
					newName = strings.ReplaceAll(newName, "%e", ext)
					newName = strings.ReplaceAll(newName, "%n", strconv.Itoa(i))
					if _, err := root.Lstat(newName); err != nil {
						break
					}
				}
				dstRel = newName
			}

			err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					errs <- fmt.Errorf("walk: %w", err)
					return nil
				}
				rel, err := filepath.Rel(src, path)
				if err != nil {
					errs <- fmt.Errorf("relative: %w", err)
					return nil
				}
				newRel := filepath.Join(dstRel, rel)
				switch {
				case info.IsDir():
					dstMode := os.ModePerm
					if slices.Contains(preserve, "mode") {
						dstMode = info.Mode().Perm()
					}
					if err := root.MkdirAll(newRel, dstMode); err != nil {
						errs <- fmt.Errorf("mkdir: %w", err)
					}
					if slices.Contains(preserve, "timestamps") {
						dirInfos[newRel] = info
					}
					nums <- info.Size()
				case info.Mode()&os.ModeSymlink != 0:
					if rlink, err := os.Readlink(path); err != nil {
						errs <- fmt.Errorf("symlink: %w", err)
					} else {
						if err := root.Symlink(rlink, newRel); err != nil {
							errs <- fmt.Errorf("symlink: %w", err)
						}
					}
					nums <- info.Size()
				default:
					copyFile(path, root, newRel, preserve, info, nums, errs)
				}
				return nil
			})
			if err != nil {
				errs <- fmt.Errorf("walk: %w", err)
			}
		}

		for rel, info := range dirInfos {
			atime := times.Get(info).AccessTime()
			mtime := info.ModTime()
			if err := root.Chtimes(rel, atime, mtime); err != nil {
				errs <- err
			}
		}

		close(errs)
	}()

	return nums, errs
}
