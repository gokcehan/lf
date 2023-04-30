package main

import (
	"log"

	"golang.org/x/sys/windows"
)

func diskFree(wd string) string {
	var free uint64

	pathPtr, err := windows.UTF16PtrFromString(wd)
	if err != nil {
		log.Printf("diskfree: %s", err)
		return ""
	}
	err = windows.GetDiskFreeSpaceEx(pathPtr, &free, nil, nil) // cwd, free, total, available
	if err != nil {
		log.Printf("diskfree: %s", err)
		return ""
	}
	return "df: " + humanize(int64(free))
}
