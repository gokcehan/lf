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
				return fmt.Errorf("walk: %s", err)
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

func copyFile(src, dst string, nums chan int64) error {
	r, err := os.Open(src)
	if err != nil {
		return err
	}
	defer r.Close()

	w, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o666)
	if err != nil {
		return err
	}

	if _, err := io.Copy(NewProgressWriter(w, nums), r); err != nil {
		w.Close()
		os.Remove(dst)
		return err
	}

	if err := w.Close(); err != nil {
		os.Remove(dst)
		return err
	}

	return nil
}

func copyAll(srcs []string, dstDir string, preserve []string) (nums chan int64, errs chan error) {
	nums = make(chan int64, 1024)
	errs = make(chan error, 1024)

	go func() {
		infos := make(map[string]os.FileInfo)

		for _, src := range srcs {
			file := filepath.Base(src)
			dst := filepath.Join(dstDir, file)

			lstat, err := os.Lstat(dst)
			if !os.IsNotExist(err) {
				ext := getFileExtension(lstat)
				basename := file[:len(file)-len(ext)]
				var newPath string
				for i := 1; !os.IsNotExist(err); i++ {
					file = strings.ReplaceAll(gOpts.dupfilefmt, "%f", basename+ext)
					file = strings.ReplaceAll(file, "%b", basename)
					file = strings.ReplaceAll(file, "%e", ext)
					file = strings.ReplaceAll(file, "%n", strconv.Itoa(i))
					newPath = filepath.Join(dstDir, file)
					_, err = os.Lstat(newPath)
				}
				dst = newPath
			}

			filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					errs <- fmt.Errorf("walk: %s", err)
					return nil
				}
				rel, err := filepath.Rel(src, path)
				if err != nil {
					errs <- fmt.Errorf("relative: %s", err)
					return nil
				}
				newPath := filepath.Join(dst, rel)
				switch {
				case info.IsDir():
					infos[newPath] = info
					if err := os.MkdirAll(newPath, os.ModePerm); err != nil {
						errs <- fmt.Errorf("mkdir: %s", err)
					}
					nums <- info.Size()
				case info.Mode()&os.ModeSymlink != 0:
					if rlink, err := os.Readlink(path); err != nil {
						errs <- fmt.Errorf("symlink: %s", err)
					} else if err := os.Symlink(rlink, newPath); err != nil {
						errs <- fmt.Errorf("symlink: %s", err)
					}
					nums <- info.Size()
				default:
					infos[newPath] = info
					if err := copyFile(path, newPath, nums); err != nil {
						errs <- fmt.Errorf("copy: %s", err)
					}
				}
				return nil
			})
		}

		for path, info := range infos {
			if slices.Contains(preserve, "mode") {
				if err := os.Chmod(path, info.Mode()); err != nil {
					errs <- fmt.Errorf("chmod: %s", err)
				}
			}
			if slices.Contains(preserve, "timestamps") {
				atime := times.Get(info).AccessTime()
				mtime := info.ModTime()
				if err := os.Chtimes(path, atime, mtime); err != nil {
					errs <- fmt.Errorf("chtimes: %s", err)
				}
			}
		}

		close(errs)
	}()

	return nums, errs
}
