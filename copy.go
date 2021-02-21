package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func copySize(srcs []string) (int64, error) {
	var total int64

	for _, src := range srcs {
		_, err := os.Lstat(src)
		if os.IsNotExist(err) {
			return total, fmt.Errorf("src does not exist: %q", src)
		}

		err = filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
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

func copyFile(src, dst string, info os.FileInfo, nums chan int64) error {
	buf := make([]byte, 4096)

	r, err := os.Open(src)
	if err != nil {
		return err
	}
	defer r.Close()

	w, err := os.Create(dst)
	if err != nil {
		return err
	}

	for {
		n, err := r.Read(buf)
		if err != nil && err != io.EOF {
			w.Close()
			os.Remove(dst)
			return err
		}

		if n == 0 {
			break
		}

		if _, err := w.Write(buf[:n]); err != nil {
			return err
		}

		nums <- int64(n)
	}

	if err := w.Close(); err != nil {
		os.Remove(dst)
		return err
	}

	if err := os.Chmod(dst, info.Mode()); err != nil {
		os.Remove(dst)
		return err
	}

	return nil
}

func copyAll(srcs []string, dstDir string) (nums chan int64, errs chan error) {
	nums = make(chan int64, 1024)
	errs = make(chan error, 1024)

	go func() {
		for _, src := range srcs {
			dst := filepath.Join(dstDir, filepath.Base(src))

			_, err := os.Lstat(dst)
			if !os.IsNotExist(err) {
				var newPath string
				for i := 1; !os.IsNotExist(err); i++ {
					newPath = fmt.Sprintf("%s.~%d~", dst, i)
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
				if info.IsDir() {
					if err := os.MkdirAll(newPath, info.Mode()); err != nil {
						errs <- fmt.Errorf("mkdir: %s", err)
					}
					nums <- info.Size()
				} else if info.Mode()&os.ModeSymlink != 0 { /* Symlink */
					if rlink, err := os.Readlink(path); err != nil {
						errs <- fmt.Errorf("symlink: %s", err)
					} else {
						if err := os.Symlink(rlink, newPath); err != nil {
							errs <- fmt.Errorf("symlink: %s", err)
						}
					}
					nums <- info.Size()
				} else {
					if err := copyFile(path, newPath, info, nums); err != nil {
						errs <- fmt.Errorf("copy: %s", err)
					}
				}
				return nil
			})
		}

		close(errs)
	}()

	return nums, errs
}
