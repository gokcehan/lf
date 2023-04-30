//go:build darwin || dragonfly || freebsd || linux

package main

import (
	"log"

	"golang.org/x/sys/unix"
)

func diskFree(wd string) string {
	var stat unix.Statfs_t

	if err := unix.Statfs(wd, &stat); err != nil {
		log.Printf("diskfree: %s", err)
		return ""
	}

	// Available blocks * size per block = available space in bytes
	return "df: " + humanize(int64(uint64(stat.Bavail)*uint64(stat.Bsize)))
}
