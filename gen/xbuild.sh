#!/bin/sh
# Generates cross builds for all supported platforms.
#
# This script is used to build binaries for all supported platforms. Cgo is
# disabled to make sure binaries are statically linked. Appropriate flags are
# given to the go compiler to strip binaries. These are then compressed in an
# archive form (`.zip` for windows and `.tar.gz` for the rest) within a folder
# named `dist`.

set -o verbose

mkdir -p dist

CGO_ENABLED=0 GOOS=darwin  GOARCH=386   go build -ldflags='-s -w' && sync && tar czf dist/lf-darwin-386.tar.gz    lf --remove-files
CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -ldflags='-s -w' && sync && tar czf dist/lf-darwin-amd64.tar.gz  lf --remove-files
CGO_ENABLED=0 GOOS=freebsd GOARCH=386   go build -ldflags='-s -w' && sync && tar czf dist/lf-freebsd-386.tar.gz   lf --remove-files
CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build -ldflags='-s -w' && sync && tar czf dist/lf-freebsd-amd64.tar.gz lf --remove-files
CGO_ENABLED=0 GOOS=freebsd GOARCH=arm   go build -ldflags='-s -w' && sync && tar czf dist/lf-freebsd-arm.tar.gz   lf --remove-files
CGO_ENABLED=0 GOOS=linux   GOARCH=386   go build -ldflags='-s -w' && sync && tar czf dist/lf-linux-386.tar.gz     lf --remove-files
CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -ldflags='-s -w' && sync && tar czf dist/lf-linux-amd64.tar.gz   lf --remove-files
CGO_ENABLED=0 GOOS=linux   GOARCH=arm   go build -ldflags='-s -w' && sync && tar czf dist/lf-linux-arm.tar.gz     lf --remove-files
CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build -ldflags='-s -w' && sync && tar czf dist/lf-linux-arm64.tar.gz   lf --remove-files
CGO_ENABLED=0 GOOS=netbsd  GOARCH=386   go build -ldflags='-s -w' && sync && tar czf dist/lf-netbsd-386.tar.gz    lf --remove-files
CGO_ENABLED=0 GOOS=netbsd  GOARCH=amd64 go build -ldflags='-s -w' && sync && tar czf dist/lf-netbsd-amd64.tar.gz  lf --remove-files
CGO_ENABLED=0 GOOS=netbsd  GOARCH=arm   go build -ldflags='-s -w' && sync && tar czf dist/lf-netbsd-arm.tar.gz    lf --remove-files
CGO_ENABLED=0 GOOS=openbsd GOARCH=386   go build -ldflags='-s -w' && sync && tar czf dist/lf-openbsd-386.tar.gz   lf --remove-files
CGO_ENABLED=0 GOOS=openbsd GOARCH=amd64 go build -ldflags='-s -w' && sync && tar czf dist/lf-openbsd-amd64.tar.gz lf --remove-files
CGO_ENABLED=0 GOOS=openbsd GOARCH=arm   go build -ldflags='-s -w' && sync && tar czf dist/lf-openbsd-arm.tar.gz   lf --remove-files

CGO_ENABLED=0 GOOS=windows GOARCH=386   go build -ldflags='-s -w' && sync && zip dist/lf-windows-386.zip   lf.exe --move
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags='-s -w' && sync && zip dist/lf-windows-amd64.zip lf.exe --move
