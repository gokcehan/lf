//go:build illumos || netbsd || solaris

package main

import (
	"log"

	"golang.org/x/sys/unix"
)

func diskFree(wd string) string {
	var stat unix.Statvfs_t

	if err := unix.Statvfs(wd, &stat); err != nil {
		log.Printf("diskfree: %s", err)
		return ""
	}

	// Available blocks * size per block = available space in bytes
	return "df: " + humanize(uint64(stat.Bavail)*uint64(stat.Bsize))
}
